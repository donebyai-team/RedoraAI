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
	ai, err := NewOpenAI(utils.GetEnvTestReq(t, "OPENAI_API_KEY_DEV"), "", LangsmithConfig{}, debugStore)
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

	relevant, err := ai.IsRedditPostRelevant(context.Background(), project, post, GPTModelRedoraDevGpt4O20240806, logger)
	if err != nil {
		t.FailNow()
	}

	comment := utils.FormatComment(relevant.SuggestedComment)
	assert.NotEmpty(t, comment)
}
