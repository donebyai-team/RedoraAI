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
	"github.com/streamingfast/derr"
	"github.com/streamingfast/dstore"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const SEED = 42

type Client struct {
	defaultLLMModel     models.LLMModel
	model               openai.Client
	langsmithConfig     LangsmithConfig
	/**/ debugFileStore dstore.Store
	log                 *zap.Logger
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

func NewOpenAI(apiKey string, defaultLLMModel models.LLMModel, config LangsmithConfig, debugFileStore dstore.Store, log *zap.Logger) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("openai api key is required, cannot be blank")
	}

	if len(strings.TrimSpace(string(defaultLLMModel))) == 0 {
		return nil, fmt.Errorf("default llm model, cannot be blank")
	}

	//if openAIOrganization == "" {
	//	return nil, fmt.Errorf("openai organization is required, cannot be blank")
	//}

	llmClient := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL("https://litellm.donebyai.team"),
	)

	return &Client{
		model:           llmClient,
		defaultLLMModel: defaultLLMModel,
		debugFileStore:  debugFileStore,
		log:             log,
	}, nil
}

func (c *Client) processTemplate(ctx context.Context, runID, path, tmplData string, vars map[string]any, logger *zap.Logger) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	err := template.Must(template.New(path).Parse(tmplData)).Execute(buf, vars)
	if err != nil {
		logger.Error("failed to execute template", zap.Error(err), zap.String("tmpl", path))
		return nil, err
	}
	return buf, nil
}

func (c *Client) buildChatMessages(ctx context.Context, runID string, templates []Template, logger *zap.Logger, vars map[string]any) ([]openai.ChatCompletionMessageParamUnion, *openai.ChatCompletionNewParamsResponseFormatUnion, error) {
	var chatPrompts []openai.ChatCompletionMessageParamUnion
	var responseSchemaTemplate openai.ChatCompletionNewParamsResponseFormatUnion

	for _, tmpl := range templates {
		tempData := tmpl.content
		if tempData == "" {
			tempData = rp(tmpl.path)
		}
		buf, err := c.processTemplate(ctx, runID, tmpl.path, tempData, vars, logger)
		if err != nil {
			return nil, nil, err
		}

		switch tmpl.promptType {
		case PromptTypeSYSTEM:
			chatPrompts = append(chatPrompts, openai.SystemMessage(buf.String()))
		case PromptTypeHUMAN:
			chatPrompts = append(chatPrompts, openai.UserMessage(buf.String()))
		case PromptTypeRESPONSESCHEMA:
			var schema shared.ResponseFormatJSONSchemaParam
			if err := json.Unmarshal(buf.Bytes(), &schema); err != nil {
				return nil, nil, err
			}
			responseSchemaTemplate.OfJSONSchema = &schema
		}

		// NOTE: Make sure you have the buff after processing
		c.saveFile(ctx, runID, strings.TrimSuffix(tmpl.path, ".gotmpl"), buf, logger)
	}

	return chatPrompts, &responseSchemaTemplate, nil
}

func (c *Client) getChatMessagesFromPrompt(ctx context.Context, runID string, p *Prompt, prefix string, vars map[string]any, logger *zap.Logger) ([]openai.ChatCompletionMessageParamUnion, *openai.ChatCompletionNewParamsResponseFormatUnion, error) {
	var templates []Template

	if p.PromptTmpl != "" {
		templates = append(templates, Template{path: fmt.Sprintf("%s.prompt.gotmpl", prefix), promptType: PromptTypeSYSTEM, content: p.PromptTmpl})
	}
	if p.HumanTmpl != "" {
		templates = append(templates, Template{path: fmt.Sprintf("%s.human.gotmpl", prefix), promptType: PromptTypeHUMAN, content: p.HumanTmpl})
	}
	if p.SchemaTmpl != "" {
		templates = append(templates, Template{path: fmt.Sprintf("%s.schema.gotmpl", prefix), promptType: PromptTypeRESPONSESCHEMA, content: p.SchemaTmpl})
	}

	return c.buildChatMessages(ctx, runID, templates, logger, vars)
}

const MAX_RETRIES = 3

func (c *Client) runChatCompletion(
	ctx context.Context,
	runID string,
	model models.LLMModel,
	userID string,
	messages []openai.ChatCompletionMessageParamUnion,
	responseFormat *openai.ChatCompletionNewParamsResponseFormatUnion,
	logger *zap.Logger,
	outputFile string,
) ([]byte, error) {
	var output string

	err := derr.RetryContext(ctx, MAX_RETRIES, func(ctx context.Context) error {
		httpResponse := &http.Response{}
		params := openai.ChatCompletionNewParams{
			Model:          string(model),
			Messages:       messages,
			User:           openai.String(userID),
			ResponseFormat: *responseFormat,
		}

		// Send the API request
		chatCompletion, err := c.model.Chat.Completions.New(ctx, params, option.WithResponseInto(&httpResponse))

		// Check if we hit rate limit even if there's an error
		if httpResponse.StatusCode == http.StatusTooManyRequests {
			c.log.Debug("rate limit hit (429), retrying after backoff", zap.String("model", string(model)))
			retryAfter := httpResponse.Header.Get("Retry-After")
			if retryAfter != "" {
				if seconds, convErr := strconv.Atoi(retryAfter); convErr == nil {
					time.Sleep(time.Duration(seconds) * time.Second)
				} else {
					time.Sleep(time.Minute)
				}
			} else {
				time.Sleep(time.Minute)
			}
			return fmt.Errorf("[%s]rate limit hit (429), retrying after backoff", model)
		}

		if err != nil {
			return fmt.Errorf("llm: %w", err)
		}

		// Proceed normally
		if len(chatCompletion.Choices) == 0 {
			return fmt.Errorf("llm: no chat completion found, model: %s", model)
		}

		// Optionally: check remaining rate limits
		remainingRequests := httpResponse.Header.Get("x-ratelimit-remaining-requests")
		remainingTokens := httpResponse.Header.Get("x-ratelimit-remaining-tokens")

		reqRemaining, _ := strconv.Atoi(remainingRequests)
		tokRemaining, _ := strconv.Atoi(remainingTokens)

		if reqRemaining <= 1 || tokRemaining <= 2000 {
			c.log.Debug("rate limit hit (remaining requests/tokens), waiting for 1min before next request", zap.String("model", string(model)))
			time.Sleep(time.Minute)
		}

		// If rate limit not exhausted, proceed normally
		output = chatCompletion.Choices[0].Message.Content
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("retry: %w", err)
	}

	// Save output to the specified file
	c.saveOutput(ctx, runID, outputFile, []byte(output), logger)

	return []byte(output), nil
}

func (c *Client) IsRedditPostRelevant(ctx context.Context, organization *models.Organization, project *models.Project, post *models.Lead, logger *zap.Logger) (*models.RedditPostRelevanceResponse, *models.LLMModelUsage, error) {
	runID := fmt.Sprintf("%s-%s", project.ID, post.PostID)
	vars := GetRedditPostRelevancyVars(project, post)
	llmModelToUse := c.defaultLLMModel
	if organization.FeatureFlags.RelevancyLLMModel != "" {
		llmModelToUse = organization.FeatureFlags.RelevancyLLMModel
	}

	messages, responseFormat, err := c.buildChatMessages(ctx, runID, redditPostRelevancyTemplates, logger, vars)
	if err != nil {
		return nil, nil, err
	}

	output, err := c.runChatCompletion(
		ctx,
		runID,
		llmModelToUse,
		project.OrganizationID,
		messages,
		responseFormat,
		logger,
		"reddit_post_relevancy.output",
	)
	if err != nil {
		return nil, nil, err
	}

	var data models.RedditPostRelevanceResponse
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, nil, fmt.Errorf("unable to unmarshal response: %w", err)
	}
	return &data, &models.LLMModelUsage{Model: llmModelToUse}, nil
}

func (c *Client) CustomerCaseDecision(ctx context.Context, orgID string, lastConversation *models.Conversation, logger *zap.Logger) (*models.CaseDecisionResponse, error) {
	runID := lastConversation.ID
	vars := GetCaseDecisionVars(lastConversation)

	messages, responseFormat, err := c.buildChatMessages(ctx, runID, caseDecisionTemplates, logger, vars)
	if err != nil {
		return nil, err
	}

	output, err := c.runChatCompletion(
		ctx,
		runID,
		c.defaultLLMModel,
		orgID,
		messages,
		responseFormat,
		logger,
		"reddit_post_relevancy.output",
	)
	if err != nil {
		return nil, err
	}

	var data models.CaseDecisionResponse
	if err := json.Unmarshal(output, &data); err != nil {
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
	}, func(r models.CustomerCaseReason) bool {
		return r == data.CaseStatusReason
	})

	if !isValid {
		return nil, fmt.Errorf("invalid case status: %s, unable to make a case decision", data.CaseStatusReason)
	}

	if data.NextCallScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, data.NextCallScheduledAt)
		if err != nil {
			return nil, fmt.Errorf("unable to parse next call scheduled at %s: %w", data.NextCallScheduledAt, err)
		}
		data.NextCallScheduledAtTime = &t
	}

	return &data, nil
}

func (c *Client) RunPrompt(ctx context.Context, prefix string, prompt *Prompt, vars map[string]any, runID string, orgID string, logger *zap.Logger) ([]byte, error) {
	messages, responseFormat, err := c.getChatMessagesFromPrompt(ctx, runID, prompt, prefix, vars, logger)
	if err != nil {
		return nil, err
	}

	llmModelToUse := c.defaultLLMModel
	if prompt.Model != "" {
		llmModelToUse = prompt.Model
	}

	output, err := c.runChatCompletion(
		ctx,
		runID,
		llmModelToUse,
		orgID,
		messages,
		responseFormat,
		logger,
		fmt.Sprintf("%s.output", prefix),
	)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (c *Client) ExtractMessages(ctx context.Context, prefix string, prompt Prompt, vars map[string]any, runID string, logger *zap.Logger) ([]openai.ChatCompletionMessageParamUnion, error) {
	messages, _, err := c.getChatMessagesFromPrompt(ctx, runID, &prompt, prefix, vars, logger)
	if err != nil {
		return nil, err
	}
	return messages, nil
}
