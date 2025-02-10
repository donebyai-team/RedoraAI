package prompttypes

import (
	"encoding/json"
	"fmt"

	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/models"
)

type basicInfo struct {
	Model       string `json:"model"`
	Description string `json:"description"`
}

type promptTypeFiles struct {
	description string `json:"description"`
	//-------------------//
	name         string
	promptConfig *ai.Prompt
}

func (m *promptTypeFiles) getPromptConfig() *ai.Prompt {
	if m.promptConfig == nil {
		m.promptConfig = &ai.Prompt{}
	}
	return m.promptConfig
}

func (m *promptTypeFiles) validate() error {
	if m.name == "" {
		return fmt.Errorf("name is required")
	}
	if m.description == "" {
		return fmt.Errorf("description is required")
	}

	if m.promptConfig != nil {
		if m.promptConfig.PromptTmpl == "" {
			return fmt.Errorf("prompt are required for llm message types")
		}
	}

	return nil

}

func (m *promptTypeFiles) PromptType() *models.PromptType {
	var err error
	cnt, err := json.Marshal(m.getPromptConfig())
	if err != nil {
		panic(fmt.Errorf("marshal getPromptConfig: %w", err))
	}

	return &models.PromptType{
		Name:        m.name,
		Description: m.description,
		Config:      cnt,
	}
}
