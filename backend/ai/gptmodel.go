package ai

import (
	"github.com/shank318/doota/models"
	"time"
)

//go:generate go-enum -f=$GOFILE

var caseDecisionTemplates = []Template{
	{path: "case_decision.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureTEXTONLY},
	{path: "case_decision.schema.gotmpl", promptType: PromptTypeRESPONSESCHEMA, promptFeature: PromptFeatureBOTH},
	{path: "case_decision.human.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureBOTH},
}

var redditPostRelevancyTemplates = []Template{
	{path: "reddit_post.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureTEXTONLY},
	{path: "reddit_post.schema.gotmpl", promptType: PromptTypeRESPONSESCHEMA, promptFeature: PromptFeatureBOTH},
	{path: "reddit_post.human.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureBOTH},
}

// ENUM(redora-dev-gpt-4o-2024-08-06, redora-prod-gpt-4o-2024-08-06, gpt-4o-2024-08-06)
type GPTModel string

func (g GPTModel) GetVars(customerCase *models.AugmentedCustomerCase, currentTime time.Time) Variable {
	out := make(Variable).
		WithCustomer(customerCase.Customer).
		WithCustomerCase(customerCase.CustomerCase).
		WithPastConversations(customerCase.Conversations).
		WithConversationDate(currentTime)
	return out
}

func (g GPTModel) GetCaseDecisionVars(customerCase *models.Conversation) Variable {
	out := make(Variable).
		WithConversationDate(customerCase.CreatedAt).
		WithCallMessages(customerCase.CallMessages)
	return out
}

func (g GPTModel) GetRedditPostRelevancyVars(project *models.Project, post *models.Lead) Variable {
	out := make(Variable).
		WithProjectDetails(project).
		WithRedditPost(post)
	return out
}

// ENUM(HUMAN,SYSTEM,IMAGE,RESPONSE_SCHEMA)
type PromptType string

// ENUM(IMAGE_ONLY,TEXT_ONLY,BOTH)
type PromptFeature string

type Template struct {
	content       string
	path          string
	promptType    PromptType
	promptFeature PromptFeature
}
