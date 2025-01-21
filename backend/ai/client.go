package ai

import (
	"fmt"
	"github.com/streamingfast/dstore"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
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
