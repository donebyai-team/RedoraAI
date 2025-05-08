package ai

import (
	"embed"
	"github.com/shank318/doota/models"
	"github.com/streamingfast/cli"
)

var caseDecisionTemplates = []Template{
	{path: "case_decision.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureTEXTONLY},
	{path: "case_decision.schema.gotmpl", promptType: PromptTypeRESPONSESCHEMA, promptFeature: PromptFeatureBOTH},
	{path: "case_decision.human.gotmpl", promptType: PromptTypeHUMAN, promptFeature: PromptFeatureBOTH},
}

var redditPostRelevancyTemplates = []Template{
	{path: "reddit_post.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureTEXTONLY},
	{path: "reddit_post.schema.gotmpl", promptType: PromptTypeRESPONSESCHEMA, promptFeature: PromptFeatureBOTH},
	{path: "reddit_post.human.gotmpl", promptType: PromptTypeHUMAN, promptFeature: PromptFeatureBOTH},
}

//go:generate go-enum -f=$GOFILE

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

type Prompt struct {
	Model      models.LLMModel `json:"model"`
	PromptTmpl string          `json:"prompt_tmpl"`
	SchemaTmpl string          `json:"schema_tmpl"`
	HumanTmpl  string          `json:"human_tmpl"`
}

//go:embed prompts/*.gotmpl
var promptTplFS embed.FS

func rp(name string) string {
	cnt, err := promptTplFS.ReadFile("prompts/" + name)
	if err != nil {
		panic(err)
	}
	return cli.Dedent(string(cnt))
}
