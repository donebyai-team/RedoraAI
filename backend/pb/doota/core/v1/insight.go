package pbcore

import (
	"github.com/shank318/doota/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (u *PostInsight) FromModel(lead *models.PostInsight) *PostInsight {
	u.Id = lead.ID
	u.ProjectId = lead.ProjectID
	u.Topic = lead.Topic
	u.RelevancyScore = lead.RelevancyScore
	u.Sentiment = lead.Sentiment
	u.Highlights = lead.Highlights
	u.Source = lead.Source.String()
	u.HighlightedComments = lead.Metadata.HighlightedComments
	u.PostTitle = lead.Metadata.Title
	u.Cot = lead.Metadata.ChainOfThought
	u.PostId = lead.PostID
	u.CreatedAt = timestamppb.New(lead.CreatedAt)
	u.PostCreatedAt = timestamppb.New(lead.Metadata.PostCreatedAt)
	return u
}
