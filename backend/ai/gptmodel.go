package ai

import (
	"fmt"
	"github.com/shank318/doota/models"
	"text/template"
	"time"

	"github.com/tmc/langchaingo/prompts"
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
	path          string
	promptType    PromptType
	promptFeature PromptFeature
}

type ResponseFormat []byte

func (g *GPTModel) getPromptTemplates(templates []Template) (prompts.ChatPromptTemplate, *template.Template, []*template.Template) {
	var chatPrompts []prompts.MessageFormatter
	var tmpls []*template.Template
	var responseSchemaTemplate *template.Template
	for _, tmpl := range templates {
		data := rp(tmpl.path)
		switch tmpl.promptType {
		case PromptTypeSYSTEM:
			chatPrompts = append(chatPrompts, prompts.NewSystemMessagePromptTemplate(data, nil))
		case PromptTypeHUMAN:
			chatPrompts = append(chatPrompts, prompts.NewHumanMessagePromptTemplate(data, nil))
		case PromptTypeRESPONSESCHEMA:
			// If the model supports structured outputs
			responseSchemaTemplate = template.Must(template.New(tmpl.path).Parse(data))
		}

		tmpls = append(tmpls, template.Must(template.New(tmpl.path).Parse(data)))
	}
	return prompts.NewChatPromptTemplate(chatPrompts), responseSchemaTemplate, tmpls
}

func (g *GPTModel) getPromptTemplate(p *Prompt, templatePrefix string, addImageSupport bool) (prompts.ChatPromptTemplate, *template.Template, []*template.Template) {
	var chatPrompts []prompts.MessageFormatter
	var debugTemplates []*template.Template
	var responseSchemaTemplate *template.Template
	if p.PromptTmpl != "" {
		chatPrompts = append(chatPrompts, prompts.NewSystemMessagePromptTemplate(p.PromptTmpl, nil))
		debugTemplates = append(debugTemplates, template.Must(template.New(fmt.Sprintf("%s.prompt.gotmpl", templatePrefix)).Parse(p.PromptTmpl)))
	}

	if p.SchemaTmpl != "" {
		responseSchemaTemplate = template.Must(template.New(fmt.Sprintf("%s.schema.gotmpl", templatePrefix)).Parse(p.SchemaTmpl))
		debugTemplates = append(debugTemplates, template.Must(template.New(fmt.Sprintf("%s.schema.gotmpl", templatePrefix)).Parse(p.SchemaTmpl)))
	}
	if p.HumanTmpl != "" {
		chatPrompts = append(chatPrompts, prompts.NewHumanMessagePromptTemplate(p.HumanTmpl, nil))
		debugTemplates = append(debugTemplates, template.Must(template.New(fmt.Sprintf("%s.human.gotmpl", templatePrefix)).Parse(p.HumanTmpl)))

	}

	return prompts.NewChatPromptTemplate(chatPrompts), responseSchemaTemplate, debugTemplates
}
