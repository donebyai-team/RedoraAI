package portal

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	pbreddit "github.com/shank318/doota/pb/doota/reddit/v1"
	"github.com/shank318/doota/services"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (p *Portal) CreateKeyword(ctx context.Context, c *connect.Request[pbportal.CreateKeywordReq]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	req := services.CreateKeyword{
		Keyword:   c.Msg.Keyword,
		ProjectID: actor.ProjectID,
	}

	_, err = p.keywordService.CreateKeyword(ctx, &req)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to create keyword: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (p *Portal) AddSubReddit(ctx context.Context, c *connect.Request[pbreddit.AddSubRedditRequest]) (*connect.Response[emptypb.Empty], error) {
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
		ProjectID: actor.ProjectID,
		URL:       c.Msg.Url,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to add subreddit: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (p *Portal) GetSubReddits(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbreddit.GetSubredditsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	redditClient, err := p.redditOauthClient.NewRedditClient(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}
	redditService := services.NewRedditService(p.logger, p.db, redditClient)
	subReddits, err := redditService.GetSubReddits(ctx, actor.ProjectID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to get subreddits: %w", err))
	}
	subRedditProto := make([]*pbreddit.SubReddit, 0, len(subReddits))
	for _, subReddit := range subReddits {
		subRedditProto = append(subRedditProto, &pbreddit.SubReddit{
			Id:          subReddit.ID,
			Url:         subReddit.URL,
			Name:        subReddit.Name,
			Description: subReddit.Description,
			Metadata:    &pbreddit.SubRedditMetadata{},
			Title:       subReddit.Title,
		})
	}

	return connect.NewResponse(&pbreddit.GetSubredditsResponse{Subreddits: subRedditProto}), nil
}

func (p *Portal) RemoveSubReddit(ctx context.Context, c *connect.Request[pbreddit.RemoveSubRedditRequest]) (*connect.Response[emptypb.Empty], error) {
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

func (p *Portal) GetRelevantLeads(ctx context.Context, c *connect.Request[pbreddit.GetRelevantLeadsRequest]) (*connect.Response[pbreddit.GetLeadsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	subReddits := []string{}
	if c.Msg.SubReddit != nil {
		subReddits = append(subReddits, *c.Msg.SubReddit)
	}

	leads, err := p.db.GetRedditLeadsByRelevancy(ctx, actor.ProjectID, c.Msg.RelevancyScore, subReddits)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetch leads: %w", err))
	}

	leadsProto := make([]*pbreddit.RedditLead, 0, len(leads))
	for _, lead := range leads {
		leadsProto = append(leadsProto, new(pbreddit.RedditLead).FromModel(lead))
	}

	return connect.NewResponse(&pbreddit.GetLeadsResponse{Leads: leadsProto}), nil
}

func (p *Portal) GetLeadsByStatus(ctx context.Context, c *connect.Request[pbreddit.GetLeadsByStatusRequest]) (*connect.Response[pbreddit.GetLeadsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	status, err := models.ParseLeadStatus(c.Msg.Status.String())
	if err != nil {
		return nil, err
	}

	leads, err := p.db.GetRedditLeadsByStatus(ctx, actor.ProjectID, status)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetch leads: %w", err))
	}

	leadsProto := make([]*pbreddit.RedditLead, 0, len(leads))
	for _, lead := range leads {
		leadsProto = append(leadsProto, new(pbreddit.RedditLead).FromModel(lead))
	}

	return connect.NewResponse(&pbreddit.GetLeadsResponse{Leads: leadsProto}), nil
}

func (p *Portal) UpdateLeadStatus(ctx context.Context, c *connect.Request[pbreddit.UpdateLeadStatusRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	lead, err := p.db.GetRedditLeadByID(ctx, actor.ProjectID, c.Msg.LeadId)
	if !errors.Is(err, datastore.NotFound) {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetch lead: %w", err))
	}

	if lead == nil {
		return connect.NewResponse(&emptypb.Empty{}), nil
	}

	status, err := models.ParseLeadStatus(c.Msg.Status.String())
	if err != nil {
		return nil, err
	}

	lead.Status = status
	err = p.db.UpdateRedditLeadStatus(ctx, lead)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to update lead status: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
