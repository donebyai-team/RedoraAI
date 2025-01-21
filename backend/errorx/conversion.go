package errorx

import (
	"errors"
	"fmt"

	"connectrpc.com/connect"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

func BaseErrToGRPC(err error) error {
	baseErr := BaseError{}

	ok := errors.As(err, &baseErr)
	if !ok {
		return err
	}
	code := codes.Internal
	switch baseErr.id {
	case pbcore.PlatformError_PLATFORM_ERROR_UNSPECIFIED:
		code = codes.Internal
	case pbcore.PlatformError_PLATFORM_ERROR_MESSAGE_ALREADY_EXISTS:
		code = codes.AlreadyExists
	case pbcore.PlatformError_PLATFORM_ERROR_INVALID_QUOTE, pbcore.PlatformError_PLATFORM_ERROR_PRICING_OPTION_INVALID_ARG:
		code = codes.InvalidArgument
	}
	grpcErr := status.New(code, baseErr.Error())

	grpcErr, detailsErr := grpcErr.WithDetails(baseErrToProto(&baseErr))
	if detailsErr != nil {
		panic(fmt.Errorf("unexpected error attaching metadata: %w", detailsErr))
	}

	return grpcErr.Err()
}

func GrpcErrToBase(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *pbcore.PlatformErrorDetails:
			return protoToBaseErr(t)
		}
	}

	return st.Err()
}

func BaseErrToConnectErr(err *BaseError) error {
	code := connect.CodeInternal
	message := "Unexpected error. Please try later."

	switch err.id {
	case pbcore.PlatformError_PLATFORM_ERROR_UNSPECIFIED:
	case pbcore.PlatformError_PLATFORM_ERROR_MESSAGE_ALREADY_EXISTS:
		code = connect.CodeAlreadyExists
		message = "Message already exists"
	case pbcore.PlatformError_PLATFORM_ERROR_INVALID_QUOTE:
		code = connect.CodeInvalidArgument
		message = "Failed to save quote"
	case pbcore.PlatformError_PLATFORM_ERROR_PRICING_OPTION_INVALID_ARG:
		code = connect.CodeInvalidArgument
		message = "Failed to fetch pricing details"
	}

	connectErr := connect.NewError(code, errors.New(message))

	if err.HasDetails() {
		connectDetails, err := connect.NewErrorDetail(err.Details())
		if err != nil {
			panic(fmt.Errorf("unexpected connect error attaching details: %w", err))
		}
		connectErr.AddDetail(connectDetails)
	}

	return connectErr
}

func baseErrToProto(in *BaseError) *pbcore.PlatformErrorDetails {
	out := &pbcore.PlatformErrorDetails{
		Error: in.id,
	}
	if in.HasDetails() {
		pbdetails, err := anypb.New(in.Details())
		if err != nil {
			panic(fmt.Errorf("unexpected error packing details: %w", err))
		}
		out.Details = pbdetails
	}

	return out

}

func protoToBaseErr(in *pbcore.PlatformErrorDetails) *BaseError {
	out := &BaseError{
		id: in.Error,
	}
	if in.Details != nil {
		details, err := pbcore.ErrorRegistry().Unpack(in.Details)
		if err != nil {
			panic(fmt.Errorf("unexpected error unpacking details: %w", err))
		}
		out.details = details
		out.hasDetails = true
	}

	return out
}
