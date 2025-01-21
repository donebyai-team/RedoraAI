package grpc

import (
	"context"
	"github.com/shank318/doota/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var EmptyMetadata = metadata.New(nil)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = auth.IdentityFromGPRPContext(ctx)
		return handler(ctx, req)
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		ctx = auth.IdentityFromGPRPContext(ctx)
		return handler(ctx, ss)
	}
}
