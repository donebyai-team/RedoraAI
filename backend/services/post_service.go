package services

import (
	"context"
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
	SchedulePost(ctx context.Context, postID string, scheduleAt time.Time, projectID string) error
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
		// Update existing post
		historyEntry := models.PostRegenerationHistory{
			PostSettings: existingPost.Metadata.Settings,
			Title:        existingPost.Title,
			Description:  existingPost.Description,
		}

		existingPost.Metadata.History = append(existingPost.Metadata.History, historyEntry)
		existingPost.Title = resp.Title
		existingPost.Description = resp.Description
		existingPost.ReferenceID = settings.ReferenceID

		existingPost.Metadata.Settings = models.PostSettings{
			Topic:       settings.Topic,
			Context:     settings.Context,
			Goal:        settings.Goal,
			Tone:        settings.Tone,
			ReferenceID: settings.ReferenceID,
		}

		if err := s.db.UpdatePost(ctx, existingPost); err != nil {
			return nil, fmt.Errorf("failed to update post: %w", err)
		}
		return existingPost, nil
	}

	//Create new post
	post.Title = resp.Title
	post.Description = resp.Description
	post.Status = models.PostStatusCREATED

	post.Metadata = models.PostMetadata{
		Settings: settings,
		History:  []models.PostRegenerationHistory{},
	}

	newPost, err := s.db.CreatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return newPost, nil
}

func (s *postService) SchedulePost(ctx context.Context, postID string, scheduleAt time.Time, expectedProjectID string) error {
	post, err := s.db.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to fetch post: %w", err)
	}

	if post.ProjectID != expectedProjectID {
		return fmt.Errorf("you do not have access to this post")
	}

	if post.Status == models.PostStatusSCHEDULED {
		return fmt.Errorf("post is already scheduled")
	}

	if scheduleAt.Before(time.Now().Add(-15 * time.Second)) {
		return fmt.Errorf("cannot schedule post in the past")
	}

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
