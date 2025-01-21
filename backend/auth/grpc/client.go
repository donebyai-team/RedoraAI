package grpc

import (
	"context"

	"github.com/shank318/doota/auth"
	"google.golang.org/grpc"
)

func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		authCred, ok := auth.FromContext(ctx)
		if ok {
			ctx = auth.ToOutgoingGRPCContext(ctx, authCred.Identity())
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		authCred, ok := auth.FromContext(ctx)
		if ok {
			ctx = auth.ToOutgoingGRPCContext(ctx, authCred.Identity())
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
