package pbcore

import (
	"github.com/shank318/doota/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (u *PostInsight) FromModel(lead *models.AugmentedPostInsight) *PostInsight {
	u.Id = lead.ID
	u.ProjectId = lead.ProjectID
	u.Topic = lead.Topic
	u.RelevancyScore = lead.RelevancyScore
	u.Sentiment = lead.Sentiment
	u.Highlights = lead.Highlights
	u.PostTitle = lead.Metadata.Title
	u.Cot = lead.Metadata.ChainOfThought
	u.Keyword = lead.Keyword.Keyword
	u.PostId = lead.PostID
	u.HighlightedComments = lead.Metadata.HighlightedComments

	u.CreatedAt = timestamppb.New(lead.CreatedAt)
	u.PostCreatedAt = timestamppb.New(lead.Metadata.PostCreatedAt)
	return u
}
