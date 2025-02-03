package ai

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/streamingfast/derr"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/langsmith"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/prompts"
	"go.uber.org/zap"
)

const MAX_RETRIES = 3

func (c *Client) llmChain(template prompts.FormatPrompter) *chains.LLMChain {
	return chains.NewLLMChain(c.model, template)
}

func (c *Client) langsmithTracer(runId string, logger *zap.Logger) (*langsmith.LangChainTracer, error) {
	client, err := langsmith.NewClient(langsmith.WithAPIKey(c.langsmithConfig.ApiKey))
	if err != nil {
		return nil, fmt.Errorf("new langsmith client: %w", err)
	}

	langChainTracer, err := langsmith.NewTracer(
		langsmith.WithLogger(&aiLogger{Logger: logger}),
		langsmith.WithProjectName(c.langsmithConfig.ProjectName),
		langsmith.WithClient(client),
		langsmith.WithRunID(runId),
	)
	if err != nil {
		return nil, fmt.Errorf("chain tracer: %w", err)
	}
	return langChainTracer, nil
}

func (c *Client) call(ctx context.Context, runId string, template prompts.FormatPrompter, responseSchemaTemplate *template.Template, values map[string]any, gptModel GPTModel, logger *zap.Logger) (string, error) {
	var output map[string]any

	llmChain := c.llmChain(template)
	llmChain.EnableMultiPrompt()
	langsmithTracer, err := c.langsmithTracer(runId, logger)
	if err != nil {
		return "", fmt.Errorf("langsmith tracer: %w", err)
	}
	langsmithTracer.GetRunID()

	err = derr.RetryContext(ctx, MAX_RETRIES, func(ctx context.Context) error {
		options := []chains.ChainCallOption{
			chains.WithCallback(langsmithTracer),
			chains.WithModel(gptModel.String()),
			chains.WithTemperature(0.1),
			chains.WithSeed(SEED),
		}
		if responseSchemaTemplate != nil {
			responseSchema := new(bytes.Buffer)
			err = responseSchemaTemplate.Execute(responseSchema, values)
			if err != nil {
				return fmt.Errorf("failed to execute response schema template[%s]: %v", err, responseSchemaTemplate.Name())
			}
			options = append(options, chains.WithJSONFormat(responseSchema.String()))
		} else {
			options = append(options, chains.WithJSONMode())
		}

		output, err = chains.Call(ctx, llmChain, values, options...)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("retry: %w", err)
	}

	contentRaw, found := output["choices"]
	if !found {
		return "", nil
	}

	choices, ok := contentRaw.([]*llms.ContentChoice)
	if !ok {
		return "", fmt.Errorf("content raw is not a list of choices")
	}

	return choices[0].Content, nil
}
