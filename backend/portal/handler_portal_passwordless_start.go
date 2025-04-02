package portal

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/streamingfast/logging"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (p *Portal) PasswordlessStart(ctx context.Context, c *connect.Request[pbportal.PasswordlessStartRequest]) (*connect.Response[emptypb.Empty], error) {
	logger := logging.Logger(ctx, p.logger)
	email := c.Msg.Email

	logger.Info("passwordless start", zap.String("email", email), zap.String("redirect_uri", c.Msg.RedirectUri))

	if !strings.Contains(email, "@") && !strings.Contains(email, ".") {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("invalid email address"))
	}

	ip, err := getIP(c.Header(), c.Peer().Addr)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get client IP: %w", err))
	}

	if err := p.authUsecase.StartPasswordless(ctx, c.Msg.Email, ip); err != nil {
		return nil, fmt.Errorf("failed to initiate passwordless flow: %w", err)
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Get the client's IP address
// [Reference] https://golangbyexample.com/golang-ip-address-http-request/
func getIP(header http.Header, addr string) (string, error) {
	// Get IP from the X-REAL-IP header
	ip := header.Get("X-REAL-IP")
	netIP := net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	// Get IP from X-FORWARDED-FOR header
	ips := header.Get("X-FORWARDED-FOR")
	splitIps := strings.Split(ips, ",")
	for _, ip := range splitIps {
		netIP := net.ParseIP(ip)
		if netIP != nil {
			return ip, nil
		}
	}

	// Get IP from RemoteAddr
	ip, _, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("failed to parse IP address: %w", err)
	}
	netIP = net.ParseIP(ip)
	if netIP != nil {
		return ip, nil
	}

	return "", fmt.Errorf("failed to capture client's IP address")
}
