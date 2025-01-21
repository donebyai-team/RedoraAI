package auth

import (
	"context"
	"encoding/base64"
	"net/url"

	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

const (
	LLHeaderIdentity string = "x-ll-identity"
)

const authenticatedIdentityKey int = 0

var emptyMetadata = metadata.New(nil)

func ToOutgoingGRPCContext(ctx context.Context, identity *pbcore.Identity) context.Context {
	cnt, err := proto.Marshal(identity)
	if err != nil {
		panic(err)
	}

	headers := map[string]string{
		LLHeaderIdentity: base64.URLEncoding.EncodeToString(cnt),
	}
	return metadata.NewOutgoingContext(ctx, metadata.New(headers))
}

func IdentityFromGPRPContext(ctx context.Context) context.Context {

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = emptyMetadata
	}
	values := url.Values(md)

	identityProto := values.Get(LLHeaderIdentity)
	if identityProto == "" {
		return ctx
	}

	cnt, err := base64.URLEncoding.DecodeString(identityProto)
	if err != nil {
		panic(err)
	}

	identity := &pbcore.Identity{}
	if err := proto.Unmarshal(cnt, identity); err != nil {
		panic(err)
	}

	return context.WithValue(ctx, authenticatedIdentityKey, identity)
}

func IdentityFromContext(ctx context.Context) (*pbcore.Identity, bool) {
	identity, ok := ctx.Value(authenticatedIdentityKey).(*pbcore.Identity)
	return identity, ok
}
