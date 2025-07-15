package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
)

type PostService interface {
	CreatePost(ctx context.Context, post *models.Post, project *models.Project) (*models.Post, error)
	DeletePost(ctx context.Context, postID string) error
	SchedulePost(ctx context.Context, postID, version string, scheduleAt time.Time) error
}
type postService struct {
	aiClient *ai.Client
	db       datastore.Repository
	logger   *zap.Logger
}

func NewPostService(logger *zap.Logger, db datastore.Repository, aiClient *ai.Client) *postService {
	return &postService{logger: logger, db: db, aiClient: aiClient}
}

func (s *postService) CreatePost(ctx context.Context, post *models.Post, project *models.Project) (*models.Post, error) {
	_, err := s.db.GetSourceByID(ctx, post.SourceID)
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

	input := &ai.PostGenerateInput{
		Id:          post.ID,
		Project:     project,
		PostSetting: &post.Metadata.Settings,
	}

	resp, _, err := s.aiClient.GeneratePost(ctx, s.aiClient.GetDefaultModel(), input, s.logger)
	if err != nil {
		return nil, fmt.Errorf("generate post failed: %w", err)
	}

	if strings.TrimSpace(resp.Title) == "" || strings.TrimSpace(resp.Description) == "" {
		return nil, fmt.Errorf("generated post is invalid: title or description is empty")
	}

	settings := post.Metadata.Settings

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

func (s *postService) SchedulePost(ctx context.Context, postID, version string, scheduleAt time.Time) error {
	post, err := s.db.GetPostByID(ctx, postID)

	if err != nil && !errors.Is(err, datastore.NotFound) {
		return fmt.Errorf("invalid post id: %w", err)
	}

	if scheduleAt.Before(time.Now().Add(-15 * time.Second)) {
		return fmt.Errorf("cannot schedule post in the past")
	}

	// Parse version string like "v1"
	var versionIndex int
	if _, err := fmt.Sscanf(version, "v%d", &versionIndex); err != nil || versionIndex < 1 {
		return fmt.Errorf("invalid version format: %s", version)
	}

	history := post.Metadata.History
	if versionIndex > len(history) || versionIndex < 1 {
		return fmt.Errorf("version index out of bounds")
	}

	realIndex := versionIndex - 1
	selectedHistory := history[realIndex]

	// Update post with selected version data
	post.Title = selectedHistory.Title
	post.Description = selectedHistory.Description
	post.Metadata.Settings = selectedHistory.PostSettings

	post.ScheduleAt = &scheduleAt
	post.Status = models.PostStatusSCHEDULED

	if err := s.db.UpdatePost(ctx, post); err != nil {
		return err
	}

	return nil
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
