package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/dstore"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/zap"
	"strings"
	"text/template"
	"time"
)

const SEED = 42

type Client struct {
	model           openai.Client
	langsmithConfig LangsmithConfig
	/**/ debugFileStore dstore.Store
}

type LangsmithConfig struct {
	ProjectName string
	ApiKey      string
}

//func newClient(
//	model llms.Model,
//	langSmithConfig LangsmithConfig,
//	debugFileStore dstore.Store,
//) (*Client, error) {
//	if langSmithConfig.ProjectName == "" {
//		return nil, fmt.Errorf("project name is required")
//	}
//
//	if langSmithConfig.ApiKey == "" {
//		return nil, fmt.Errorf("langsmith api key is required")
//	}
//	return &Client{
//		model:           model,
//		langsmithConfig: langSmithConfig,
//		debugFileStore:  debugFileStore,
//	}, nil
//}

func NewOpenAI(apiKey, openAIOrganization string, config LangsmithConfig, debugFileStore dstore.Store) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai api key is required, cannot be blank")
	}
	//if openAIOrganization == "" {
	//	return nil, fmt.Errorf("openai organization is required, cannot be blank")
	//}

	llmClient := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://litellm.donebyai.team"),
	)

	return &Client{
		model:          llmClient,
		debugFileStore: debugFileStore,
	}, nil
}

func (c *Client) getChatMessages(ctx context.Context, runID string, templates []Template, vars map[string]any, logger *zap.Logger) ([]openai.ChatCompletionMessageParamUnion, *openai.ChatCompletionNewParamsResponseFormatUnion, error) {
	var chatPrompts []openai.ChatCompletionMessageParamUnion
	var responseSchemaTemplate openai.ChatCompletionNewParamsResponseFormatUnion
	for _, tmpl := range templates {
		data := rp(tmpl.path)
		buf := new(bytes.Buffer)
		err := template.Must(template.New(tmpl.path).Parse(data)).Execute(buf, vars)
		if err != nil {
			logger.Debug("failed to execute template", zap.Error(err), zap.String("tmpl", tmpl.path))
			return nil, nil, err
		}

		switch tmpl.promptType {
		case PromptTypeSYSTEM:
			chatPrompts = append(chatPrompts, openai.SystemMessage(buf.String()))
		case PromptTypeHUMAN:
			chatPrompts = append(chatPrompts, openai.UserMessage(buf.String()))
		case PromptTypeRESPONSESCHEMA:
			schema := shared.ResponseFormatJSONSchemaParam{}
			err := json.Unmarshal(buf.Bytes(), &schema)
			if err != nil {
				return nil, nil, nil
			}
			responseSchemaTemplate.OfJSONSchema = &schema
		}

		c.saveFile(ctx, runID, strings.TrimSuffix(tmpl.path, ".gotmpl"), buf, logger)
	}

	return chatPrompts, &responseSchemaTemplate, nil
}

func (c *Client) IsRedditPostRelevant(ctx context.Context, project *models.Project, post *models.Lead, gptModel GPTModel, logger *zap.Logger) (*models.RedditPostRelevanceResponse, error) {
	vars := gptModel.GetRedditPostRelevancyVars(project, post)

	runID := fmt.Sprintf("%s-%s", project.ID, post.PostID)
	messages, responseFormat, err := c.getChatMessages(ctx, runID, redditPostRelevancyTemplates, vars, logger)
	if err != nil {
		return nil, err
	}

	params := openai.ChatCompletionNewParams{
		Model:          gptModel.String(),
		Messages:       messages,
		User:           openai.String(project.OrganizationID),
		ResponseFormat: *responseFormat,
	}

	chatCompletion, err := c.model.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	if len(chatCompletion.Choices) == 0 {
		return nil, fmt.Errorf("llm: no chat completion found, model: %s", gptModel.String())
	}

	output := chatCompletion.Choices[0].Message.Content

	c.saveOutput(ctx, runID, "reddit_post_relevancy.output", []byte(output), logger)
	var data models.RedditPostRelevanceResponse
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response: %w", err)
	}

	return &data, nil
}

func (c *Client) CustomerCaseDecision(ctx context.Context, lastConversation *models.Conversation, gptModel GPTModel, logger *zap.Logger) (*models.CaseDecisionResponse, error) {
	vars := gptModel.GetCaseDecisionVars(lastConversation)

	runID := lastConversation.ID
	prompts, responseSchemaTemplate, debugTemplates := gptModel.getPromptTemplates(caseDecisionTemplates)
	c.debugTemplates(ctx, runID, vars, debugTemplates, logger)

	output, err := c.call(ctx, runID, prompts, responseSchemaTemplate, vars, gptModel, logger)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	c.saveOutput(ctx, runID, "classification.output", []byte(output), logger)
	var data models.CaseDecisionResponse
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return nil, fmt.Errorf("unable to unmarshal response: %w", err)
	}

	if data.CaseStatusReason == "" {
		return nil, fmt.Errorf("unable to make a case decision")
	}

	isValid := utils.Some([]models.CustomerCaseReason{
		models.CustomerCaseReasonPARTIALLYPAID,
		models.CustomerCaseReasonPAID,
		models.CustomerCaseReasonUNKNOWN,
		models.CustomerCaseReasonTALKTOSUPPORT,
		models.CustomerCaseReasonWILLPAYLATER,
		models.CustomerCaseReasonWILLNOTPAY,
	}, func(matchType models.CustomerCaseReason) bool {
		return matchType == data.CaseStatusReason
	})

	if !isValid {
		return nil, fmt.Errorf("invalid case status: %s, unable to make a case decision", data.CaseStatusReason)
	}

	if data.NextCallScheduledAt != "" {
		// try to parse it in a date
		t, err := time.Parse(time.RFC3339, data.NextCallScheduledAt)
		if err != nil {
			return nil, fmt.Errorf("unable to parse next call scheduled at %s: %w", data.NextCallScheduledAt, err)
		}
		data.NextCallScheduledAtTime = &t
	}

	return &data, nil
}

func (c *Client) RunPrompt(ctx context.Context, prefix string, prompt Prompt, vars map[string]any, runID string, logger *zap.Logger) ([]byte, error) {

	p, responseSchemaTemplate, debugTemplates := prompt.getPromptTemplate(prefix, false)

	c.debugTemplates(ctx, runID, vars, debugTemplates, logger)

	output, err := c.call(ctx, runID, p, responseSchemaTemplate, vars, prompt.Model, logger)
	if err != nil {
		return nil, fmt.Errorf("llm: %w", err)
	}

	c.saveOutput(ctx, runID, fmt.Sprintf("%s.output", prefix), []byte(output), logger)

	return []byte(output), nil
}

func (c *Client) ExtractMessages(ctx context.Context, prefix string, prompt Prompt, vars map[string]any, runID string, logger *zap.Logger) ([]llms.ChatMessage, error) {
	promptsTemplates, _, debugTemplates := prompt.getPromptTemplate(prefix, false)
	c.debugTemplates(ctx, runID, vars, debugTemplates, logger)
	chatMessages := []llms.ChatMessage{}
	for _, temp := range promptsTemplates.Messages {
		messages, err := temp.FormatMessages(vars)
		if err != nil {
			return nil, err
		}

		chatMessages = append(chatMessages, messages...)
	}

	return chatMessages, nil
}
