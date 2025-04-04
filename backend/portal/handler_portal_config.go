package portal

import (
	"context"

	"connectrpc.com/connect"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (p *Portal) GetConfig(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.Config], error) {
	return connect.NewResponse(p.config), nil
}
