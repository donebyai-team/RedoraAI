package pbcore

import (
	"github.com/shank318/doota/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (p *Post) FromModel(post *models.Post) *Post {
	history := []*PostRegenerationHistory{}
	for _, h := range post.Metadata.History {
		history = append(history, &PostRegenerationHistory{
			Text: h.Text,
			PostSettings: &PostSettings{
				Topic:       h.PostSettings.Topic,
				Context:     h.PostSettings.Context,
				Goal:        h.PostSettings.Goal,
				Tone:        h.PostSettings.Tone,
				ReferenceId: h.PostSettings.ReferenceID,
			},
		})
	}

	p.Id = post.ID
	p.ProjectId = post.ProjectID
	p.Source = post.SourceID
	p.Topic = post.Title
	p.Description = post.Description
	p.Status = string(post.Status)
	p.Reason = post.Reason
	p.CreatedAt = timestamppb.New(post.CreatedAt)

	if post.ScheduleAt != nil {
		p.ScheduledAt = timestamppb.New(*post.ScheduleAt)
	}

	p.Metadata = &PostMetadata{
		Settings: &PostSettings{
			Topic:       post.Metadata.Settings.Topic,
			Context:     post.Metadata.Settings.Context,
			Goal:        post.Metadata.Settings.Goal,
			Tone:        post.Metadata.Settings.Tone,
			ReferenceId: post.Metadata.Settings.ReferenceID,
		},
		History: history,
	}

	return p
}
