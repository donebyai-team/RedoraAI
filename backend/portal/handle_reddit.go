package portal

import (
	"connectrpc.com/connect"
	"context"
	"fmt"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/services"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (p *Portal) CreateKeyword(ctx context.Context, c *connect.Request[pbportal.CreateKeywordReq]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	req := services.CreateKeyword{
		Keyword: c.Msg.Keyword,
		OrgID:   actor.OrganizationID,
	}

	_, err = p.keywordService.CreateKeyword(ctx, &req)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to create keyword: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (p *Portal) AddSubReddit(ctx context.Context, c *connect.Request[pbportal.AddSubRedditRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	redditClient, err := p.redditOauthClient.NewRedditClient(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}
	redditService := services.NewRedditService(p.logger, p.db, redditClient)
	err = redditService.CreateSubReddit(ctx, &models.SubReddit{
		OrganizationID: actor.OrganizationID,
		URL:            c.Msg.Url,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to add subreddit: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (p *Portal) GetSubReddits(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.GetSubredditsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	redditClient, err := p.redditOauthClient.NewRedditClient(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}
	redditService := services.NewRedditService(p.logger, p.db, redditClient)
	subReddits, err := redditService.GetSubReddits(ctx, actor.OrganizationID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to get subreddits: %w", err))
	}
	subRedditProto := make([]*pbportal.SubReddit, 0, len(subReddits))
	for _, subReddit := range subReddits {
		subRedditProto = append(subRedditProto, &pbportal.SubReddit{
			Id:          subReddit.ID,
			Url:         subReddit.URL,
			Name:        subReddit.Name,
			Description: subReddit.Description,
			Metadata:    &pbportal.SubRedditMetadata{},
			Title:       subReddit.Title,
		})
	}

	return connect.NewResponse(&pbportal.GetSubredditsResponse{Subreddits: subRedditProto}), nil
}

func (p *Portal) RemoveSubReddit(ctx context.Context, c *connect.Request[pbportal.RemoveSubRedditRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	redditClient, err := p.redditOauthClient.NewRedditClient(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}
	redditService := services.NewRedditService(p.logger, p.db, redditClient)
	err = redditService.RemoveSubReddit(ctx, c.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to add subreddit: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
