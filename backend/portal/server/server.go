package server

import (
	_ "embed"
	"errors"
	"fmt"
	"github.com/shank318/doota/agents"
	"github.com/shank318/doota/auth"
	"github.com/shank318/doota/auth/middleware"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/errorx"
	"github.com/shank318/doota/pb/doota/portal/v1/pbportalconnect"
	"net/http"
	"regexp"
	"strings"

	"connectrpc.com/connect"

	dgrpcserver "github.com/streamingfast/dgrpc/server"
	"github.com/streamingfast/dgrpc/server/connectrpc"
	"github.com/streamingfast/shutter"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	*shutter.Shutter
	httpListenAddr    string
	authenticator     *auth.Authenticator
	corsURLRegexAllow *regexp.Regexp
	isAppReady        func() bool
	logger            *zap.Logger
}

func New(
	httpListenAddr string,
	authenticator *auth.Authenticator,
	corsURLRegexAllow *regexp.Regexp,
	isAppReady func() bool,
	logger *zap.Logger,
) *Server {
	return &Server{
		Shutter:           shutter.New(),
		authenticator:     authenticator,
		httpListenAddr:    httpListenAddr,
		corsURLRegexAllow: corsURLRegexAllow,
		isAppReady:        isAppReady,
		logger:            logger,
	}
}

type AgentHandler func(agent agents.AIAgent) http.HandlerFunc

// this is a blocking call
func (s *Server) Run(
	portalHandler pbportalconnect.PortalServiceHandler,
	callStatusUpdateHandler AgentHandler,
	endConversationHandler AgentHandler,
	adminBatchUpload AgentHandler,
) {
	tracerProvider := otel.GetTracerProvider()
	options := []dgrpcserver.Option{
		dgrpcserver.WithLogger(s.logger),
		dgrpcserver.WithHealthCheck(dgrpcserver.HealthCheckOverGRPC|dgrpcserver.HealthCheckOverHTTP, s.healthzHandler()),
		dgrpcserver.WithPostUnaryInterceptor(otelgrpc.UnaryServerInterceptor(otelgrpc.WithTracerProvider(tracerProvider))),
		dgrpcserver.WithPostStreamInterceptor(otelgrpc.StreamServerInterceptor(otelgrpc.WithTracerProvider(tracerProvider))),
		dgrpcserver.WithGRPCServerOptions(grpc.MaxRecvMsgSize(25 * 1024 * 1024)),
		// TODO: Uncomment when auth is implemented
		dgrpcserver.WithConnectInterceptor(middleware.NewAuthInterceptor(s.authenticator, s.logger)),
		dgrpcserver.WithConnectInterceptor(connectrpc.NewErrorsInterceptor(s.logger, connectrpc.WithErrorMapper(func(err error) error {

			if baseError := (*errorx.BaseError)(nil); errors.As(err, &baseError) {
				return errorx.BaseErrToConnectErr(baseError)
			}

			if errors.Is(err, datastore.NotFound) {
				return connect.NewError(connect.CodeNotFound, err)
			}

			if errors.Is(err, datastore.ErrMessageSourceAlreadyExists) {
				return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("Message Sources already configured for this user"))
			}

			return err
		}))),
		dgrpcserver.WithConnectCORS(s.corsOption()),
	}
	if strings.Contains(s.httpListenAddr, "*") {
		s.logger.Info("grpc server with insecure server")
		options = append(options, dgrpcserver.WithInsecureServer())
	} else {
		s.logger.Info("grpc server with plain text server")
		options = append(options, dgrpcserver.WithPlainTextServer())
	}

	options = append(options,
		dgrpcserver.WithConnectWebHTTPHandlers([]dgrpcserver.HTTPHandlerGetter{
			func() (string, http.Handler) {
				return "/webhook/vana/call_status/{id}", callStatusUpdateHandler(agents.AIAgentVANA)
			},
			func() (string, http.Handler) {
				return "/webhook/vana/end_conversation/{id}", endConversationHandler(agents.AIAgentVANA)
			},
			func() (string, http.Handler) {
				return "/vana/admin/batch", adminBatchUpload(agents.AIAgentVANA)
			},
		}),
	)

	portalHandlerGetter := func(opts ...connect.HandlerOption) (string, http.Handler) {
		return pbportalconnect.NewPortalServiceHandler(portalHandler, opts...)
	}

	srv := connectrpc.New([]connectrpc.HandlerGetter{
		portalHandlerGetter,
	}, options...)

	s.OnTerminating(func(_ error) {
		s.logger.Info("shutting down connect web server")
		srv.Shutdown(nil)
	})

	addr := strings.ReplaceAll(s.httpListenAddr, "*", "")
	srv.Launch(addr)
	<-srv.Terminated()
}
