package ai

import (
	"context"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/dstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommunityGuidelines(t *testing.T) {
	debugStore, err := dstore.NewStore("../../data/debugstore", "", "", false)
	if err != nil {
		t.FailNow()
	}
	//defaultModel := models.LLMModel("redora-dev-gpt-4.1-2025-04-14")
	defaultModel := models.LLMModel("redora-gemini-2.0-flash")
	ai, err := NewOpenAI(utils.GetEnvTestReq(t, "OPENAI_API_KEY_DEV"), defaultModel, LangsmithConfig{}, debugStore, logger)
	if err != nil {
		t.FailNow()
	}

	rules := []string{
		"Content that is unlikely to be useful to the experienced SEO professional. Homework and \"help me with my interview\" is included in this.",
		"Post contributes nothing of value and/or is blog spam that is UNLIKELY to help the average SEO professional.",
		"Do not copy and paste from blogs or news sites. This isn't a news subreddit. Mods may make exceptions around core updates and similar, but may also remove news links at their discretion.",
		"Post is not topical to SEO.",
		"Offers of services (sale or free), for hire posts, link-exchange or guest posting is not permitted. Affiliate links are not allowed. No prospecting for work of any kind. No \"free tools\" or beta tests. We don't care about your ProductHunt launch.",
		"Posts that are a link or headline only are subject to removal. Put some effort into it, please.",
		"Automod will remove posts if a user is posting multiple times in a day.",
		"Automod will slay you.",
	}

	source := &models.Source{
		Name:     "r/saas",
		Metadata: models.SubRedditMetadata{Rules: rules},
		OrgID:    "redora_dev__test",
	}

	relevant, _, errRel := ai.GetSourceCommunityRulesEvaluation(context.Background(), "", source, logger)
	if errRel != nil {
		t.FailNow()
	}

	assert.Equal(t, relevant.ProductMentionAllowed, false)
}
