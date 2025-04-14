package redora

import (
	"context"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"os"
	"testing"
)

var logger, _ = logging.PackageLogger("subreddit_tracker.text", "github.com/shank318/doota/redora/subreddit_tracker.test")

func init() {
	logging.InstantiateLoggers(logging.WithDefaultLevel(zap.DebugLevel))
}

func testRedditClient(t *testing.T) *reddit.Client {
	redditConfig := &models.RedditConfig{
		AccessToken: utils.GetEnvTestReq(t, "TEST_REDDIT_TOKEN"),
	}
	return reddit.NewClientWithConfig(redditConfig, logger)
}

func testClient(t *testing.T) *ai.Client {
	openAIAPiKey := os.Getenv("TEST_OPENAI_API_KEY")
	if openAIAPiKey == "" {
		t.Skip("skipping test, TEST_OPENAI_API_KEY environment must be set for those tests to run.")
	}

	openAIOrganization := os.Getenv("TEST_OPENAI_ORGANIZATION")
	if openAIOrganization == "" {
		t.Skip("skipping test, TEST_OPENAI_ORGANIZATION environment must be set for those tests to run.")
	}

	langsmithApiKey := os.Getenv("TEST_LANGSMITH_API_KEY")
	if langsmithApiKey == "" {
		t.Skip("skipping test, TEST_LANGSMITH_API_KEY environment must be set for those tests to run.")
	}

	client, err := ai.NewOpenAI(
		openAIAPiKey,
		openAIOrganization,
		ai.LangsmithConfig{
			ProjectName: "doota-test.",
			ApiKey:      langsmithApiKey,
		},
		nil)
	require.NoError(t, err)
	return client
}

func TestSubRedditTracker(t *testing.T) {
	tracker := SubRedditTracker{
		gptModel: ai.GPTModelGpt4O20240806,
		aiClient: testClient(t),
		logger:   logger,
	}

	subRedditToTrack := &models.AugmentedSubReddit{
		SubReddit: &models.SubReddit{
			ID:          "internal-subreddit",
			SubRedditID: "subreddit to track",
		},
		Keywords: []*models.Keyword{{
			Keyword: "AI",
		}, {
			Keyword: "AI",
		}},
		Project: &models.Project{
			ID:                 "test-project",
			OrganizationID:     "test-org",
			Name:               "",
			ProductDescription: "",
			CustomerPersona:    "",
			EngagementGoals:    "",
		},
	}

	posts, err := tracker.searchLeadsFromPosts(context.Background(), subRedditToTrack, testRedditClient(t))
	assert.NoError(t, err)
	assert.Len(t, posts, 1)
}
