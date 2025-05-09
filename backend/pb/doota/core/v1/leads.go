package pbcore

import (
	"fmt"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
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

func (u *Lead) FromModel(lead *models.AugmentedLead) *Lead {
	u.Id = lead.ID
	u.ProjectId = lead.ProjectID
	u.SourceId = lead.SourceID
	u.Keyword = new(Keyword).FromModel(lead.Keyword)
	u.Author = fmt.Sprintf("/u/%s", lead.Author)
	u.PostId = lead.PostID
	u.Type.FromModel(lead.Type)
	u.Status.FromModel(lead.Status)
	u.RelevancyScore = lead.RelevancyScore
	u.PostCreatedAt = timestamppb.New(lead.PostCreatedAt)
	u.CreatedAt = timestamppb.New(lead.CreatedAt)
	u.Title = lead.Title
	u.Description = lead.Description
	u.Metadata = new(LeadMetadata).FromModel(lead.LeadMetadata)
	categories := CategorizePost(lead.Intents)
	for _, intent := range categories {
		u.Intents = append(u.Intents, string(intent))
	}
	return u
}

func (u *Keyword) FromModel(lead *models.Keyword) *Keyword {
	u.Id = lead.ID
	u.Name = lead.Keyword
	return u
}

func (u *Project) FromModel(product *models.Project, sources []*models.Source, keywords []*models.Keyword) *Project {
	u.Id = product.ID
	u.Name = product.Name
	u.Description = product.ProductDescription
	u.Website = product.WebsiteURL
	u.TargetPersona = product.CustomerPersona

	sourcesProto := make([]*Source, 0, len(sources))
	for _, source := range sources {
		sourcesProto = append(sourcesProto, new(Source).FromModel(source, new(Source_RedditMetadata).FromModel(&source.Metadata)))
	}

	keywordsProto := make([]*Keyword, 0, len(keywords))
	for _, keyword := range keywords {
		keywordsProto = append(keywordsProto, new(Keyword).FromModel(keyword))
	}

	u.Keywords = keywordsProto
	u.Sources = sourcesProto
	u.SuggestedKeywords = []string{"SEO Agency", "AI SDR", "AI SEO"}
	u.SuggestedSources = []string{"r/saas", "r/sales", "r/marketing"}
	return u
}

func (u *LeadMetadata) FromModel(metadata models.LeadMetadata) *LeadMetadata {
	u.ChainOfThought = utils.FormatComment(metadata.ChainOfThought)
	u.SuggestedComment = utils.FormatComment(metadata.SuggestedComment)
	u.SuggestedDm = utils.FormatDM(metadata.SuggestedDM)
	u.ChainOfThoughtSuggestedComment = utils.FormatComment(metadata.ChainOfThoughtSuggestedComment)
	u.ChainOfThoughtSuggestedDm = utils.FormatComment(metadata.ChainOfThoughtSuggestedDM)
	u.PostUrl = metadata.PostURL
	u.NoOfComments = metadata.NoOfComments
	u.Ups = metadata.Ups
	u.AuthorUrl = metadata.AuthorURL
	u.DmUrl = metadata.DmURL
	u.SubredditPrefixed = metadata.SubRedditPrefixed
	u.DescriptionHtml = metadata.SelfTextHTML
	u.AutomatedCommentUrl = metadata.AutomatedCommentURL
	u.RelevancyLlmModel = string(metadata.RelevancyLLMModel)
	u.DmLlmModel = string(metadata.DMLLMModel)
	u.CommentLlmModel = string(metadata.CommentLLMModel)
	u.LlmModelResponseOverriddenBy = string(metadata.LLMModelResponseOverriddenBy)
	return u
}

func (u LeadStatus) ToModel() models.LeadStatus {
	model := models.LeadStatus(strings.ToUpper(u.String()))
	if !model.IsValid() {
		panic(fmt.Errorf("unknown lead status type pb %q", u.String()))
	}

	return model
}
