package portal

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/services"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (p *Portal) CreateKeyword(ctx context.Context, c *connect.Request[pbportal.CreateKeywordReq]) (*connect.Response[emptypb.Empty], error) {
	req := services.CreateKeyword{
		Keyword: c.Msg.Keyword,
		OrgID:   c.Msg.OrganizationId,
	}

	_, err := p.keywordService.CreateKeyword(ctx, &req)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to create keyword: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (p *Portal) AddSubReddit(ctx context.Context, c *connect.Request[pbportal.AddSubRedditRequest]) (*connect.Response[emptypb.Empty], error) {
	//TODO implement me
	panic("implement me")
}

func (p *Portal) GetSubReddits(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.GetSubredditsResponse], error) {
	//TODO implement me
	panic("implement me")
}
