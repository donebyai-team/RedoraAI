package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/dstore"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
	"go.uber.org/zap"
	"time"
)

const SEED = 42

type Client struct {
	model           llms.Model
	langsmithConfig LangsmithConfig
	/**/ debugFileStore dstore.Store
}

type LangsmithConfig struct {
	ProjectName string
	ApiKey      string
}

func newClient(
	model llms.Model,
	langSmithConfig LangsmithConfig,
	debugFileStore dstore.Store,
) (*Client, error) {
	if langSmithConfig.ProjectName == "" {
		return nil, fmt.Errorf("project name is required")
	}

	if langSmithConfig.ApiKey == "" {
		return nil, fmt.Errorf("langsmith api key is required")
	}
	return &Client{
		model:           model,
		langsmithConfig: langSmithConfig,
		debugFileStore:  debugFileStore,
	}, nil
}

func NewOpenAI(apiKey, openAIOrganization string, config LangsmithConfig, debugFileStore dstore.Store) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai api key is required, cannot be blank")
	}
	if openAIOrganization == "" {
		return nil, fmt.Errorf("openai organization is required, cannot be blank")
	}

	model, err := openai.New(
		openai.WithToken(apiKey),
		openai.WithOrganization(openAIOrganization),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create openai model: %w", err)
	}

	return newClient(model, config, debugFileStore)
}

func (c *Client) CustomerCaseDecision(ctx context.Context, lastConversation *models.Conversation, gptModel GPTModel, logger *zap.Logger) (*models.CaseDecisionResponse, error) {
	vars := gptModel.GetCaseDecisionVars(lastConversation)

	runID := lastConversation.ID
	prompts, responseSchemaTemplate, debugTemplates := gptModel.getCaseDecisionTemplates()
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
