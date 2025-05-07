package ai

import (
	"context"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/dstore"
	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

var logger, _ = logging.PackageLogger("subreddit_tracker.text", "github.com/shank318/doota/redora/subreddit_tracker.test")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zap.DebugLevel))
}

func TestRelevancyOutputFormating(t *testing.T) {
	debugStore, err := dstore.NewStore("../../data/debugstore", "", "", false)
	if err != nil {
		t.FailNow()
	}
	//defaultModel := models.LLMModel("redora-dev-gpt-4.1-2025-04-14")
	defaultModel := models.LLMModel("redora-dev-gpt-4.1-mini-2025-04-14")
	ai, err := NewOpenAI(utils.GetEnvTestReq(t, "OPENAI_API_KEY_DEV"), defaultModel, LangsmithConfig{}, debugStore)
	if err != nil {
		t.FailNow()
	}

	project := &models.Project{
		ID:                 "XXX",
		OrganizationID:     "XXXXX",
		Name:               "MiraaAI",
		ProductDescription: "Miraa helps B2B SaaS businesses generate high quality leads and organic growth via Content-led SEO and GEO. We offers services like Content, SEO, and Reddit lead generation",
		CustomerPersona:    "Founders, CMOs, CEOs, and Head of marketing/growth",
	}

	post := &models.Lead{
		Author:      "Full-Entrepreneur591",
		Title:       utils.Ptr("Reddit Post: Title: Where to listing content agency?"),
		Description: "Hello, hi! It is my first post here. English my second language and I am still in process, so.. don't shame me please We’re small content agency and we trying to expand into english speaking markets. Our team is small, we don’t have big budgets — but we do have lots of enthusiasm and decent portfolio. Right now, I’m looking into listing us in directories and databases (ideally free or low-cost just to be present). Do you think this is a smart move for an agency like ours? Any tips on where to list ourselves besides Clutch, SE Ranking, and UpWork? Any advice would mean a lot, thank you in advance!",
	}

	org := &models.Organization{}

	relevant, usage, err := ai.IsRedditPostRelevant(context.Background(), org, project, post, logger)
	if err != nil {
		t.FailNow()
	}

	comment := utils.FormatComment(relevant.SuggestedComment)
	assert.NotEmpty(t, comment)
	assert.Equal(t, usage.Model, defaultModel)
}

//func Test2(t *testing.T) {
//	var resp = `{
//  "chain_of_thought": "1. The author is a founder/builder (direct fit for our persona).\n2. The post is BUILDING_IN_PUBLIC, announcing and seeking feedback on Olive Drift, a new tool for text analysis and subreddit identification.\n3. The author is implicitly interested in growing adoption/user base (mentions sign-ups, free trial, and excitement for its evolution), which often signals an openness to acquisition channels and growth tools.\n4. No explicit ask for a tool/recommendation, but the builder/founder persona plus the \"supporting platforms in the future\" hints at a growth mindset compatible with lead gen tools. \n5. Confidence is moderately high (75) due to persona fit and implied growth intent, but not as direct as a request for leads or outreach tools.",
//  "chain_of_thought_suggested_comment": "Since the author built something that organizes messy text data, and mentions launching a product/trial and asking for feedback, the comment should encourage discussion about their go-to-market and how they plan to get early users. This opens the door to later conversations about lead generation tools. No hard pitch. Just peer curiosity.",
//  "chain_of_thought_suggested_dm": "The DM can be more direct since founders appreciate growth tips. Reference their launch and ask about their strategy for finding early Reddit/online communities or leads, and mention that there are tools designed to surface relevant Reddit conversations—planting the RedoraAI idea without overt pitching.",
//  "intents": [
//    "BUILDING_IN_PUBLIC",
//    "ASKING_FOR_FEEDBACK",
//    "EXPRESSING_GOAL"
//  ],
//  "relevant_confidence_score": 75,
//  "suggested_comment": "Congrats on launching Olive Drift! Always love seeing projects that help make sense of noisy data.\n\nCurious—do you have a plan for finding your first batch of users, especially from platforms like Reddit? That part is usually the trickiest. Would love to hear your early growth approach!",
//  "suggested_dm": "Hey, just read about Olive Drift—really cool how you’re tackling messy data.\\n\\nAs you get ready to launch, are you thinking about ways to spot early users on places like Reddit? There are some tools that help surface relevant convos before most people see them. Happy to share more if you’re interested!"
//}`
//
//	resp = strings.ReplaceAll(resp, `\"`, `"`)
//
//	var relResponse models.RedditPostRelevanceResponse
//
//	err := json.Unmarshal([]byte(resp), &relResponse)
//	if err != nil {
//		return
//	}
//}
