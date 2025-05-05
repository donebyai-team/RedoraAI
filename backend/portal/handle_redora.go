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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
	"net/url"
)

func (p *Portal) getProject(ctx context.Context, headers http.Header, orgID string) (string, error) {
	projectID := ""
	if in := headers.Get("X-Project-Id"); in != "" {
		project, err := p.db.GetProject(ctx, in)
		if err != nil {
			return "", fmt.Errorf("failed to get project by org id: %w", err)
		}
		projectID = project.ID
	} else {
		// TODO: For now, since we have only one project per org. This is a workaround
		// Remove it later and make X-Project-Id mandatory on frontend
		projects, err := p.db.GetProjects(ctx, orgID)
		if err != nil {
			return "", fmt.Errorf("failed to get project by org id: %w", err)
		}
		if len(projects) == 0 {
			return "", status.New(codes.PermissionDenied, "no project not found").Err()
		}
		projectID = projects[0].ID
	}

	return projectID, nil
}

func (p *Portal) CreateOrEditProject(ctx context.Context, c *connect.Request[pbportal.CreateProjectRequest]) (*connect.Response[pbcore.Project], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	if len(c.Msg.Name) < 3 {
		return nil, status.New(codes.InvalidArgument, "project name should be at least 3 character").Err()
	}

	if len(c.Msg.Description) < 10 {
		return nil, status.New(codes.InvalidArgument, "project description should be at least 10 characters").Err()
	}

	if len(c.Msg.TargetPersona) < 10 {
		return nil, status.New(codes.InvalidArgument, "project target persona should be at least 10 characters").Err()
	}

	// Validate website URL
	_, err = url.ParseRequestURI(c.Msg.Website)
	if err != nil {
		return nil, status.New(codes.InvalidArgument, "invalid website URL").Err()
	}

	var project *models.Project

	if c.Msg.Id != "" {
		existingProject, err := p.db.GetProject(ctx, c.Msg.Id)
		if err != nil {
			return nil, status.New(codes.NotFound, err.Error()).Err()
		}

		// Are we changing the name?
		if existingProject.Name != c.Msg.Name {
			existingProjectName, err := p.db.GetProjectByName(ctx, c.Msg.Name, actor.OrganizationID)
			if err != nil && !errors.Is(err, datastore.NotFound) {
				return nil, err
			}

			if existingProjectName != nil {
				return nil, status.New(codes.AlreadyExists, "a project with same name already exists").Err()
			}
		}

		project, err = p.db.UpdateProject(ctx, &models.Project{
			OrganizationID:     actor.OrganizationID,
			Name:               c.Msg.Name,
			ProductDescription: c.Msg.Description,
			CustomerPersona:    c.Msg.TargetPersona,
			WebsiteURL:         c.Msg.Website,
			ID:                 existingProject.ID,
		})
		if err != nil {
			return nil, err
		}

	} else {
		project, err = p.db.GetProjectByName(ctx, c.Msg.Name, actor.OrganizationID)
		if err != nil && !errors.Is(err, datastore.NotFound) {
			return nil, err
		}

		if project != nil {
			return nil, status.New(codes.AlreadyExists, "project already exists").Err()
		}

		project, err = p.db.CreateProject(ctx, &models.Project{
			OrganizationID:     actor.OrganizationID,
			Name:               c.Msg.Name,
			ProductDescription: c.Msg.Description,
			CustomerPersona:    c.Msg.TargetPersona,
			WebsiteURL:         c.Msg.Website,
		})

		if err != nil {
			return nil, err
		}
	}

	return connect.NewResponse(&pbcore.Project{
		Id:            project.ID,
		Name:          project.Name,
		Description:   project.ProductDescription,
		Website:       project.WebsiteURL,
		TargetPersona: project.CustomerPersona,
	}), nil
}

func (p *Portal) GetProjects(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.GetProjectsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	projects, err := p.db.GetProjects(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	projectsProto := make([]*pbcore.Project, 0, len(projects))
	isOnboardingDone := false
	for _, project := range projects {
		keywords, err := p.db.GetKeywords(ctx, project.ID)
		if err != nil {
			return nil, err
		}

		keywordsProto := make([]*pbcore.Keyword, 0, len(keywords))
		for _, keyword := range keywords {
			keywordsProto = append(keywordsProto, &pbcore.Keyword{
				Id:   keyword.ID,
				Name: keyword.Keyword,
			})
		}

		sources, err := p.db.GetSourcesByProject(ctx, project.ID)
		if err != nil {
			return nil, err
		}

		sourcesProto := make([]*pbcore.Source, 0, len(sources))
		for _, source := range sources {
			sourcesProto = append(sourcesProto, &pbcore.Source{
				Id:   source.ID,
				Name: source.Name,
			})
		}

		if len(sources) > 0 && len(keywords) > 0 {
			isOnboardingDone = true
		}

		projectsProto = append(projectsProto, &pbcore.Project{
			Id:            project.ID,
			Name:          project.Name,
			Description:   project.ProductDescription,
			Website:       project.WebsiteURL,
			TargetPersona: project.CustomerPersona,
			Keywords:      keywordsProto,
			Sources:       sourcesProto,
		})
	}

	return connect.NewResponse(&pbportal.GetProjectsResponse{
		Projects:         projectsProto,
		IsOnboardingDone: isOnboardingDone,
	}), nil
}

func (p *Portal) CreateKeyword(ctx context.Context, c *connect.Request[pbportal.CreateKeywordReq]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	projectID, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	req := services.CreateKeyword{
		Keyword:   c.Msg.Keyword,
		ProjectID: projectID,
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

	projectID, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	redditClient, err := p.redditOauthClient.NewRedditClient(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}
	redditService := services.NewRedditService(p.logger, p.db, redditClient)
	err = redditService.CreateSubReddit(ctx, &models.Source{
		ProjectID: projectID,
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

	projectID, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	redditService := services.NewRedditService(p.logger, p.db, nil)
	sources, err := redditService.GetSubReddits(ctx, projectID)
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

	projectID, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	subReddits := []string{}
	if c.Msg.SubReddit != nil {
		subReddits = append(subReddits, *c.Msg.SubReddit)
	}

	leads, err := p.db.GetLeadsByRelevancy(ctx, projectID, c.Msg.RelevancyScore, subReddits)
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

	projectID, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	status, err := models.ParseLeadStatus(c.Msg.Status.String())
	if err != nil {
		return nil, err
	}

	leads, err := p.db.GetLeadsByStatus(ctx, projectID, status)
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

	projectID, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	lead, err := p.db.GetLeadByID(ctx, projectID, c.Msg.LeadId)
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
