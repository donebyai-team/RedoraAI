package errorx

import (
	"context"

	"google.golang.org/grpc"
)

func ErrorUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		resp, err = handler(ctx, req)
		if err != nil {
			return nil, BaseErrToGRPC(err)
		}
		return resp, nil
	}
}

func ErrorStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(ss.Context(), ss)
		if err != nil {
			return BaseErrToGRPC(err)
		}
		return nil
	}
}
