package ai

import (
	"context"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"github.com/streamingfast/dstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommunityGuidelines(t *testing.T) {
	t.Parallel()

	debugStore, err := dstore.NewStore("../../data/debugstore", "", "", false)
	require.NoError(t, err)

	tests := []struct {
		name               string
		rules              []string
		model              models.LLMModel
		wantMentionAllowed bool
	}{
		{
			name:  "bigseo",
			model: models.LLMModel("redora-dev-gpt-4.1-2025-04-14"),
			//model: models.LLMModel("redora-gemini-2.0-flash"),
			rules: []string{
				"Content that is unlikely to be useful to the experienced SEO professional. Homework and \"help me with my interview\" is included in this.",
				"Post contributes nothing of value and/or is blog spam that is UNLIKELY to help the average SEO professional.",
				"Do not copy and paste from blogs or news sites. This isn't a news subreddit. Mods may make exceptions around core updates and similar, but may also remove news links at their discretion.",
				"Post is not topical to SEO.",
				"Offers of services (sale or free), for hire posts, link-exchange or guest posting is not permitted. Affiliate links are not allowed. No prospecting for work of any kind. No \"free tools\" or beta tests. We don't care about your ProductHunt launch.",
				"Posts that are a link or headline only are subject to removal. Put some effort into it, please.",
				"Automod will remove posts if a user is posting multiple times in a day.",
				"Automod will slay you.",
			},
			wantMentionAllowed: false,
		},
		{
			name:  "saas",
			model: models.LLMModel("redora-dev-gpt-4.1-2025-04-14"),
			//model: models.LLMModel("redora-gemini-2.0-flash"),
			rules: []string{
				"Follow the Reddit site-wide rules and please treat others with respect, stay on-topic, and avoid non-productive self-promotion.\n\nNo spam.\n\nFeedback requests must be posted in the weekly feedback thread! (A post that will always be pinned at the top of the community)",
				"Promotion is ok here, but please don’t mention your SaaS/blog/company unless it’s relevant and actually helpful for someone reading. Overdoing it results in a ban.\n\nDirect sales that are unsolicited are forbidden as well. No PM requests please (unless people really request it), and no promotion for other communities outside Reddit.\n\nFeedback requests must be posted in the weekly feedback thread! (A post that will always be pinned at the top of the community)\n\nNo promotion of other communities.",
				"Please keep the discussions oriented around SaaS, tech companies, business in general or even personal aspects of the business world.\n\nIf posts are not somehow helping anyone in regards to the topic, removals and bans will be enforced.",
				"You may submit your blog post as long as the main ideas are in the Reddit post. \n\nYou need to provide value to people through your post and not simply present what you're talking about in your article. \n\nThe more value you provide in the body of the Reddit post, the safer it is to say that it won't be removed.\n\nThe only way a link is allowed is at the end of the post (“Originally posted here”), unless highly relevant (don't abuse this).\n\nAnything else will be removed/banned.",
				"At the end of the day, we’re all trying to make the world better for us and for those around us.\\n\\n\\\"Be nice and supportive\\\" is common sense. Try to criticise objectively and not personally.",
				"* Doxing－Posting or seeking personal information, dox attempts or threats\n* Flooding－Posting excessively frequently\n* Duplicates－Reposting news or information\n* Plagiarism－Not giving credit properly\n* Hyping－Pushing speculative, volatile, illiquid, or meme investments, especially flippantly, tersely, or implying huge returns\n* Missing－Disappearing after posting a discussion, posting for another with inadequate information",
				"We don't allow:\n\n* Moralizing issues\n* Petitions or calls-to-action\n* Political discussions\n* Political baiting\n* Soapboxing",
			},
			wantMentionAllowed: true,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ai, err := NewLLMClient(utils.GetEnvTestReq(t, "OPENAI_API_KEY_DEV"), tc.model, tc.model, LangsmithConfig{}, debugStore, logger)
			assert.NoError(t, err)

			source := &models.Source{
				Name:     tc.name,
				Metadata: models.SubRedditMetadata{Rules: tc.rules},
				OrgID:    "redora_dev__test",
			}

			relevant, _, err := ai.GetSourceCommunityRulesEvaluation(context.Background(), "", source, logger)
			require.NoError(t, err)

			assert.Equal(t, tc.wantMentionAllowed, relevant.ProductMentionAllowed)
		})
	}
}
