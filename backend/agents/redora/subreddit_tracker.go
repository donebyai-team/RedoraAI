package redora

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"sort"
	"strings"
	"time"
)

type SubRedditTracker struct {
	gptModel          ai.GPTModel
	db                datastore.Repository
	aiClient          *ai.Client
	logger            *zap.Logger
	state             state.ConversationState
	redditOauthClient *reddit.OauthClient
}

func NewSubRedditTracker(
	gptModel ai.GPTModel,
	redditOauthClient *reddit.OauthClient,
	db datastore.Repository,
	aiClient *ai.Client,
	logger *zap.Logger,
	state state.ConversationState) *SubRedditTracker {
	return &SubRedditTracker{
		gptModel:          gptModel,
		db:                db,
		aiClient:          aiClient,
		logger:            logger,
		state:             state,
		redditOauthClient: redditOauthClient,
	}
}

func (s *SubRedditTracker) WithLogger(logger *zap.Logger) *SubRedditTracker {
	return &SubRedditTracker{
		gptModel:          s.gptModel,
		db:                s.db,
		aiClient:          s.aiClient,
		logger:            logger,
		state:             s.state,
		redditOauthClient: s.redditOauthClient,
	}
}

func (s *SubRedditTracker) TrackSubreddit(ctx context.Context, subReddit *models.AugmentedSubReddit) error {
	defer func() {
		err := s.state.Release(ctx, subReddit.SubReddit.ID)
		if err != nil {
			s.logger.Error("failed to release lock on subreddit", zap.Error(err))
		}
	}()

	redditClient, err := s.redditOauthClient.NewRedditClient(ctx, subReddit.Project.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to create reddit client: %w", err)
	}

	// Lock the subreddit tracking
	err = s.state.KeepAlive(ctx, subReddit.Project.OrganizationID, subReddit.SubReddit.ID)
	if err != nil {
		return fmt.Errorf("unable to lock subReddit: %w", err)
	}

	for _, keyword := range subReddit.Keywords {
		err = s.searchLeadsFromPosts(ctx, keyword, subReddit.Project, subReddit.SubReddit, redditClient)
		if err != nil {
			return err
		}

		err := s.db.UpdateSubRedditLastTrackedAt(ctx, subReddit.SubReddit.ID)
		if err != nil {
			s.logger.Error("failed to update subRedditLastTrackedAt", zap.Error(err))
		}
	}

	return nil
}

// Call GetPosts of a subreddit created on and after subReddit LastPostCreatedAt
// Filter them via a criteria - https://www.notion.so/Criteria-for-filtering-the-relevant-post-1c70029aaf8f80ec8ba6fd4e29342d6a
// After filtering, ask AI to filter again
// Save it into the table sub_reddits_leads (models.RedditLead)
func (s *SubRedditTracker) searchLeadsFromPosts(
	ctx context.Context,
	keyword *models.Keyword,
	project *models.Project,
	subReddit *models.SubReddit,
	redditClient *reddit.Client) error {

	redditQuery := reddit.PostFilters{
		Keywords: []string{keyword.Keyword},
		SortBy:   utils.Ptr(reddit.SortByNEW),
		Limit:    100,
	}
	posts, err := redditClient.GetPosts(ctx, subReddit.Name, redditQuery)

	if err != nil {
		return fmt.Errorf("unable to fetch posts: %w", err)
	}

	// Sort in DESC
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt > posts[j].CreatedAt
	})

	s.logger.Info("got posts from reddit",
		zap.Any("query", redditQuery),
		zap.Int("total_posts", len(posts)))

	newPosts := []*reddit.Post{}
	for _, post := range posts {
		lead, err := s.db.GetRedditLeadByPostID(ctx, project.ID, post.ID)
		if err != nil && !errors.Is(err, datastore.NotFound) {
			// Unexpected error, log and skip
			s.logger.Error("error while checking if lead exists by post id", zap.Error(err))
			continue
		}
		if err == nil && lead != nil {
			s.logger.Info("post already exists", zap.String("post_id", post.ID))
			continue
		}
		// Post doesn't exist, keep it
		newPosts = append(newPosts, post)
	}

	// Hard filters
	filteredPosts := s.filterAndEnrichPosts(newPosts)
	countPostsWithHighRelevancy := 0

	s.logger.Info("posts to be evaluated on relevancy via ai", zap.Int("total_posts", len(newPosts)))
	// Filter by AI
	for _, post := range filteredPosts {
		redditLead := &models.RedditLead{
			ProjectID:     project.ID,
			SubRedditID:   subReddit.ID,
			Author:        post.Author,
			PostID:        post.ID,
			Type:          models.LeadTypePOST,
			Title:         utils.Ptr(post.Title),
			Description:   post.Selftext,
			PostCreatedAt: time.Unix(int64(post.CreatedAt), 0),
		}

		relevanceResponse, err := s.aiClient.IsRedditPostRelevant(ctx, project, redditLead, s.gptModel, s.logger)
		if err != nil {
			s.logger.Error("failed to get relevance response", zap.Error(err))
			continue
		}

		redditLead.RelevancyScore = relevanceResponse.IsRelevantConfidenceScore
		if redditLead.RelevancyScore >= 90 {
			countPostsWithHighRelevancy++
		}

		redditLead.LeadMetadata = models.LeadMetadata{
			ChainOfThought:                   relevanceResponse.ChainOfThoughtIsRelevant,
			SuggestedComment:                 relevanceResponse.SuggestedComment,
			SuggestedDM:                      relevanceResponse.SuggestedDM,
			ChainOfThoughtSuggestedComment:   relevanceResponse.ChainOfThoughtSuggestedComment,
			ChainOfThoughtCommentSuggestedDM: relevanceResponse.ChainOfThoughtSuggestedDM,
			PostURL:                          post.URL,
			AuthorInfo:                       post.AuthorInfo,
		}
		err = s.db.CreateRedditLead(ctx, redditLead)
		if err != nil {
			return fmt.Errorf("unable to create reddit lead: %w", err)
		}
	}

	s.logger.Info("reddit_leads_summary",
		zap.Int("total_posts_queried", len(posts)),
		zap.Int("total_new_posts", len(newPosts)),
		zap.Int("total_posts_after_filtering", len(filteredPosts)),
		zap.Int("high_relevancy_posts", countPostsWithHighRelevancy))

	return nil
}

const (
	minSelftextLength   = 30
	minTitleLength      = 5
	maxPostAgeInMonths  = 6
	minCommentThreshold = 5
)

var systemAuthors = []string{"[deleted]", "AutoModerator"}

func isValidAuthor(author string) bool {
	if strings.TrimSpace(author) == "" {
		return false
	}
	// Check if the author contains any of the system author strings (case-insensitive)
	for _, a := range systemAuthors {
		if strings.Contains(strings.ToLower(author), strings.ToLower(a)) {
			return false
		}
	}

	return true
}

func (s *SubRedditTracker) filterAndEnrichPosts(posts []*reddit.Post) []*reddit.Post {
	filteredPosts := []*reddit.Post{}
	sixMonthsAgo := time.Now().AddDate(0, -maxPostAgeInMonths, 0).Unix()

	for _, post := range posts {
		var reason string
		author := strings.TrimSpace(post.Author)

		if !isValidAuthor(author) {
			reason = "invalid or system author"
		}

		if len(strings.TrimSpace(post.Selftext)) < minSelftextLength || len(strings.TrimSpace(post.Title)) < minTitleLength {
			reason = "title or selftext is not big enough"
		}

		if int64(post.CreatedAt) < sixMonthsAgo && post.NumComments < minCommentThreshold {
			reason = fmt.Sprintf("post is older than %d months and has less than %d comments", maxPostAgeInMonths, minCommentThreshold)
		}

		if reason != "" {
			s.logger.Info("ignoring reddit post",
				zap.String("post_id", post.ID),
				zap.String("author", post.Author),
				zap.String("reason", reason),
			)
			continue
		}

		filteredPosts = append(filteredPosts, post)
	}

	return filteredPosts
}
