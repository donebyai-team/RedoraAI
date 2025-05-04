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
	"github.com/shank318/doota/notifiers/alerts"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"sort"
	"strings"
	"time"
)

const redoraChannel = "https://hooks.slack.com/services/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"

type redditKeywordTracker struct {
	gptModel          ai.GPTModel
	db                datastore.Repository
	aiClient          *ai.Client
	logger            *zap.Logger
	state             state.ConversationState
	redditOauthClient *reddit.OauthClient
	isDev             bool
}

func newRedditKeywordTracker(
	isDev bool,
	gptModel ai.GPTModel,
	redditOauthClient *reddit.OauthClient,
	db datastore.Repository,
	aiClient *ai.Client,
	logger *zap.Logger,
	state state.ConversationState) KeywordTracker {
	return &redditKeywordTracker{
		gptModel:          gptModel,
		db:                db,
		aiClient:          aiClient,
		logger:            logger,
		state:             state,
		redditOauthClient: redditOauthClient,
		isDev:             isDev,
	}
}

func (s *redditKeywordTracker) WithLogger(logger *zap.Logger) KeywordTracker {
	return &redditKeywordTracker{
		gptModel:          s.gptModel,
		db:                s.db,
		aiClient:          s.aiClient,
		logger:            logger,
		state:             s.state,
		redditOauthClient: s.redditOauthClient,
		isDev:             s.isDev,
	}
}

func (s *redditKeywordTracker) TrackKeyword(ctx context.Context, tracker *models.AugmentedKeywordTracker) error {
	redditClient, err := s.redditOauthClient.NewRedditClient(ctx, tracker.Project.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to create reddit client: %w", err)
	}

	err = s.searchLeadsFromPosts(ctx, tracker.Keyword, tracker.Project, tracker.Source, redditClient)
	if err != nil {
		return err
	}

	err = s.db.UpdatKeywordTrackerLastTrackedAt(ctx, tracker.Tracker.ID)
	if err != nil {
		return err
	}

	// Once done, send the summary
	go s.sendAlert(context.Background(), tracker.Project)

	return nil
}

func (s *redditKeywordTracker) isTrackingDone(ctx context.Context, projectID string) (bool, error) {
	trackers, err := s.db.GetKeywordTrackerByProjectID(ctx, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch keyword trackers while checking completion: %w", err)
	}

	today := time.Now().Format(time.DateOnly)

	for _, tracker := range trackers {
		if tracker.LastTrackedAt == nil {
			return false, nil
		}
		if tracker.LastTrackedAt.Format(time.DateOnly) != today {
			return false, nil
		}
	}

	return true, nil
}

func (s *redditKeywordTracker) sendAlert(ctx context.Context, project *models.Project) {
	//isTrackingDoneKey := fmt.Sprintf("daily_tracking_summary:%s", projectID)
	//// Check if a call is already running across organizations
	//isRunning, err := s.state.IsRunning(ctx, isTrackingDoneKey)
	//if err == nil && isRunning {
	//	return false, nil
	//}
	//
	//if isRunning {
	//	return false, nil
	//}
	//
	//// Try to acquire the lock
	//if err := s.state.KeepAlive(ctx, orgID, isTrackingDoneKey); err != nil {
	//	s.logger.Error("could not acquire lock for keyword tracker summary, skipping", zap.Error(err))
	//}
	//
	//defer func() {
	//	if err := s.state.Release(ctx, isTrackingDoneKey); err != nil {
	//		s.logger.Error("failed to release lock on keyword tracker summary", zap.Error(err))
	//	}
	//}()

	integration, err := s.db.GetIntegrationByOrgAndType(ctx, project.OrganizationID, models.IntegrationTypeSLACKWEBHOOK)
	if err != nil && errors.Is(err, datastore.NotFound) {
		s.logger.Info("so integration configured for alerts")
		return
	}

	done, err := s.isTrackingDone(ctx, project.ID)
	if err != nil {
		s.logger.Error("check if tracking is done", zap.Error(err))
		return
	}

	// Send alert
	if done {
		dailyCount, err := s.getLeadsCountOfTheDay(ctx, project.ID, defaultRelevancyScore)
		if err != nil {
			s.logger.Error("failed to get leads count", zap.Error(err))
			return
		}
		leadsURL := "https://app.redoraai.com/dashboard/leads"

		msg := fmt.Sprintf(
			"*RedoraAI*\n\nProduct: %s\n\n *Daily Lead Summary*\n%d leads found today.\n<%s|View leads →>",
			project.Name,
			dailyCount,
			leadsURL,
		)

		clientWebhook := integration.GetSlackWebhook().Webhook

		err = alerts.NewSlackNotifier(clientWebhook).Send(ctx, msg)
		if err != nil {
			s.logger.Error("failed to send slack notification", zap.Error(err))
		}

		if clientWebhook != redoraChannel {
			err = alerts.NewSlackNotifier(clientWebhook).Send(ctx, msg)
			if err != nil {
				s.logger.Error("failed to send slack notification", zap.Error(err))
			}
		}
	}
}

//func (s *redditKeywordTracker) TrackPost(ctx context.Context,
//	post *models.Lead,
//	project *models.Project,
//	subReddit *models.Source,
//	redditClient *reddit.Client) error {
//	comments, err := redditClient.GetPostWithAllComments(ctx, post.PostID)
//	if err != nil {
//		return fmt.Errorf("failed to get reddit comments: %w", err)
//	}
//
//	return nil
//}

// Call GetPosts of a subreddit created on and after subReddit LastPostCreatedAt
// Filter them via a criteria - https://www.notion.so/Criteria-for-filtering-the-relevant-post-1c70029aaf8f80ec8ba6fd4e29342d6a
// After filtering, ask AI to filter again
// Save it into the table sub_reddits_leads (models.Lead)
func (s *redditKeywordTracker) searchLeadsFromPosts(
	ctx context.Context,
	keyword *models.Keyword,
	project *models.Project,
	source *models.Source,
	redditClient *reddit.Client) error {

	if ok, err := s.isMaxLeadLimitReached(ctx, project.ID, defaultRelevancyScore); err != nil || ok {
		return err
	}

	redditQuery := reddit.PostFilters{
		Keywords: []string{keyword.Keyword},
		SortBy:   utils.Ptr(reddit.SortByNEW),
		Limit:    100,
	}

	s.logger.Info("started tracking reddit keyword",
		zap.String("keyword", keyword.Keyword),
		zap.String("sub_reddit", source.Name),
		zap.Any("query", redditQuery))

	posts, err := redditClient.GetPosts(ctx, source.Name, redditQuery)

	if err != nil {
		return fmt.Errorf("unable to fetch posts: %w", err)
	}

	// Sort in DESC, we want to start from the most latest and keep going down till we find maxRelevantLeads per day
	sort.Slice(posts, func(i, j int) bool {
		return posts[j].CreatedAt > posts[i].CreatedAt
	})

	s.logger.Info("got posts from reddit", zap.Int("total_posts", len(posts)))

	newPosts := []*reddit.Post{}
	for _, post := range posts {
		lead, err := s.db.GetLeadByPostID(ctx, project.ID, post.ID)
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
	countPostsWithHighRelevancy := 0
	countSkippedPosts := 0
	countTestPosts := 0

	s.logger.Info("posts to be evaluated on relevancy via ai", zap.Int("total_posts", len(newPosts)))
	// Filter by AI
	for _, post := range newPosts {
		// TODO: Only on dev to avoid openai calls
		if countTestPosts >= 5 && s.isDev {
			s.logger.Info("dev mode is on, max 5 posts extracted via openai")
			break
		}

		redditLead := &models.Lead{
			ProjectID:     project.ID,
			SourceID:      source.ID,
			Author:        post.Author, // actual username not id
			PostID:        post.ID,
			Type:          models.LeadTypePOST,
			Title:         utils.Ptr(post.Title),
			Description:   post.Selftext,
			KeywordID:     keyword.ID,
			PostCreatedAt: time.Unix(int64(post.CreatedAt), 0),
			LeadMetadata: models.LeadMetadata{
				PostURL:           post.URL,
				AuthorURL:         fmt.Sprintf("https://www.reddit.com/user/%s/", post.Author),
				DmURL:             fmt.Sprintf("https://chat.reddit.com/user/%s/", post.AuthorFullName),
				SelfTextHTML:      post.SelftextHTML,
				SubRedditPrefixed: post.SubRedditPrefixed,
				Ups:               post.Ups,
				NoOfComments:      post.NumComments,
			},
		}

		isValid, reason := s.isValidPost(post)
		if isValid {
			countTestPosts++
			relevanceResponse, err := s.aiClient.IsRedditPostRelevant(ctx, project, redditLead, s.gptModel, s.logger)
			if err != nil {
				s.logger.Error("failed to get relevance response", zap.Error(err))
				continue
			}

			redditLead.RelevancyScore = relevanceResponse.IsRelevantConfidenceScore
			if redditLead.RelevancyScore >= defaultRelevancyScore {
				countPostsWithHighRelevancy++
			}

			redditLead.LeadMetadata.ChainOfThought = relevanceResponse.ChainOfThoughtIsRelevant
			redditLead.LeadMetadata.SuggestedComment = relevanceResponse.SuggestedComment
			redditLead.Intents = relevanceResponse.Intents
			redditLead.LeadMetadata.SuggestedDM = relevanceResponse.SuggestedDM
			redditLead.LeadMetadata.ChainOfThoughtSuggestedComment = relevanceResponse.ChainOfThoughtSuggestedComment
			redditLead.LeadMetadata.ChainOfThoughtSuggestedDM = relevanceResponse.ChainOfThoughtSuggestedDM
		} else {
			countSkippedPosts++
			s.logger.Info("ignoring reddit post for ai relevancy check",
				zap.String("post_id", post.ID),
				zap.String("author", post.Author),
				zap.String("reason", reason),
			)

			redditLead.RelevancyScore = 0
			redditLead.LeadMetadata.ChainOfThought = reason
		}

		err = s.db.CreateLead(ctx, redditLead)
		if err != nil {
			if datastore.IsUniqueViolation(err) {
				s.logger.Warn(
					"failed to create reddit lead",
					zap.Error(err),
					zap.String("post_id", post.ID))
			} else {
				return fmt.Errorf("unable to create reddit lead: %w", err)
			}
		}

		// Check if we have got enough relevant leads for the dat
		ok, err := s.isMaxLeadLimitReached(ctx, project.ID, defaultRelevancyScore)
		if err != nil || ok {
			if err != nil {
				return err
			}
			break
		}

	}

	s.logger.Info("reddit_leads_summary",
		zap.Int("total_posts_queried", len(posts)),
		zap.Int("total_new_posts", len(newPosts)),
		zap.Int("total_invalid_posts", countSkippedPosts),
		zap.Int("high_relevancy_posts", countPostsWithHighRelevancy))

	return nil
}

func (s *redditKeywordTracker) isMaxLeadLimitReached(ctx context.Context, projectID string, relevancyScore int) (bool, error) {
	count, err := s.getLeadsCountOfTheDay(ctx, projectID, relevancyScore)
	if err != nil {
		return false, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	if count >= maxLeadsPerDay {
		s.logger.Info("reached max leads per day",
			zap.Uint32("count", count),
			zap.String("start_date", today.String()),
			zap.String("end_date", tomorrow.String()))
		return true, nil
	}
	return false, nil
}

func (s *redditKeywordTracker) getLeadsCountOfTheDay(ctx context.Context, projectID string, relevancyScore int) (uint32, error) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	leadsData, err := s.db.CountLeadByCreatedAt(ctx, projectID, relevancyScore, today, tomorrow)
	if err != nil {
		return 0, err
	}
	return leadsData.Count, nil
}

const (
	minSelftextLength     = 30
	minTitleLength        = 5
	maxPostAgeInMonths    = 6
	minCommentThreshold   = 5
	maxLeadsPerDay        = 25
	defaultRelevancyScore = 90
)

var systemAuthors = []string{"[deleted]", "AutoModerator"}
var noSelfTextPosts = []string{"This post contains content not supported on old Reddit"}

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

func isValidPostDescription(selfText string) (bool, string) {
	if strings.TrimSpace(selfText) == "" {
		return false, "no post description"
	}
	// Check if the author contains any of the system author strings (case-insensitive)
	for _, a := range noSelfTextPosts {
		if strings.Contains(strings.ToLower(selfText), strings.ToLower(a)) {
			return false, selfText
		}
	}

	return true, ""
}

func (s *redditKeywordTracker) isValidPost(post *reddit.Post) (bool, string) {
	sixMonthsAgo := time.Now().AddDate(0, -maxPostAgeInMonths, 0).Unix()

	var reason string
	author := strings.TrimSpace(post.Author)

	if !isValidAuthor(author) {
		reason = "invalid or system author"
	}

	if len(strings.TrimSpace(post.Selftext)) < minSelftextLength || len(strings.TrimSpace(post.Title)) < minTitleLength {
		reason = "title or selftext is not big enough"
	}

	if int64(post.CreatedAt) < sixMonthsAgo || post.Archived {
		reason = fmt.Sprintf("post is older than %d months or has been archived", maxPostAgeInMonths)
	}

	isValid, rsn := isValidPostDescription(post.Selftext)
	if !isValid {
		reason = rsn
	}

	if reason != "" {
		return false, reason
	}

	return true, ""
}
