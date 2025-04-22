package pbcore

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/anypb"
)

const errorRegistryFile = "pb/freighstream/core/error_registry.go"

type errorRegistry struct {
	types *protoregistry.Types
}

func newErrorRegistry() *errorRegistry {
	return &errorRegistry{
		types: &protoregistry.Types{},
	}
}

func (r *errorRegistry) Register(msgType protoreflect.MessageType) *errorRegistry {
	if err := r.types.RegisterMessage(msgType); err != nil {
		panic(fmt.Errorf("unable to register error details %T: %w", msgType.New().Interface(), err))
	}
	return r
}

func (r *errorRegistry) Unpack(msg *anypb.Any) (out proto.Message, err error) {
	out, err = anypb.UnmarshalNew(msg, proto.UnmarshalOptions{
		Resolver: r.types,
	})
	return
}

func newPopulatedErrorRegistry() *errorRegistry {
	return newErrorRegistry()
}

var registry *errorRegistry

func ErrorRegistry() *errorRegistry {
	// Due to how Golang deals with init() functions, it's possible we get some weird error like
	// `panic: runtime error: index out of range [24] with length 0` inside `protoreflect`
	// package directly.
	//
	// This is due to init ordering problem where some needed elements from `protoreflect` are not
	// initialized yet.
	//
	// By using a `init`, we delay a bit more when the initialization happens (not prefect but
	// works today).

	if registry == nil {
		registry = newPopulatedErrorRegistry()
	}
	return registry
}
