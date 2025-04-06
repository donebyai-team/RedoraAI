package portal

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/services"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

func (p *Portal) CreateCustomerCase(ctx context.Context, c *connect.Request[pbportal.CreateCustomerCaseReq]) (*connect.Response[emptypb.Empty], error) {
	t, err := time.Parse(time.DateOnly, c.Msg.DueDate)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("we ran into an issue, please try again : %w", err))
	}

	req := services.CreateCustomerCase{
		FirstName:  c.Msg.FirstName,
		LastName:   c.Msg.LastName,
		Phone:      c.Msg.Phone,
		OrgID:      c.Msg.OrganizationId,
		PromptType: c.Msg.PromptType,
		DueDate:    t,
	}
	err = p.customerCaseService.Create(ctx, &req)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to create customerCase: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

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
