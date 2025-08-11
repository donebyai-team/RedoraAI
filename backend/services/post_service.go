package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
)

type PostService interface {
	CreatePost(ctx context.Context, post *models.Post, project *models.Project) (*models.Post, error)
	DeletePost(ctx context.Context, postID string) error
	UpdatePost(ctx context.Context, updated *models.Post) (*models.Post, error)
}
type postService struct {
	aiClient          *ai.Client
	db                datastore.Repository
	logger            *zap.Logger
	redditOauthClient *reddit.OauthClient
}

func NewPostService(logger *zap.Logger, db datastore.Repository, aiClient *ai.Client, redditOauthClient *reddit.OauthClient) *postService {
	return &postService{logger: logger, db: db, aiClient: aiClient, redditOauthClient: redditOauthClient}
}

func (s *postService) CreatePost(ctx context.Context, post *models.Post, project *models.Project) (*models.Post, error) {
	source, err := s.db.GetSourceByID(ctx, post.SourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source by ID: %w", err)
	}

	var existingPost *models.Post
	if post.ID != "" {
		existingPost, err = s.db.GetPostByID(ctx, post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing post: %w", err)
		}
	}

	postRequirements, flairs, err := s.fetchPostRequirementsAndFlairs(ctx, post, project, source, existingPost)
	if err != nil {
		return nil, err
	}

	flairTexts := extractFlairTexts(flairs)

	// Prepare AI post generation input
	input := &ai.PostGenerateInput{
		Id:          post.ID,
		Project:     project,
		PostSetting: &post.Metadata.Settings,
		Rules:       postRequirements.ToRules(),
		Flairs:      flairTexts,
	}

	resp, _, err := s.aiClient.GeneratePost(ctx, s.aiClient.GetDefaultModel(), input, s.logger)

	if err != nil {
		return nil, fmt.Errorf("generate post failed: %w", err)
	}

	if strings.TrimSpace(resp.Title) == "" || strings.TrimSpace(resp.Description) == "" {
		return nil, fmt.Errorf("generated post is invalid: title or description is empty")
	}

	// Find flair ID by matching AI-selected flair text
	var selectedFlairID string
	if resp.SelectedFlair != "" && !strings.EqualFold(strings.TrimSpace(resp.SelectedFlair), reddit.DummyFlair) {
		for _, f := range flairs {
			if strings.EqualFold(strings.TrimSpace(f.Text), strings.TrimSpace(resp.SelectedFlair)) {
				selectedFlairID = f.ID
				break
			}
		}
	}

	settings := post.Metadata.Settings
	settings.FlairID = &selectedFlairID

	if existingPost != nil {
		// Append generated result to history but DO NOT update existing title/desc/settings
		historyEntry := models.PostRegenerationHistory{
			PostSettings: settings,
			Title:        resp.Title,
			Description:  resp.Description,
		}

		existingPost.Metadata.History = append(existingPost.Metadata.History, historyEntry)
		existingPost.Metadata.Settings.ReferenceID = settings.ReferenceID

		if err := s.db.UpdatePost(ctx, existingPost); err != nil {
			return nil, fmt.Errorf("failed to update post: %w", err)
		}
		return existingPost, nil
	}

	// New Post: Set everything and initialize history
	post.Title = resp.Title
	post.Description = resp.Description
	post.Status = models.PostStatusCREATED

	post.Metadata = models.PostMetadata{
		Settings: settings,
		History: []models.PostRegenerationHistory{
			{
				PostSettings: settings,
				Title:        resp.Title,
				Description:  resp.Description,
			},
		},
		PostRequirements: *postRequirements,
		Flairs:           flairs,
	}

	newPost, err := s.db.CreatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return newPost, nil
}

func (s *postService) fetchPostRequirementsAndFlairs(
	ctx context.Context,
	post *models.Post,
	project *models.Project,
	source *models.Source,
	existingPost *models.Post,
) (*models.PostRequirements, []models.Flair, error) {
	var postRequirement models.PostRequirements
	var flairs []models.Flair

	if existingPost != nil {
		// Regeneration → use saved metadata
		postRequirement = existingPost.Metadata.PostRequirements
		flairs = existingPost.Metadata.Flairs
		return &postRequirement, flairs, nil
	}

	// First-time post → fetch from Reddit API
	client, err := s.redditOauthClient.GetRedditAPIClient(ctx, project.OrganizationID, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Reddit API client: %w", err)
	}

	postReqPtr, err := client.GetPostRequirements(ctx, source.Name)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch post requirements: %w", err)
	}

	postRequirement = *postReqPtr

	if postRequirement.IsFlairRequired == true {
		flairResp, err := client.GetSubredditFlairs(ctx, source.Name)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to fetch flairs: %w", err)
		}

		// filter non-moderator flairs
		flairs = make([]models.Flair, 0, len(flairResp))
		for _, f := range flairResp {
			if !f.ModOnly && len(f.Text) > 0 {
				flairs = append(flairs, f)
			}
		}

		// Save rules and flairs into metadata for future regenerations
		post.Metadata.PostRequirements = postRequirement
		post.Metadata.Flairs = flairs

		return &postRequirement, flairs, nil
	}

	return &postRequirement, nil, nil
}

func extractFlairTexts(flairs []models.Flair) []string {
	var texts []string
	for _, f := range flairs {
		texts = append(texts, f.Text)
	}

	// insert dummy flair in case flair is empty or not required
	if len(texts) == 0 {
		return []string{reddit.DummyFlair}
	}

	return texts
}

func (s *postService) UpdatePost(ctx context.Context, updated *models.Post) (*models.Post, error) {
	existing, err := s.db.GetPostByID(ctx, updated.ID)
	if err != nil && !errors.Is(err, datastore.NotFound) {
		return nil, fmt.Errorf("invalid post id: %w", err)
	}

	postRequirement := existing.Metadata.PostRequirements
	validationErr := postRequirement.Validate(*updated)

	if len(validationErr) != 0 {
		return nil, fmt.Errorf(strings.Join(validationErr, "\n"))
	}

	if updated.ScheduleAt != nil && updated.ScheduleAt.Before(time.Now().UTC().Add(-30*time.Second)) {
		return nil, fmt.Errorf("cannot schedule post in the past")
	}

	if existing.Status == models.PostStatusFAILED || existing.Status == models.PostStatusSENT {
		return nil, fmt.Errorf("post is in %s status, cannot update", existing.Status)
	}

	// Apply updates from input
	existing.Title = updated.Title
	existing.Description = updated.Description
	existing.SourceID = updated.SourceID
	existing.ReferenceID = updated.ReferenceID
	existing.Metadata.Settings = updated.Metadata.Settings

	if updated.ScheduleAt != nil {
		existing.ScheduleAt = updated.ScheduleAt
		existing.Status = models.PostStatusSCHEDULED
	}

	// Save the update
	if err := s.db.UpdatePost(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return existing, nil
}

func (s *postService) DeletePost(ctx context.Context, postID string) error {
	_, err := s.db.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to fetch post: %w", err)
	}

	// Set deleted_at to soft delete
	if err := s.db.DeletePostByID(ctx, postID); err != nil {
		return fmt.Errorf("failed to soft delete post: %w", err)
	}

	return nil
}
