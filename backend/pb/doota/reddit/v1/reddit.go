package pbreddit

import (
	"fmt"
	"github.com/shank318/doota/models"
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
)

func (r *LeadStatus) FromModel(status models.LeadStatus) {
	enum, found := LeadStatus_value[strings.ToUpper(string(status))]
	if !found {
		panic(fmt.Errorf("unknown lead status %q", status))
	}
	*r = LeadStatus(enum)
}

func (r *LeadType) FromModel(status models.LeadType) {
	enum, found := LeadType_value[strings.ToUpper(string(status))]
	if !found {
		panic(fmt.Errorf("unknown lead type %q", status))
	}
	*r = LeadType(enum)
}

func (u *RedditLead) FromModel(lead *models.RedditLead) *RedditLead {
	u.Id = lead.ID
	u.ProjectId = lead.ProjectID
	u.SubredditId = lead.SubRedditID
	u.Author = lead.Author
	u.PostId = lead.PostID
	u.Type.FromModel(lead.Type)
	u.Status.FromModel(lead.Status)
	u.RelevancyScore = lead.RelevancyScore
	u.PostCreatedAt = timestamppb.New(lead.PostCreatedAt)
	u.CreatedAt = timestamppb.New(lead.CreatedAt)
	u.Title = lead.Title
	u.Description = lead.Description
	u.Metadata = new(LeadMetadata).FromModel(lead.LeadMetadata)
	return u
}

func (u *LeadMetadata) FromModel(metadata models.LeadMetadata) *LeadMetadata {
	u.ChainOfThought = metadata.ChainOfThought
	u.SuggestedComment = metadata.SuggestedComment
	u.SuggestedDm = metadata.SuggestedDM
	u.ChainOfThoughtSuggestedComment = metadata.ChainOfThoughtSuggestedComment
	u.ChainOfThoughtCommentSuggestedDm = metadata.ChainOfThoughtCommentSuggestedDM
	u.PostUrl = metadata.PostURL
	return u
}
