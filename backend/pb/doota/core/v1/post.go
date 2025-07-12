package pbcore

import (
	"time"

	"github.com/shank318/doota/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FromModel converts a models.Post into a protobuf Post.
func (p *Post) FromModel(post *models.Post) *Post {
	return &Post{
		Id:          post.ID,
		ProjectId:   post.ProjectID,
		Source:      post.SourceID,
		Topic:       post.Title,
		Description: post.Description,
		Status:      string(post.Status),
		Reason:      post.Reason,
		CreatedAt:   timestamppb.New(post.CreatedAt),
		ScheduledAt: toTimestamp(post.ScheduleAt),
		Metadata:    new(PostMetadata).FromModel(post.Metadata),
	}
}

// FromModel converts a models.PostSettings into a protobuf PostSettings.
func (ps *PostSettings) FromModel(m models.PostSettings) *PostSettings {
	return &PostSettings{
		Topic:       m.Topic,
		Context:     m.Context,
		Goal:        m.Goal,
		Tone:        m.Tone,
		ReferenceId: m.ReferenceID,
	}
}

// FromModel converts a models.PostMetadata into a protobuf PostMetadata.
func (pm *PostMetadata) FromModel(m models.PostMetadata) *PostMetadata {
	var history []*PostRegenerationHistory
	for _, h := range m.History {
		history = append(history, new(PostRegenerationHistory).FromModel(h))
	}
	return &PostMetadata{
		Settings: new(PostSettings).FromModel(m.Settings),
		History:  history,
	}
}

// FromModel converts a models.PostRegenerationHistory into a protobuf PostRegenerationHistory.
func (prh *PostRegenerationHistory) FromModel(m models.PostRegenerationHistory) *PostRegenerationHistory {
	return &PostRegenerationHistory{
		PostSettings: new(PostSettings).FromModel(m.PostSettings),
		Title:        m.Title,
		Description:  m.Description,
	}
}

// Helper function to convert *time.Time to *timestamppb.Timestamp
func toTimestamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func (p *Post) FromAugmentedModel(post *models.AugmentedPost) *Post {
	return &Post{
		Id:          post.ID,
		ProjectId:   post.ProjectID,
		Topic:       post.Title,
		Description: post.Description,
		Source:      post.SourceID,
		Status:      string(post.Status),
		Reason:      post.Reason,
		CreatedAt:   timestamppb.New(post.CreatedAt),
		ScheduledAt: toTimestamp(post.ScheduleAt),
		Metadata:    new(PostMetadata).FromModel(post.Metadata),
	}
}
