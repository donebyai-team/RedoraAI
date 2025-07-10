package services

import (
	"context"
	"fmt"

	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
)

type PostService interface {
	CreatePost(ctx context.Context, post *models.Post, project *models.Project) (*models.Post, error)
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
		Project:     project,
		PostSetting: &post.Metadata.Settings,
	}

	resp, _, err := s.aiClient.GeneratePost(ctx, s.aiClient.GetAdvanceModel(), input, s.logger)
	if err != nil {
		return nil, fmt.Errorf("generate post failed: %w", err)
	}

	// resp := struct {
	// 	Title       string
	// 	Description string
	// }{
	// 	Title:       "My First Reddit Post",
	// 	Description: "This is a scheduled post using inline struct.",
	// }

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
		existingPost.Status = models.PostStatusPROCESSING
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

// generateAIContent simulates an OpenAI API call
// func generateAIContent(ctx context.Context, settings models.PostSettings) (title, description string, err error) {
// 	// Simulated content generation logic. Replace with actual API call.
// 	content := fmt.Sprintf(
// 		"Title for topic: %s in tone: %s\nDescription: This is a simulated AI-generated post about '%s' for the goal '%s'.",
// 		settings.Topic, settings.Tone, settings.Context, settings.Goal,
// 	)

// 	// Assume that the AI API could return JSON-encoded structured response
// 	title = fmt.Sprintf("AI Generated: %s", settings.Topic)
// 	descriptionBytes, err := json.Marshal(map[string]string{
// 		"topic":   settings.Topic,
// 		"context": settings.Context,
// 		"goal":    settings.Goal,
// 		"tone":    settings.Tone,
// 		"content": content,
// 	})
// 	if err != nil {
// 		return "", "", fmt.Errorf("failed to marshal description: %w", err)
// 	}

// 	return title, string(descriptionBytes), nil
// }
