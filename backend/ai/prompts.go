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

var postInsightTemplates = []Template{
	{path: "post_insight.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureTEXTONLY},
	{path: "post_insight.schema.gotmpl", promptType: PromptTypeRESPONSESCHEMA, promptFeature: PromptFeatureBOTH},
	{path: "post_insight.human.gotmpl", promptType: PromptTypeHUMAN, promptFeature: PromptFeatureBOTH},
}

var subredditRulesEvalTemplates = []Template{
	{path: "subreddit_rules.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureTEXTONLY},
	{path: "subreddit_rules.schema.gotmpl", promptType: PromptTypeRESPONSESCHEMA, promptFeature: PromptFeatureBOTH},
	{path: "subreddit_rules.human.gotmpl", promptType: PromptTypeHUMAN, promptFeature: PromptFeatureBOTH},
}

var keywordSuggestionRedditTemplates = []Template{
	{path: "reddit_keyword_suggestion.prompt.gotmpl", promptType: PromptTypeSYSTEM, promptFeature: PromptFeatureTEXTONLY},
	{path: "reddit_keyword_suggestion.schema.gotmpl", promptType: PromptTypeRESPONSESCHEMA, promptFeature: PromptFeatureBOTH},
	{path: "reddit_keyword_suggestion.human.gotmpl", promptType: PromptTypeHUMAN, promptFeature: PromptFeatureBOTH},
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
