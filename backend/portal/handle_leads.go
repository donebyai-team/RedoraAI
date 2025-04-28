package portal

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/services"
	"github.com/shank318/doota/utils"
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

func (p *Portal) AddSource(ctx context.Context, c *connect.Request[pbportal.AddSourceRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	redditClient, err := p.redditOauthClient.NewRedditClient(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}
	redditService := services.NewRedditService(p.logger, p.db, redditClient)
	err = redditService.CreateSubReddit(ctx, &models.Source{
		ProjectID: actor.ProjectID,
		Name:      utils.CleanSubredditName(c.Msg.Name),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to add subreddit: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (p *Portal) GetSources(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.GetSourceResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	redditClient, err := p.redditOauthClient.NewRedditClient(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}
	redditService := services.NewRedditService(p.logger, p.db, redditClient)
	sources, err := redditService.GetSubReddits(ctx, actor.ProjectID)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to get subreddits: %w", err))
	}
	sourcesProto := make([]*pbcore.Source, 0, len(sources))
	for _, source := range sources {
		sourcesProto = append(sourcesProto, new(pbcore.Source).FromModel(source, new(pbcore.Source_RedditMetadata).FromModel(&source.Metadata)))
	}

	return connect.NewResponse(&pbportal.GetSourceResponse{Sources: sourcesProto}), nil
}

func (p *Portal) RemoveSource(ctx context.Context, c *connect.Request[pbportal.RemoveSourceRequest]) (*connect.Response[emptypb.Empty], error) {
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

func (p *Portal) GetRelevantLeads(ctx context.Context, c *connect.Request[pbportal.GetRelevantLeadsRequest]) (*connect.Response[pbportal.GetLeadsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	subReddits := []string{}
	if c.Msg.SubReddit != nil {
		subReddits = append(subReddits, *c.Msg.SubReddit)
	}

	leads, err := p.db.GetLeadsByRelevancy(ctx, actor.ProjectID, c.Msg.RelevancyScore, subReddits)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetch leads: %w", err))
	}

	leadsProto := make([]*pbcore.Lead, 0, len(leads))
	for _, lead := range leads {
		leadsProto = append(leadsProto, new(pbcore.Lead).FromModel(lead))
	}

	return connect.NewResponse(&pbportal.GetLeadsResponse{Leads: leadsProto}), nil
}

func (p *Portal) GetLeadsByStatus(ctx context.Context, c *connect.Request[pbportal.GetLeadsByStatusRequest]) (*connect.Response[pbportal.GetLeadsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	status, err := models.ParseLeadStatus(c.Msg.Status.String())
	if err != nil {
		return nil, err
	}

	leads, err := p.db.GetLeadsByStatus(ctx, actor.ProjectID, status)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetch leads: %w", err))
	}

	leadsProto := make([]*pbcore.Lead, 0, len(leads))
	for _, lead := range leads {
		leadsProto = append(leadsProto, new(pbcore.Lead).FromModel(lead))
	}

	return connect.NewResponse(&pbportal.GetLeadsResponse{Leads: leadsProto}), nil
}

func (p *Portal) UpdateLeadStatus(ctx context.Context, c *connect.Request[pbportal.UpdateLeadStatusRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	lead, err := p.db.GetLeadByID(ctx, actor.ProjectID, c.Msg.LeadId)
	if err != nil && !errors.Is(err, datastore.NotFound) {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetch lead: %w", err))
	}

	if lead == nil {
		return connect.NewResponse(&emptypb.Empty{}), nil
	}

	lead.Status = c.Msg.Status.ToModel()
	err = p.db.UpdateLeadStatus(ctx, lead)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to update lead status: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
