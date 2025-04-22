package errorx

import (
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	"google.golang.org/protobuf/proto"
)

type BaseError struct {
	id         pbcore.PlatformError
	details    proto.Message
	hasDetails bool
}

func (b BaseError) ID() pbcore.PlatformError {
	return b.id
}

func (b BaseError) Details() proto.Message {
	return b.details
}

func (b BaseError) HasDetails() bool {
	return b.hasDetails
}

func (b BaseError) Error() string {
	return b.id.String()
}

func NewErrMessageAlreadyExists() BaseError {
	return BaseError{id: pbcore.PlatformError_PLATFORM_ERROR_MESSAGE_ALREADY_EXISTS}
}
