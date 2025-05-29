package portal

import (
	"connectrpc.com/connect"
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/agents/redora"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/services"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"net/http"
	"net/url"
	"strings"
)

func (p *Portal) getProject(ctx context.Context, headers http.Header, orgID string) (*models.Project, error) {
	var project *models.Project
	if in := headers.Get("X-Project-Id"); in != "" {
		project, err := p.db.GetProject(ctx, in)
		if err != nil {
			return nil, fmt.Errorf("failed to get project by org id: %w", err)
		}
		project = project
	} else {
		// TODO: For now, since we have only one project per org. This is a workaround
		// Remove it later and make X-Project-Id mandatory on frontend
		projects, err := p.db.GetProjects(ctx, orgID)
		if err != nil {
			return nil, fmt.Errorf("failed to get project by org id: %w", err)
		}
		if len(projects) == 0 {
			return nil, status.New(codes.PermissionDenied, "no project not found").Err()
		}
		project = projects[0]
	}

	return project, nil
}

func (p *Portal) CreateOrEditProject(ctx context.Context, c *connect.Request[pbportal.CreateProjectRequest]) (*connect.Response[pbcore.Project], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	if !utils.IsValidProductName(c.Msg.Name) {
		return nil, status.New(codes.InvalidArgument, "project name should be between 3 and 30 characters or max 3 words").Err()
	}

	if len(c.Msg.Description) < 10 {
		return nil, status.New(codes.InvalidArgument, "project description should be at least 10 characters").Err()
	}

	if len(c.Msg.TargetPersona) < 10 {
		return nil, status.New(codes.InvalidArgument, "project target persona should be at least 10 characters").Err()
	}

	if !strings.HasPrefix(c.Msg.Website, "http://") && !strings.HasPrefix(c.Msg.Website, "https://") {
		c.Msg.Website = "https://" + c.Msg.Website
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

		// reset the suggestions if the project name or description or target has changed
		if existingProject.Name != c.Msg.Name ||
			existingProject.ProductDescription != c.Msg.Description ||
			existingProject.CustomerPersona != c.Msg.TargetPersona {
			existingProject.Metadata.SuggestedKeywords = []string{}
			existingProject.Metadata.SuggestedSubReddits = []string{}
		}

		project, err = p.db.UpdateProject(ctx, &models.Project{
			OrganizationID:     actor.OrganizationID,
			Name:               c.Msg.Name,
			ProductDescription: c.Msg.Description,
			CustomerPersona:    c.Msg.TargetPersona,
			WebsiteURL:         c.Msg.Website,
			Metadata:           existingProject.Metadata,
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

		// notify admin
		go p.alertNotifier.SendNewProductAddedAlert(context.Background(), project.Name, project.WebsiteURL)
	}

	// TODO: Remove this section later
	if len(project.Metadata.SuggestedKeywords) == 0 || len(project.Metadata.SuggestedSubReddits) == 0 {
		p.logger.Info("suggesting keywords", zap.String("project_id", project.ID))
		suggestions, usage, err := p.aiClient.SuggestKeywordsAndSubreddits(ctx, p.aiClient.GetAdvanceModel(), project, p.logger)
		if err != nil {
			p.logger.Error("failed to get keyword suggestions", zap.Error(err))
		}

		if suggestions != nil {
			p.logger.Info("adding keyword suggestions",
				zap.String("model_used", string(usage.Model)),
				zap.Int("num_suggestions", len(suggestions.Keywords)),
				zap.Int("num_subreddits", len(suggestions.Subreddits)))

			for _, keyword := range suggestions.Keywords {
				if keyword.Keyword == "" {
					continue
				}
				project.Metadata.SuggestedKeywords = append(project.Metadata.SuggestedKeywords, keyword.Keyword)
			}

			for _, subreddit := range suggestions.Subreddits {
				if subreddit.Subreddit == "" {
					continue
				}
				if !strings.HasPrefix(subreddit.Subreddit, "r/") {
					subreddit.Subreddit = "r/" + subreddit.Subreddit
				}

				project.Metadata.SuggestedSubReddits = append(project.Metadata.SuggestedSubReddits, subreddit.Subreddit)
			}

			// Update project metadata
			p.db.UpdateProject(ctx, project)
		}
	}
	// =====

	projectProto, err := p.projectToProto(ctx, project)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(projectProto), nil
}

func (p *Portal) SuggestKeywordsAndSources(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbcore.Project], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	if len(project.Metadata.SuggestedKeywords) == 0 || len(project.Metadata.SuggestedSubReddits) == 0 {
		p.logger.Info("suggesting keywords", zap.String("project_id", project.ID))
		suggestions, usage, err := p.aiClient.SuggestKeywordsAndSubreddits(ctx, p.aiClient.GetAdvanceModel(), project, p.logger)
		if err != nil {
			p.logger.Error("failed to get keyword suggestions", zap.Error(err))
		}

		if suggestions != nil {
			p.logger.Info("adding keyword suggestions",
				zap.String("model_used", string(usage.Model)),
				zap.Int("num_suggestions", len(suggestions.Keywords)),
				zap.Int("num_subreddits", len(suggestions.Subreddits)))

			for _, keyword := range suggestions.Keywords {
				if keyword.Keyword == "" {
					continue
				}
				project.Metadata.SuggestedKeywords = append(project.Metadata.SuggestedKeywords, keyword.Keyword)
			}

			for _, subreddit := range suggestions.Subreddits {
				if subreddit.Subreddit == "" {
					continue
				}
				if !strings.HasPrefix(subreddit.Subreddit, "r/") {
					subreddit.Subreddit = "r/" + subreddit.Subreddit
				}

				project.Metadata.SuggestedSubReddits = append(project.Metadata.SuggestedSubReddits, subreddit.Subreddit)
			}

			// Update project metadata
			p.db.UpdateProject(ctx, project)
		}
	}
	projectProto, err := p.projectToProto(ctx, project)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(projectProto), nil
}

func (p *Portal) getProjects(ctx context.Context, orgID string) ([]*pbcore.Project, bool, error) {
	projects, err := p.db.GetProjects(ctx, orgID)
	if err != nil {
		return nil, false, err
	}

	projectsProtos := make([]*pbcore.Project, 0, len(projects))
	isOnboardingDone := false
	for _, project := range projects {
		projectProto, err := p.projectToProto(ctx, project)
		if err != nil {
			return nil, isOnboardingDone, err
		}

		if len(projectProto.Sources) > 0 && len(projectProto.Keywords) > 0 {
			isOnboardingDone = true
		}

		projectsProtos = append(projectsProtos, projectProto)
	}

	return projectsProtos, isOnboardingDone, nil
}

func (p *Portal) projectToProto(ctx context.Context, project *models.Project) (*pbcore.Project, error) {
	keywords, err := p.db.GetKeywords(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	sources, err := p.db.GetSourcesByProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	return new(pbcore.Project).FromModel(project, sources, keywords), nil
}

func (p *Portal) CreateKeywords(ctx context.Context, c *connect.Request[pbportal.CreateKeywordReq]) (*connect.Response[pbportal.CreateKeywordsRes], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	if len(c.Msg.Keywords) == 0 {
		return nil, status.New(codes.InvalidArgument, "at least one keyword is required").Err()
	}

	if len(c.Msg.Keywords) > 5 {
		return nil, status.New(codes.InvalidArgument, "maximum 5 keywords are allowed").Err()
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	for _, keyword := range c.Msg.Keywords {
		err = utils.ValidateKeyword(keyword)
		if err != nil {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
	}

	err = p.db.CreateKeywords(ctx, project.ID, c.Msg.Keywords)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to create keyword: %w", err))
	}

	keywords, err := p.db.GetKeywords(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	keywordProto := make([]*pbcore.Keyword, 0, len(keywords))
	for _, keyword := range keywords {
		keywordProto = append(keywordProto, new(pbcore.Keyword).FromModel(keyword))
	}

	return connect.NewResponse(&pbportal.CreateKeywordsRes{Keywords: keywordProto}), nil
}

func (p *Portal) AddSource(ctx context.Context, c *connect.Request[pbportal.AddSourceRequest]) (*connect.Response[pbcore.Source], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	sources, err := p.db.GetSourcesByProject(ctx, project.ID)
	if err != nil {
		return nil, err
	}

	if len(sources) >= 5 {
		return nil, status.New(codes.InvalidArgument, "maximum 5 sources are allowed").Err()
	}

	redditClient, err := p.redditOauthClient.GetOrCreate(ctx, actor.OrganizationID, false)
	if err != nil {
		return nil, err
	}
	redditService := services.NewRedditService(p.logger, p.db, redditClient, p.aiClient, p.cache)

	source := &models.Source{
		ProjectID: project.ID,
		Name:      utils.CleanSubredditName(c.Msg.Name),
		OrgID:     actor.OrganizationID,
	}
	err = redditService.CreateSubReddit(ctx, source)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(new(pbcore.Source).FromModel(source, new(pbcore.Source_RedditMetadata).FromModel(&source.Metadata))), nil
}

func (p *Portal) GetSources(ctx context.Context, c *connect.Request[emptypb.Empty]) (*connect.Response[pbportal.GetSourceResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	redditService := services.NewRedditService(p.logger, p.db, nil, nil, nil)
	sources, err := redditService.GetSubReddits(ctx, project.ID)
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
	_, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	redditService := services.NewRedditService(p.logger, p.db, nil, nil, nil)
	err = redditService.RemoveSubReddit(ctx, c.Msg.Id)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to add subreddit: %w", err))
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (p *Portal) GetRelevantLeads(ctx context.Context, c *connect.Request[pbportal.GetRelevantLeadsRequest]) (*connect.Response[pbportal.GetLeadsResponse], error) {
	status, err := models.ParseLeadStatus(c.Msg.Status.String())
	if err != nil {
		return nil, err
	}

	if status != models.LeadStatusNEW {
		return p.GetLeadsByStatus(ctx, &connect.Request[pbportal.GetLeadsByStatusRequest]{
			Msg: &pbportal.GetLeadsByStatusRequest{
				Status:    c.Msg.Status,
				PageNo:    c.Msg.PageNo,
				DateRange: c.Msg.DateRange,
			},
		})
	}

	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	subReddits := []string{}
	if c.Msg.SubReddit != nil {
		subReddits = append(subReddits, *c.Msg.SubReddit)
	}

	leads, err := p.db.GetLeadsByRelevancy(ctx, project.ID, datastore.LeadsFilter{
		RelevancyScore: c.Msg.RelevancyScore,
		Sources:        subReddits,
		Limit:          pageCount,
		DateRange:      c.Msg.DateRange,
		Offset:         int(c.Msg.PageNo),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetch leads: %w", err))
	}

	leadsProto := make([]*pbcore.Lead, 0, len(leads))
	for _, lead := range leads {
		leadsProto = append(leadsProto, new(pbcore.Lead).FromModel(redactPlatformOnlyMetadata(actor.Role, lead)))
	}

	analysis, err := redora.NewLeadAnalysis(p.db, p.logger).GenerateLeadAnalysis(ctx, project.ID, c.Msg.DateRange)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pbportal.GetLeadsResponse{
		Leads:    leadsProto,
		Analysis: analysis,
	}), nil
}

func redactPlatformOnlyMetadata(role models.UserRole, lead *models.AugmentedLead) *models.AugmentedLead {
	if role != models.UserRolePLATFORMADMIN {
		lead.LeadMetadata.RelevancyLLMModel = ""
		lead.LeadMetadata.CommentLLMModel = ""
		lead.LeadMetadata.DMLLMModel = ""
		lead.LeadMetadata.LLMModelResponseOverriddenBy = ""
	}
	return lead
}

func (p *Portal) UpdateAutomationSettings(ctx context.Context, c *connect.Request[pbportal.UpdateAutomationSettingRequest]) (*connect.Response[pbportal.Organization], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	org, err := p.db.GetOrganizationById(ctx, actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	if c.Msg.Comment != nil {
		if c.Msg.Comment.RelevancyScore < 80 {
			return nil, status.New(codes.InvalidArgument, "relevancy score should be at least 80").Err()
		}

		maxAllowedCommentPerDay := org.FeatureFlags.GetSubscriptionPlanMetadata().Comments.PerDay

		if c.Msg.Comment.MaxPerDay > maxAllowedCommentPerDay {
			return nil, status.New(codes.InvalidArgument, fmt.Sprintf("max %s automated comments allows as per the subscribed plan", maxAllowedCommentPerDay)).Err()
		}
		// If 0 is given, we default to the max allowed
		org.FeatureFlags.MaxCommentsPerDay = c.Msg.Comment.MaxPerDay
		org.FeatureFlags.EnableAutoComment = c.Msg.Comment.Enabled
		org.FeatureFlags.RelevancyScoreComment = float64(c.Msg.Comment.RelevancyScore)
	}

	//if c.Msg.Dm != nil {
	//	if c.Msg.Dm.RelevancyScore < 70 {
	//		return nil, status.New(codes.InvalidArgument, "relevancy score should be at least 70").Err()
	//	}
	//	org.FeatureFlags.EnableAutoDM = c.Msg.Dm.Enabled
	//	org.FeatureFlags.RelevancyScoreDM = float64(c.Msg.Dm.RelevancyScore)
	//}

	err = p.db.UpdateOrganization(ctx, org)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(new(pbportal.Organization).FromModel(org)), nil
}

const pageCount = 200

func (p *Portal) GetLeadsByStatus(ctx context.Context, c *connect.Request[pbportal.GetLeadsByStatusRequest]) (*connect.Response[pbportal.GetLeadsResponse], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	status, err := models.ParseLeadStatus(c.Msg.Status.String())
	if err != nil {
		return nil, err
	}
	leads, err := p.db.GetLeadsByStatus(ctx, project.ID, datastore.LeadsFilter{
		Status:    status,
		Limit:     pageCount,
		DateRange: c.Msg.DateRange,
		Offset:    int(c.Msg.PageNo), // starting with 0
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("unable to fetch leads: %w", err))
	}

	leadsProto := make([]*pbcore.Lead, 0, len(leads))
	for _, lead := range leads {
		leadsProto = append(leadsProto, new(pbcore.Lead).FromModel(redactPlatformOnlyMetadata(actor.Role, lead)))
	}

	analysis, err := redora.NewLeadAnalysis(p.db, p.logger).GenerateLeadAnalysis(ctx, project.ID, c.Msg.DateRange)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&pbportal.GetLeadsResponse{Leads: leadsProto, Analysis: analysis}), nil
}

func (p *Portal) UpdateLeadStatus(ctx context.Context, c *connect.Request[pbportal.UpdateLeadStatusRequest]) (*connect.Response[emptypb.Empty], error) {
	actor, err := p.gethAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	project, err := p.getProject(ctx, c.Header(), actor.OrganizationID)
	if err != nil {
		return nil, err
	}

	lead, err := p.db.GetLeadByID(ctx, project.ID, c.Msg.LeadId)
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
