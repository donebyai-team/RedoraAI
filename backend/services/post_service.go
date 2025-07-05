package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	pbcore "github.com/shank318/doota/pb/doota/core/v1"
	"go.uber.org/zap"
)

type PostService interface {
	CreatePost(ctx context.Context, post *models.Post, settings *pbcore.PostSettings) (*models.Post, error)
}
type postService struct {
	aiClient *ai.Client
	db       datastore.Repository
	logger   *zap.Logger
}

func NewPostService(logger *zap.Logger, db datastore.Repository, aiClient *ai.Client) *postService {
	return &postService{logger: logger, db: db, aiClient: aiClient}
}

func (s *postService) CreatePost(ctx context.Context, post *models.Post, settings *pbcore.PostSettings) (*models.Post, error) {
	var existingPost *models.Post
	var err error

	if settings.Id != nil {
		// Only query if ID is present
		existingPost, err = s.db.GetPostByID(ctx, *settings.Id)
		if err != nil && !errors.Is(err, datastore.NotFound) {
			return nil, fmt.Errorf("failed to check existing post: %w", err)
		}
	}

	// Generate new AI content
	title, description, err := generateAIContent(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("AI generation failed: %w", err)
	}

	if existingPost != nil {
		// Marshal previous title & description into JSON
		historyTextBytes, err := json.Marshal(map[string]string{
			"title":       existingPost.Title,
			"description": existingPost.Description,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal history text: %w", err)
		}

		// Create history entry
		historyEntry := models.PostRegenerationHistory{
			PostSettings: existingPost.Metadata.Settings,
			Text:         string(historyTextBytes),
		}

		// Append to history and update post
		existingPost.Metadata.History = append(existingPost.Metadata.History, historyEntry)
		existingPost.Title = title
		existingPost.Description = description
		existingPost.Status = models.PostStatusPROCESSING
		existingPost.ReferenceID = settings.ReferenceId

		existingPost.Metadata.Settings = models.PostSettings{
			Topic:       settings.Topic,
			Context:     settings.Context,
			Goal:        settings.Goal,
			Tone:        settings.Tone,
			ReferenceID: settings.ReferenceId,
		}

		if err := s.db.UpdatePost(ctx, existingPost); err != nil {
			return nil, fmt.Errorf("failed to update post: %w", err)
		}
		return existingPost, nil
	}

	// New post path
	post.Title = title
	post.Description = description
	post.Status = models.PostStatusCREATED

	historyTextBytes, err := json.Marshal(map[string]string{
		"title":       title,
		"description": description,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal history text: %w", err)
	}

	historyEntry := models.PostRegenerationHistory{
		PostSettings: models.PostSettings{
			Topic:       settings.Topic,
			Context:     settings.Context,
			Goal:        settings.Goal,
			Tone:        settings.Tone,
			ReferenceID: settings.ReferenceId,
		},
		Text: string(historyTextBytes),
	}

	post.Metadata = models.PostMetadata{
		Settings: historyEntry.PostSettings,
		History:  []models.PostRegenerationHistory{
			//historyEntry
		},
	}

	newPost, err := s.db.CreatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	return newPost, nil
}

// generateAIContent simulates an OpenAI API call
func generateAIContent(ctx context.Context, settings *pbcore.PostSettings) (title, description string, err error) {
	// Simulated content generation logic. Replace with actual API call.
	content := fmt.Sprintf(
		"Title for topic: %s in tone: %s\nDescription: This is a simulated AI-generated post about '%s' for the goal '%s'.",
		settings.Topic, settings.Tone, settings.Context, settings.Goal,
	)

	// Assume that the AI API could return JSON-encoded structured response
	title = fmt.Sprintf("AI Generated: %s", settings.Topic)
	descriptionBytes, err := json.Marshal(map[string]string{
		"topic":   settings.Topic,
		"context": settings.Context,
		"goal":    settings.Goal,
		"tone":    settings.Tone,
		"content": content,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal description: %w", err)
	}

	return title, string(descriptionBytes), nil
}
