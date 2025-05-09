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

func TestRelevancyOutputFormating2(t *testing.T) {
	debugStore, err := dstore.NewStore("../../data/debugstore", "", "", false)
	if err != nil {
		t.FailNow()
	}
	defaultModel := models.LLMModel("redora-dev-gpt-4.1-2025-04-14")
	//defaultModel := models.LLMModel("redora-gemini-2.0-flash")
	ai, err := NewOpenAI(utils.GetEnvTestReq(t, "OPENAI_API_KEY_DEV"), defaultModel, LangsmithConfig{}, debugStore, logger)
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
		Author:      "Feeling_Ad_4458",
		Title:       utils.Ptr("Got fired. Trying to build a SaaS that actually helps small service providers stand out. Stuck on the value part — would love your thoughts"),
		Description: "Hey everyone, I recently got fired and decided it’s time to take a real shot at independence. I live in a country of about 7 million people. It’s not a huge market, but I believe people are willing to pay for something that speaks their language , literally and culturally.\n\nOver here, there are tons of fitness trainers, yoga and pilates instructors, NLP coaches, private tutors, emotional therapists, and other solo service providers. Most of them are active on Instagram or TikTok. They post reels, stories, run ads, hire digital marketers. And honestly? It’s all starting to look the same.\n\nSame voiceovers. Same captions. Same generic editing. When I’m looking for a service myself, I hate being sold to like that. I don’t want a slogan or a hype video — I want to get a real feel for the person. I want to hear them talk, see how they think.\n\nThat led me to an idea.\n\nWhat if I gave these professionals a way to show who they really are not through another ad but through a mini-course generator. Something simple and beautiful. They upload a video or two, write a few lines about their method, maybe add a short quiz — and boom, it becomes a personal landing page they can link to from their bio. The idea is to help them stand out, explain their approach better than a 15-second reel ever could, and win trust.\n\nSo I started working on this.\n\nThe platform generates a beautiful little course, like a teaser that helps the potential client understand the trainer’s mindset, style, or process. For example, a fitness coach might share videos on how to train arms, and one video about motivation and how he gets clients to stay consistent.\n\nSounds great, right?\n\nHere’s where I got stuck.\n\nThat coach now has a great-looking link they can share. People might watch, get value, and even reach out. But then… that’s it. They don’t need to create another mini-course. They just needed a nice way to present themselves once. Why would they pay monthly for that?\n\nI started realizing that the coach doesn’t want a platform — they want more clients. So now I’m at a crossroads. Do I pivot into something that helps generate leads, not just present better? Or is there a way to build ongoing value around that mini-course idea?\n\nI still love the concept of helping service providers differentiate themselves through deeper, more honest content. But I’m not sure how to turn that into something they’ll happily pay for every month.\n\nWould really appreciate any insights, directions, or even examples of tools that are doing something similar.\n\nThanks for reading.",
	}

	org := &models.Organization{}

	relevant, usage, errRel := ai.IsRedditPostRelevant(context.Background(), org, project, post, logger)
	if errRel != nil {
		t.FailNow()
	}

	comment := utils.FormatComment(relevant.SuggestedComment)
	assert.NotEmpty(t, comment)
	assert.True(t, relevant.IsRelevantConfidenceScore >= 90)
	assert.Equal(t, usage.Model, defaultModel)
}

func TestRelevancyOutputFormating(t *testing.T) {
	debugStore, err := dstore.NewStore("../../data/debugstore", "", "", false)
	if err != nil {
		t.FailNow()
	}
	defaultModel := models.LLMModel("redora-dev-gpt-4.1-2025-04-14")
	//defaultModel := models.LLMModel("redora-gemini-2.0-flash")
	ai, err := NewOpenAI(utils.GetEnvTestReq(t, "OPENAI_API_KEY_DEV"), defaultModel, LangsmithConfig{}, debugStore, logger)
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
		Author:      "Feeling_Ad_4458",
		Title:       utils.Ptr("Query Alert: Need some help here - Any content SEO strategies to rank in another country? -\n"),
		Description: "Say I am sitting in KL (Malaysia) and want my content to rank in the NY and New Jersey area (US). Looking for some grey hat ideas.",
	}

	org := &models.Organization{}

	relevant, usage, errRel := ai.IsRedditPostRelevant(context.Background(), org, project, post, logger)
	if errRel != nil {
		t.FailNow()
	}

	comment := utils.FormatComment(relevant.SuggestedComment)
	assert.NotEmpty(t, comment)
	assert.True(t, relevant.IsRelevantConfidenceScore >= 90)
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
