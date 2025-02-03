package ai

import (
	"fmt"
	"github.com/shank318/doota/models"
	"slices"
	"text/template"

	"github.com/tmc/langchaingo/prompts"
)

//go:generate go-enum -f=$GOFILE

var knownTemplates = []Template{
	{path: "classification.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureTEXTONLY},
	{path: "classification_vision.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureIMAGEONLY},
	{path: "classification.schema.gotmpl", promptType: PromptTypeResponseSchema, promptFeature: PromptFeatureBOTH},
	{path: "human.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureBOTH},
}

// ENUM(gpt-4-vision-preview, gpt-4-turbo, gpt-4-turbo-preview, gpt-4-0125-preview, gpt-4-turbo-2024-04-09, gpt-4o-2024-05-13, gpt-4o-2024-08-06)
type GPTModel string

func (g GPTModel) GetVars(customerCase *models.AugmentedCustomerCase) Variable {
	out := make(Variable).
		WithCustomer(customerCase.Customer).
		WithCustomerCase(customerCase.CustomerCase)
	if len(customerCase.Conversations) > 0 {
		out = out.WithPastConversations(customerCase.Conversations)
	}
	return out
}

// TODO: Create list of all gptModels that support images
var visionGPTModelList = []GPTModel{
	"gpt-4-turbo",
	"gpt-4-vision-preview",
	"gpt-4o-2024-05-13",
	"gpt-4o-2024-08-06",
}

var structuredOutputGPTModelList = []GPTModel{
	"gpt-4o-2024-08-06",
}

func (g GPTModel) SupportsImage() bool {
	return slices.Contains(visionGPTModelList, g)
}

func (g GPTModel) SupportsStructuredOutput() bool {
	return slices.Contains(structuredOutputGPTModelList, g)
}

// ENUM(HUMAN,SYSTEM,IMAGE)
type PromptType string

// ENUM(IMAGE_ONLY,TEXT_ONLY,BOTH)
type PromptFeature string

type Template struct {
	path          string
	promptType    PromptType
	promptFeature PromptFeature
}

type ResponseFormat []byte

func (g *GPTModel) getPromptTemplate(p *Prompt, templatePrefix string, addImageSupport bool) (prompts.ChatPromptTemplate, *template.Template, []*template.Template) {
	var chatPrompts []prompts.MessageFormatter
	var debugTemplates []*template.Template
	var responseSchemaTemplate *template.Template
	if p.PromptTmpl != "" {
		chatPrompts = append(chatPrompts, prompts.NewSystemMessagePromptTemplate(p.PromptTmpl, nil))
		debugTemplates = append(debugTemplates, template.Must(template.New(fmt.Sprintf("%s.prompt.gotmpl", templatePrefix)).Parse(p.PromptTmpl)))
	}

	if p.SchemaTmpl != "" {
		// Return responseSchemaTemplate only when the model supports structured outputs
		if g.SupportsStructuredOutput() {
			responseSchemaTemplate = template.Must(template.New(fmt.Sprintf("%s.schema.gotmpl", templatePrefix)).Parse(p.SchemaTmpl))
		} else {
			chatPrompts = append(chatPrompts, prompts.NewSystemMessagePromptTemplate(p.PromptTmpl, nil))
		}
		debugTemplates = append(debugTemplates, template.Must(template.New(fmt.Sprintf("%s.schema.gotmpl", templatePrefix)).Parse(p.SchemaTmpl)))
	}
	if p.HumanTmpl != "" {
		chatPrompts = append(chatPrompts, prompts.NewHumanMessagePromptTemplate(p.HumanTmpl, nil))
		debugTemplates = append(debugTemplates, template.Must(template.New(fmt.Sprintf("%s.human.gotmpl", templatePrefix)).Parse(p.HumanTmpl)))

	}

	if g.SupportsImage() && addImageSupport {
		chatPrompts = append(chatPrompts, prompts.MessagesPlaceholder{VariableName: "Images"})
	}

	return prompts.NewChatPromptTemplate(chatPrompts), responseSchemaTemplate, debugTemplates
}
