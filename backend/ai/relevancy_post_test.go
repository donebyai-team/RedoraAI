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

func TestRelevancyOutputFormatting(t *testing.T) {
	debugStore, err := dstore.NewStore("../../data/debugstore", "", "", false)
	assert.NoError(t, err)

	cases := []struct {
		name    string
		model   models.LLMModel
		project *models.Project
		post    *models.Lead
		source  *models.Source
	}{
		{
			name:  "MiraaAI - mini course for solo creators",
			model: models.LLMModel("redora-dev-gpt-4.1-2025-04-14"),
			//model: models.LLMModel("redora-gemini-2.5-flash-preview-04-17"),
			project: &models.Project{
				ID:                 "XXX",
				OrganizationID:     "XXXXX",
				Name:               "MiraaAI",
				ProductDescription: "Miraa helps B2B SaaS businesses generate high quality leads and organic growth via Content-led SEO and GEO. We offers services like Content, SEO, and Reddit lead generation",
				CustomerPersona:    "Founders, CMOs, CEOs, and Head of marketing/growth",
			},
			source: &models.Source{
				Name: "bigseo",
				Metadata: models.SubRedditMetadata{
					RulesEvaluation: &models.RuleEvaluationResult{
						ProductMentionAllowed: false,
					},
				},
			},
			post: &models.Lead{
				Author:      "Feeling_Ad_4458",
				Title:       utils.Ptr("Got fired. Trying to build a SaaS that actually helps small service providers stand out. Stuck on the value part — would love your thoughts"),
				Description: "Hey everyone, I recently got fired and decided it’s time to take a real shot at independence. I live in a country of about 7 million people. It’s not a huge market, but I believe people are willing to pay for something that speaks their language , literally and culturally.\n\nOver here, there are tons of fitness trainers, yoga and pilates instructors, NLP coaches, private tutors, emotional therapists, and other solo service providers. Most of them are active on Instagram or TikTok. They post reels, stories, run ads, hire digital marketers. And honestly? It’s all starting to look the same.\n\nSame voiceovers. Same captions. Same generic editing. When I’m looking for a service myself, I hate being sold to like that. I don’t want a slogan or a hype video — I want to get a real feel for the person. I want to hear them talk, see how they think.\n\nThat led me to an idea.\n\nWhat if I gave these professionals a way to show who they really are not through another ad but through a mini-course generator. Something simple and beautiful. They upload a video or two, write a few lines about their method, maybe add a short quiz — and boom, it becomes a personal landing page they can link to from their bio. The idea is to help them stand out, explain their approach better than a 15-second reel ever could, and win trust.\n\nSo I started working on this.\n\nThe platform generates a beautiful little course, like a teaser that helps the potential client understand the trainer’s mindset, style, or process. For example, a fitness coach might share videos on how to train arms, and one video about motivation and how he gets clients to stay consistent.\n\nSounds great, right?\n\nHere’s where I got stuck.\n\nThat coach now has a great-looking link they can share. People might watch, get value, and even reach out. But then… that’s it. They don’t need to create another mini-course. They just needed a nice way to present themselves once. Why would they pay monthly for that?\n\nI started realizing that the coach doesn’t want a platform — they want more clients. So now I’m at a crossroads. Do I pivot into something that helps generate leads, not just present better? Or is there a way to build ongoing value around that mini-course idea?\n\nI still love the concept of helping service providers differentiate themselves through deeper, more honest content. But I’m not sure how to turn that into something they’ll happily pay for every month.\n\nWould really appreciate any insights, directions, or even examples of tools that are doing something similar.\n\nThanks for reading.",
			},
		},
		//{
		//	name:  "MiraaAI - SEO rank in US from Malaysia",
		//	model: models.LLMModel("redora-dev-gpt-4.1-2025-04-14"),
		//	project: &models.Project{
		//		ID:                 "XXX",
		//		OrganizationID:     "XXXXX",
		//		Name:               "MiraaAI",
		//		ProductDescription: "Miraa helps B2B SaaS businesses generate high quality leads and organic growth via Content-led SEO and GEO. We offers services like Content, SEO, and Reddit lead generation",
		//		CustomerPersona:    "Founders, CMOs, CEOs, and Head of marketing/growth",
		//	},
		//	post: &models.Lead{
		//		Author:      "Feeling_Ad_4458",
		//		Title:       utils.Ptr("Query Alert: Need some help here - Any content SEO strategies to rank in another country? -"),
		//		Description: "Say I am sitting in KL (Malaysia) and want my content to rank in the NY and New Jersey area (US). Looking for some grey hat ideas.",
		//	},
		//},
		//{
		//	name:  "SalesForge.ai - AI prompts to boost conversion",
		//	model: models.LLMModel("redora-gemini-2.0-flash"),
		//	project: &models.Project{
		//		ID:                 "XXX",
		//		OrganizationID:     "XXXXX",
		//		Name:               "SalesForge.ai",
		//		ProductDescription: "Salesforge is an email outreach tool...",
		//		CustomerPersona:    "Sales Reps, Founders, CMOs, Head of Sales, Head of Growth.",
		//	},
		//	post: &models.Lead{
		//		Author:      "Feeling_Ad_4458",
		//		Title:       utils.Ptr("5 AI Prompts That Boost Conversions Instantly"),
		//		Description: "If you’re running a business and haven’t tapped into AI for sales copy, you’re leaving money on the table.\n\nHere are 5 prompts I personally use:\n\n1. Cold email template for SaaS founders targeting tech startups.\n2. Landing page structure for a new SaaS product launch.\n3. Re-engagement email for inactive users — get them back!\n4. LinkedIn DM prompt for selling SaaS consulting services.\n5. Chat simulation: Handling the \"We’re not ready to commit\" objection.\n\nIf you want 45 more prompts for every stage of the funnel, comment “Prompts, please!” and I’ll DM it to you instantly.",
		//	},
		//},
		//{
		//	name:  "RedoraAI - App Store optimization tips",
		//	model: models.LLMModel("redora-gemini-2.0-flash"),
		//	project: &models.Project{
		//		ID:                 "XXX",
		//		OrganizationID:     "XXXXX",
		//		Name:               "RedoraAI",
		//		ProductDescription: "Redora is a Reddit lead generation tool...",
		//		CustomerPersona:    "Sales Reps, Founders, CMOs, Head of Sales, Head of Growth.",
		//	},
		//	post: &models.Lead{
		//		Author:      "Feeling_Ad_4458",
		//		Title:       utils.Ptr("App store screenshots best practices that actually increase downloads (7 data-backed tips)"),
		//		Description: "Hey appmarketing folks!\n\nI'm Cristian, from Apptweak. After analyzing hundreds of apps across categories, I wanted to share some real, actionable insights about optimizing your app store screenshots - something I see many developers struggle with.\n\nWhy app screenshots matter more than you think\nResearch shows users form an opinion about your app within just 50 milliseconds based on visuals alone. Your screenshots are often the first (and sometimes only) impression users get before deciding to download.\n\nIn my experience, well-optimized screenshots can boost conversion rates significantly, especially since:\n\nThe first three screenshots have the biggest impact - users rarely scroll past them\nScreenshots directly affect search rankings indirectly through improved conversion rates\nVisuals communicate value far more effectively than descriptions (which many users skip entirely)\n7 data-backed best practices for app store screenshots\n1. Focus on the first three screenshots\nPlace your most compelling features in these prime positions. For example, Duolingo strategically showcases their core user benefits in their first three screenshots, highlighting the gamification and learning aspects immediately.\n\nDuolingo’s first three app screenshots on the App Store highlight user benefits.\n\n2. Add short, benefit-driven captions\nInstead of generic labels like \"Workout tracker app,\" use benefit-focused text like \"Track your workouts effortlessly\" or \"Learn a new language in 3 months.\" This conversion-focused approach performs better in testing.\n\nBonus tip: If you have strong brand recognition, subtly place your insignia in the first screenshot for social proof (like Asana does).\n\nFitness app Asana leverages insignia for social proof in its app screenshots.\n\n3. Include human elements for emotional connection\nNeuromarketing studies show users respond better to emotion-driven visuals than plain interface designs. Apps like Bumble effectively use human faces in their screenshots to create immediate emotional connections.\n\n4. Use bright colors and high contrast\nApps with bright colors and high-contrast designs consistently see higher conversion rates. Blue and green hues tend to convey trust, while red and orange create urgency and excitement.\n\n5. Don't forget dark mode screenshots\nIf your app supports dark mode, showcase it! This highlights adaptability and helps your listing stand out in app stores where most screenshots use light mode.\n\n6. Capture attention with video previews\nVideo previews appear before screenshots and can significantly boost conversions. Keep them 15-30 seconds, focusing on real in-app experiences (YouTube does this really well).\n\nVideo app previews really grab attention like this of YouTube’s.\n\n7. A/B test everything\nDon't just guess what works - test it. Use App Store custom product pages or external A/B testing tools to validate your hypotheses. One game developer, AppQuantum, increased downloads by 21.5% through creative A/B testing.\n\nCommon mistakes to avoid\nText overload: Keep captions concise and impactful\nGeneric stock images: These feel disconnected and create distrust\nPoor contrast: Ensure text is readable on all devices\nNot using all available slots: Use the maximum allowed (10 for App Store, 8 for Google Play)\nIgnoring user intent: Align screenshots with what users are looking for\nNot showing actual UI: Users want to see the real experience\nDifferent approaches by app category\nI've noticed distinct patterns across categories:\n\nGaming apps often benefit from landscape screenshots to showcase gameplay and environments.\n\nGaming apps like X-War: Clash of Zombies benefit from utilizing landscape screenshots to better showcase gameplay.\n\nUtility apps need clear demonstrations of functionality with strong CTAs.\n\nLifestyle apps focus on aesthetics and aspirational imagery that creates emotional connections.\n\nFinance apps emphasize trust and security first, showcasing features second.\n\nTechnical requirements you should know\nApp Store:\n\nUp to 10 screenshots per localization\nJPEG or PNG format (72 dpi, RGB color space)\nNo promotional text like \"Free\" or discount mentions\nCommon sizes:\niPhone 6.5\": 1242 × 2688 pixels (portrait)\niPad Pro 12.9\": 2048 × 2732 pixels (portrait)\nGoogle Play:\n\nUp to 8 screenshots per device type\nJPEG or 24-bit PNG format (no transparency)\nMinimum dimension: 320px; maximum: 3840px\nWhat screenshot optimization techniques have worked for your apps? Have you noticed any unexpected patterns in what converts best for your specific category?\n\nI'm actively monitoring this thread and would love to share more insights based on your specific challenges. Drop your questions below!\n\nWant to get the full scoop? We’ve put together a blog post packed with the best practices to optimize your app screenshots. Don’t miss it!",
		//	},
		//},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ai, err := NewOpenAI(utils.GetEnvTestReq(t, "OPENAI_API_KEY_DEV"), tc.model, tc.model, LangsmithConfig{}, debugStore, logger)
			assert.NoError(t, err)

			relevant, usage, err := ai.IsRedditPostRelevant(context.Background(), tc.model, IsPostRelevantInput{
				Project: tc.project,
				Post:    tc.post,
				Source:  tc.source,
			}, logger)
			assert.NoError(t, err)

			comment := utils.FormatComment(relevant.SuggestedComment)
			assert.NotEmpty(t, comment)
			assert.GreaterOrEqual(t, relevant.IsRelevantConfidenceScore, 90)
			assert.Equal(t, usage.Model, tc.model)
		})
	}
}
