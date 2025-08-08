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
		return nil, fmt.Errorf("failed to get source by ID")
	}

	var existingPost *models.Post
	if post.ID != "" {
		existingPost, err = s.db.GetPostByID(ctx, post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing post: %w", err)
		}
	}

	var rules []string
	var flairs []*reddit.Flair
	var postRequirement *reddit.ValidationRules

	err = s.redditOauthClient.WithRotatingAPIClient(ctx, project.OrganizationID, func(client *reddit.Client) error {
		// fetch Post requirements
		postRequirement, err = client.GetPostRequirements(ctx, source.Name)
		if err != nil {
			return fmt.Errorf("failed to fetch post requirements: %w", err)
		}

		rules = postRequirement.ToRules()

		// Fetch flairs
		flairsResp, err := client.GetSubredditFlairs(ctx, source.Name)

		if err != nil {
			return fmt.Errorf("failed to fetch flairs: %w", err)
		}
		flairs = flairsResp

		return nil
	}, reddit.MostQualifiedAccountStrategy(s.logger))
	if err != nil {
		return nil, err
	}

	var flairTexts []string
	for _, flair := range flairs {
		flairTexts = append(flairTexts, flair.Text)
	}

	// Prepare AI post generation input
	input := &ai.PostGenerateInput{
		Id:          post.ID,
		Project:     project,
		PostSetting: &post.Metadata.Settings,
		Rules:       rules,
		Flairs:      flairTexts,
	}
	// fmt.Println("Flair text", flairTexts)
	// fmt.Println("Rules text", rules)

	resp, _, err := s.aiClient.GeneratePost(ctx, s.aiClient.GetDefaultModel(), input, s.logger)
	if err != nil {
		return nil, fmt.Errorf("generate post failed: %w", err)
	}

	if strings.TrimSpace(resp.Title) == "" || strings.TrimSpace(resp.Description) == "" {
		return nil, fmt.Errorf("generated post is invalid: title or description is empty")
	}

	// Find flair ID by matching AI-selected flair text
	var selectedFlairID string
	if resp.SelectedFlair != "" {
		for _, flair := range flairs {
			if strings.EqualFold(strings.TrimSpace(flair.Text), strings.TrimSpace(resp.SelectedFlair)) {
				selectedFlairID = flair.ID
				break
			}
		}
	}

	// If flair is required but AI's flair not found, use the first available flair
	if postRequirement.IsFlairRequired && selectedFlairID == "" && len(flairs) > 0 {
		selectedFlairID = flairs[0].ID
		s.logger.Warn("AI flair not matched, using first available flair as default",
			zap.String("default_flair", flairs[0].Text))
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
	}

	newPost, err := s.db.CreatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return newPost, nil
}

func (s *postService) UpdatePost(ctx context.Context, updated *models.Post) (*models.Post, error) {
	existing, err := s.db.GetPostByID(ctx, updated.ID)
	if err != nil && !errors.Is(err, datastore.NotFound) {
		return nil, fmt.Errorf("invalid post id: %w", err)
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
