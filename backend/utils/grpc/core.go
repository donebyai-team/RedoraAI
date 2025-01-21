package grpc

import (
	"context"

	"github.com/shank318/doota/auth"
	authgrpc "github.com/shank318/doota/auth/grpc"
	"github.com/shank318/doota/errorx"
	"github.com/streamingfast/dgrpc"
	dgrpcserver "github.com/streamingfast/dgrpc/server"
	"github.com/streamingfast/dgrpc/server/factory"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const MaxMsgSize = 100 * 1024 * 1024 //100MB

func InternalGRPCServerConn(logger *zap.Logger, healthCheck dgrpcserver.HealthCheck) dgrpcserver.Server {
	return factory.ServerFromOptions(
		dgrpcserver.WithLogger(logger),
		dgrpcserver.WithHealthCheck(dgrpcserver.HealthCheckOverGRPC|dgrpcserver.HealthCheckOverHTTP, healthCheck),
		dgrpcserver.WithPostUnaryInterceptor(authgrpc.UnaryServerInterceptor()),
		dgrpcserver.WithPostStreamInterceptor(authgrpc.StreamServerInterceptor()),
		dgrpcserver.WithPostUnaryInterceptor(errorx.ErrorUnaryServerInterceptor()),
		dgrpcserver.WithPostStreamInterceptor(errorx.ErrorStreamServerInterceptor()),
		dgrpcserver.WithGRPCServerOptions(grpc.MaxRecvMsgSize(MaxMsgSize)),
	)
}

func NewInternalClientConn(grpcListenAddr string) (*grpc.ClientConn, error) {
	options := []grpc.DialOption{
		grpc.WithUnaryInterceptor(loadLogicClientInterceptor()),
		grpc.WithStreamInterceptor(loadLogicStreamInterceptor()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MaxMsgSize)),
	}
	return dgrpc.NewInternalClientConn(grpcListenAddr, options...)
}

func loadLogicClientInterceptor() grpc.UnaryClientInterceptor {

	otelUnaryInterceptorFunc := otelgrpc.UnaryClientInterceptor()
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		authCred, ok := auth.FromContext(ctx)
		if ok {
			ctx = auth.ToOutgoingGRPCContext(ctx, authCred.Identity())
		}

		err := otelUnaryInterceptorFunc(ctx, method, req, reply, cc, invoker, opts...)
		if err != nil {
			return errorx.GrpcErrToBase(err)
		}
		return nil

	}
}

func loadLogicStreamInterceptor() grpc.StreamClientInterceptor {
	otelStreamClientInterceptorFunc := otelgrpc.StreamClientInterceptor()
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		authCred, ok := auth.FromContext(ctx)
		if ok {
			ctx = auth.ToOutgoingGRPCContext(ctx, authCred.Identity())
		}

		stream, err := otelStreamClientInterceptorFunc(ctx, desc, cc, method, streamer, opts...)
		if err != nil {
			return nil, errorx.GrpcErrToBase(err)
		}
		return stream, nil
	}
}
