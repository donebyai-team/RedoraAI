package redora

import (
	"context"
	"errors"
	"fmt"
	"github.com/shank318/doota/agents/redora/interactions"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/ai"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/datastore/psql"
	"github.com/shank318/doota/errorx"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/notifiers/alerts"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
	"sort"
	"strings"
	"time"
)

type redditKeywordTracker struct {
	db                    datastore.Repository
	aiClient              *ai.Client
	logger                *zap.Logger
	automatedInteractions interactions.AutomatedInteractions
	state                 state.ConversationState
	redditOauthClient     *reddit.OauthClient
	isDev                 bool
	alertNotifier         alerts.AlertNotifier
}

func newRedditKeywordTracker(
	isDev bool,
	redditOauthClient *reddit.OauthClient,
	automatedInteractions interactions.AutomatedInteractions,
	db datastore.Repository,
	aiClient *ai.Client,
	logger *zap.Logger,
	state state.ConversationState,
	alertNotifier alerts.AlertNotifier) KeywordTracker {
	return &redditKeywordTracker{
		db:                    db,
		aiClient:              aiClient,
		logger:                logger,
		state:                 state,
		redditOauthClient:     redditOauthClient,
		automatedInteractions: automatedInteractions,
		isDev:                 isDev,
		alertNotifier:         alertNotifier,
	}
}

func (s *redditKeywordTracker) WithLogger(logger *zap.Logger) KeywordTracker {
	return &redditKeywordTracker{
		db:                    s.db,
		aiClient:              s.aiClient,
		logger:                logger,
		state:                 s.state,
		redditOauthClient:     s.redditOauthClient,
		isDev:                 s.isDev,
		automatedInteractions: s.automatedInteractions,
		alertNotifier:         s.alertNotifier,
	}
}

func (s *redditKeywordTracker) TrackKeyword(ctx context.Context, tracker *models.AugmentedKeywordTracker) error {
	if !s.shouldTrack(tracker) {
		go s.disableProject(ctx, tracker.Organization)
		return nil
	}

	redditClient, err := s.redditOauthClient.GetRedditAPIClient(ctx, tracker.Project.OrganizationID, false)
	if err != nil {
		if errors.Is(err, datastore.IntegrationNotFoundOrActive) {
			return nil
		}
		var refreshTokenErr *errorx.RefreshTokenError
		if errors.As(err, &refreshTokenErr) {
			// TODO: Mark project as inactive
			s.logger.Warn("failed to refresh token, skipping tracking", zap.Error(refreshTokenErr))
			return nil
		}
		return err
	}

	err = s.searchLeadsFromPosts(ctx, tracker, redditClient)
	if err != nil {
		s.alertNotifier.SendTrackingError(ctx, tracker.GetID(), tracker.Project.Name, err)
		return err
	}

	err = s.db.UpdatKeywordTrackerLastTrackedAt(ctx, tracker.Tracker.ID)
	if err != nil {
		return err
	}

	// Once done, send the summary
	go s.sendAlert(context.Background(), tracker.Project, tracker.Organization)

	return nil
}

func (s *redditKeywordTracker) isTrackingDone(ctx context.Context, projectID string) (bool, error) {
	trackers, err := s.db.GetKeywordTrackerByProjectID(ctx, projectID)
	if err != nil {
		return false, fmt.Errorf("failed to fetch keyword trackers while checking completion: %w", err)
	}

	today := time.Now().UTC().Format(time.DateOnly)

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

func (s *redditKeywordTracker) shouldTrack(tracker *models.AugmentedKeywordTracker) bool {
	if tracker.Organization.FeatureFlags.IsSubscriptionExpired() || !tracker.Organization.FeatureFlags.IsSubscriptionActive() {
		return false
	}
	return tracker.Project.IsActive
}

func (s *redditKeywordTracker) disableProject(ctx context.Context, organization *models.Organization) {
	isDisableProjectKey := fmt.Sprintf("disable_project:%s", organization.ID)
	// Check if a call is already running across organizations
	isRunning, err := s.state.IsRunning(ctx, isDisableProjectKey)
	if err != nil {
		s.logger.Error("failed to check if disable project is running", zap.Error(err))
		return
	}
	if isRunning {
		return
	}

	// Try to acquire the lock
	if err := s.state.Acquire(ctx, organization.ID, isDisableProjectKey); err != nil {
		s.logger.Warn("could not acquire lock for daily_tracking_summary, skipped", zap.Error(err))
		return
	}

	defer func() {
		if err := s.state.Release(ctx, isDisableProjectKey); err != nil {
			s.logger.Error("failed to release lock on daily_tracking_summary", zap.Error(err))
		}
	}()

	err = s.db.UpdateProjectIsActive(ctx, organization.ID, false)
	if err != nil {
		s.logger.Error("failed to update project isActive", zap.Error(err))
		return
	}

	// Reason expired
	if organization.FeatureFlags.IsSubscriptionExpired() {
		// Notify User
		err = s.alertNotifier.SendTrialExpiredEmail(ctx, organization.ID, 7)
		if err != nil {
			s.logger.Error("failed to send trial expired email", zap.Error(err))
			return
		}
	}
}

func (s *redditKeywordTracker) sendAlert(ctx context.Context, project *models.Project, organization *models.Organization) {
	if !organization.FeatureFlags.ShouldSendRelevantPostAlert() {
		s.logger.Info("notification disabled, skipped sending alert")
		return
	}

	isTrackingDoneKey := fmt.Sprintf("daily_tracking_summary:%s", project.ID)
	// Check if a call is already running across organizations
	isRunning, err := s.state.IsRunning(ctx, isTrackingDoneKey)
	if err != nil {
		s.logger.Error("failed to check if daily tracking summary is running", zap.Error(err))
		return
	}
	if isRunning {
		return
	}

	// Try to acquire the lock
	if err := s.state.Acquire(ctx, project.OrganizationID, isTrackingDoneKey); err != nil {
		s.logger.Warn("could not acquire lock for daily_tracking_summary, skipped", zap.Error(err))
		return
	}

	defer func() {
		if err := s.state.Release(ctx, isTrackingDoneKey); err != nil {
			s.logger.Error("failed to release lock on daily_tracking_summary", zap.Error(err))
		}
	}()

	done, err := s.isTrackingDone(ctx, project.ID)
	if err != nil {
		s.logger.Error("check if tracking is done", zap.Error(err))
		return
	}

	notificationFrequency := organization.FeatureFlags.GetNotificationFrequency()

	// Send alert
	if done {
		s.logger.Info("daily tracking summary for project",
			zap.String("frequency", notificationFrequency.String()),
			zap.String("project_name", project.Name))

		defaultDateRange := pbportal.DateRangeFilter_DATE_RANGE_TODAY
		if notificationFrequency == models.NotificationFrequencyWEEKLY {
			defaultDateRange = pbportal.DateRangeFilter_DATE_RANGE_7_DAYS
		}

		analysis, err := NewLeadAnalysis(s.db, s.logger).GenerateLeadAnalysis(ctx, project.ID, defaultDateRange)
		if err != nil {
			s.logger.Error("failed to generate lead analysis", zap.Error(err))
			return
		}

		// Send alert on redora
		err = s.alertNotifier.SendLeadsSummary(ctx, alerts.LeadSummary{
			OrgID:                  project.OrganizationID,
			ProjectName:            project.Name,
			TotalPostsAnalysed:     analysis.PostsTracked,
			TotalCommentsScheduled: analysis.CommentScheduled,
			TotalDMScheduled:       analysis.DmScheduled,
			RelevantPostsCount:     analysis.RelevantPostsFound,
		})
		if err != nil {
			s.logger.Error("failed to send slack notification", zap.Error(err))
		}

		// Send Daily or Weekly alert to users
		// ----

		// if we found less than 2 relevant posts in a day, check for relevant posts in the last week
		// notify user if we found less relevant posts in the last week if the frequency is not weekly
		if notificationFrequency != models.NotificationFrequencyWEEKLY &&
			analysis.RelevantPostsFound < 2 {

			if !organization.FeatureFlags.ShouldSendNotEnoughRelevantPostsAlert() {
				return
			}

			weeklyAnalysis, err := NewLeadAnalysis(s.db, s.logger).GenerateLeadAnalysis(ctx, project.ID, pbportal.DateRangeFilter_DATE_RANGE_7_DAYS)
			if err != nil {
				s.logger.Error("failed to generate lead analysis", zap.Error(err))
				return
			}

			// if we find enough posts last week, we don't need to notify
			if weeklyAnalysis.RelevantPostsFound >= 8 {
				return
			}

			analysis = weeklyAnalysis
			// temo update the frequency to weekly
			// IMP - Do not update it in the DB
			notificationFrequency = models.NotificationFrequencyWEEKLY
		}

		err = s.alertNotifier.SendLeadsSummaryEmail(ctx, alerts.LeadSummary{
			OrgID:                  project.OrganizationID,
			ProjectName:            project.Name,
			TotalPostsAnalysed:     analysis.PostsTracked,
			TotalCommentsScheduled: analysis.CommentScheduled,
			TotalDMScheduled:       analysis.DmScheduled,
			RelevantPostsCount:     analysis.RelevantPostsFound,
		}, notificationFrequency)
		if err != nil {
			s.logger.Error("failed to send email notification", zap.Error(err))
		}

		// update last sent alert
		updates := map[string]any{
			psql.FEATURE_FLAG_NITIFICATION_LAST_SENT_AT_PATH: time.Now(),
		}

		err = s.db.UpdateOrganizationFeatureFlags(ctx, organization.ID, updates)
		if err != nil {
			s.logger.Error("failed to update organization while saving last_relevant_post_alert_sent_at", zap.Error(err))
			return
		}
	}
}

// Call GetPosts of a subreddit created on and after subReddit LastPostCreatedAt
// Filter them via a criteria - https://www.notion.so/Criteria-for-filtering-the-relevant-post-1c70029aaf8f80ec8ba6fd4e29342d6a
// After filtering, ask AI to filter again
// Save it into the table sub_reddits_leads (models.Lead)
func (s *redditKeywordTracker) searchLeadsFromPosts(
	ctx context.Context,
	tracker *models.AugmentedKeywordTracker,
	redditClient *reddit.Client) error {
	project := tracker.Project
	keyword := tracker.Keyword
	source := tracker.Source
	// to make it available downstream
	source.OrgID = project.OrganizationID

	// We will try to keep searching until we reach the max relevant posts per day >= defaultRelevancyScore
	if ok, err := s.isMaxLeadLimitReached(ctx, tracker.Organization); err != nil || ok {
		return err
	}

	redditQuery := reddit.QueryFilters{
		Keywords: []string{keyword.Keyword},
		SortBy:   utils.Ptr(reddit.SortByNEW),
		Limit:    100,
	}

	s.logger.Info("started tracking reddit keyword",
		zap.String("keyword", keyword.Keyword),
		zap.Any("query", redditQuery))

	if source.Metadata.RulesEvaluation != nil && source.Metadata.RulesEvaluation.ProductMentionAllowed {
		s.logger.Info("product mention allowed",
			zap.Bool("product_mention_allowed", source.Metadata.RulesEvaluation.ProductMentionAllowed))
	}

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
			s.logger.Debug("post already exists", zap.String("post_id", post.ID))
			continue
		}
		// Post doesn't exist, keep it
		newPosts = append(newPosts, post)
	}

	// Hard filters
	countPostsWithHighRelevancy := 0
	countSkippedPosts := 0
	aiErrorsCount := 0

	s.logger.Info("posts to be evaluated on relevancy via ai", zap.Int("total_posts", len(newPosts)))
	// Filter by AI
	for _, post := range newPosts {
		if aiErrorsCount >= defaultLLMFailedCount {
			return fmt.Errorf("more than %d llm called failed, skipped processing", defaultLLMFailedCount)
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
			Status:        models.LeadStatusNEW, // need it for calling update below
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
			relevanceResponse, usage, err := s.aiClient.IsRedditPostRelevant(ctx, tracker.Organization.FeatureFlags.RelevancyLLMModel, ai.IsPostRelevantInput{
				Project: project,
				Post:    redditLead,
				Source:  source,
			}, s.logger)
			if err != nil {
				s.logger.Error("failed to get relevance response", zap.Error(err), zap.String("post_id", post.ID))
				aiErrorsCount++
				continue
			}

			// if the lower model thinks it is relevant, verify it with the higher one and override it if it is
			if relevanceResponse.IsRelevantConfidenceScore >= defaultRelevancyScoreGlobal {
				s.logger.Info("calling relevancy with higher model", zap.String("higher_model", string(s.aiClient.GetAdvanceModel())), zap.String("post_id", post.ID))
				relevanceResponseHigherModel, usageHigherModel, errHigherModel := s.aiClient.IsRedditPostRelevant(ctx, s.aiClient.GetAdvanceModel(), ai.IsPostRelevantInput{
					Project: project,
					Post:    redditLead,
					Source:  source,
				}, s.logger)
				if errHigherModel != nil {
					s.logger.Error("failed to get relevance response from the higher model, continuing with the existing one", zap.Error(errHigherModel), zap.String("post_id", post.ID))
					aiErrorsCount++
				} else {
					s.logger.Info("llm response overridden",
						zap.String("old_model", string(usage.Model)),
						zap.String("new_model", string(usageHigherModel.Model)),
						zap.Any("old_relevancy_score", relevanceResponse.IsRelevantConfidenceScore),
						zap.Any("new_relevancy_score", relevanceResponseHigherModel.IsRelevantConfidenceScore),
						zap.String("post_id", post.ID))

					relevanceResponse = relevanceResponseHigherModel
					redditLead.LeadMetadata.LLMModelResponseOverriddenBy = usageHigherModel.Model
				}
			}

			redditLead.RelevancyScore = relevanceResponse.IsRelevantConfidenceScore
			redditLead.LeadMetadata.ChainOfThought = relevanceResponse.ChainOfThoughtIsRelevant
			redditLead.LeadMetadata.SuggestedComment = relevanceResponse.SuggestedComment
			redditLead.Intents = relevanceResponse.Intents
			redditLead.LeadMetadata.SuggestedDM = relevanceResponse.SuggestedDM
			redditLead.LeadMetadata.ChainOfThoughtSuggestedComment = relevanceResponse.ChainOfThoughtSuggestedComment
			redditLead.LeadMetadata.ChainOfThoughtSuggestedDM = relevanceResponse.ChainOfThoughtSuggestedDM
			redditLead.LeadMetadata.RelevancyLLMModel = usage.Model
			redditLead.LeadMetadata.AppliedRules = relevanceResponse.AppliedRules

			// Mark the tracker alive in case the execution taking too much time
			// Doing it here because that's the only place that takes time
			if err := s.state.KeepAliveTracker(ctx, project.ID, tracker.GetID()); err != nil {
				s.logger.Error("failed to mark tracker alive", zap.Error(err))
			}
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

		if redditLead.RelevancyScore < minRelevancyScore {
			redditLead.Title = utils.Ptr("[Redacted]")
			redditLead.Description = "[Redacted]"
		}

		// is post highly relevant
		// TODO: Do it after CreateLead as we have anyways did the relevancy check
		if redditLead.RelevancyScore >= defaultRelevancyScoreGlobal {
			countPostsWithHighRelevancy++
			isAllowed, err := s.isMaxLeadLimitUnderLimit(ctx, tracker.Organization)
			if err != nil {
				return err
			}
			if !isAllowed {
				s.logger.Info("max leads limit reached, skipping comment", zap.String("post_id", post.ID))
				break
			}
		}

		_, err = s.db.CreateLead(ctx, redditLead)
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

		// IMP: Make sure to send comment after saving the lead as we need lead id
		s.scheduleInteractions(ctx, tracker.Organization, redditLead)

		// skip the tracking counter for posts which are rejected because of aging
		if !strings.Contains(reason, "post is older than") {
			// track max posts to track per day
			shouldContinue, err := s.state.CheckIfUnderLimitAndIncrement(ctx, dailyCounterKey(project.OrganizationID), keyTrackedPostPerDay, maxPostsToTrackPerDay, 24*time.Hour)
			if err != nil {
				s.logger.Error("failed to check if daily_tracked_posts under limit and increment", zap.Error(err))
			}

			if !shouldContinue {
				s.logger.Info("daily_tracked_posts limit reached, skipping tracking", zap.String("post_id", post.ID))
				break
			}
		}

		// TODO:
		// CheckIfUnderLimitAndIncrement and isMaxLeadLimitUnderLimit can be combined into one function
		// it should increment the daily counters for both
		// and we should not need isMaxLeadLimitReached separately

		// We will try to keep searching until we reach the max relevant posts per day >= defaultRelevancyScore
		ok, err := s.isMaxLeadLimitReached(ctx, tracker.Organization)
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

const (
	keyCommentScheduledPerDay = "comment_scheduled"
	keyDMScheduledPerDay      = "dm_scheduled"
	keyRelevantLeadsPerDay    = "relevant_posts"
	keyTrackedPostPerDay      = "posts_tracked"
)

func (s *redditKeywordTracker) scheduleInteractions(ctx context.Context, org *models.Organization, redditLead *models.Lead) {
	var redditConfig *models.RedditConfig
	if redditLead.RelevancyScore >= org.FeatureFlags.GetRelevancyScoreComment() &&
		org.FeatureFlags.IsCommentAutomationEnabled() &&
		len(strings.TrimSpace(redditLead.LeadMetadata.SuggestedComment)) > 0 {
		// Get the client
		redditClient, err := s.redditOauthClient.GetRedditAPIClient(ctx, org.ID, true)
		if err == nil {
			redditConfig = redditClient.GetConfig()
			// Schedule comment
			err := s.sendAutomatedComment(ctx, org, redditConfig, redditLead)
			if err != nil {
				s.logger.Error("failed to schedule automated comment", zap.Error(err), zap.String("post_id", redditLead.PostID))
			}
		} else {
			s.logger.Error("failed to get reddit client, while scheduling comment", zap.Error(err))
		}
	}

	if redditLead.RelevancyScore >= org.FeatureFlags.GetRelevancyScoreDM() &&
		org.FeatureFlags.IsDMAutomationEnabled() &&
		len(strings.TrimSpace(redditLead.LeadMetadata.SuggestedDM)) > 0 {
		// Schedule DM
		err := s.sendAutomatedDM(ctx, org, redditConfig, redditLead)
		if err != nil {
			s.logger.Error("failed to schedule automated DM", zap.Error(err), zap.String("post_id", redditLead.ID))
		}
	}
}

func (s *redditKeywordTracker) sendAutomatedDM(ctx context.Context, org *models.Organization, redditConfig *models.RedditConfig, redditLead *models.Lead) error {
	if !org.FeatureFlags.IsDMAutomationEnabled() {
		return nil
	}

	// TODO: We need to use the name from the RedditDMConfig
	from := org.Name
	if redditConfig != nil {
		from = redditConfig.Name
	}

	// Continue
	redisKey := dailyCounterKey(org.ID)
	shouldDM, err := s.state.CheckIfUnderLimitAndIncrement(ctx, redisKey, keyDMScheduledPerDay, org.FeatureFlags.GetMaxDMsPerDay(), 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check if dm_scheduled under limit and increment: %w", err)
	}

	if shouldDM {
		interactionDM := &models.LeadInteraction{
			LeadID:    redditLead.ID,
			ProjectID: redditLead.ProjectID,
			From:      from,
			To:        redditLead.Author,
		}

		// Schedule DM a few minutes after the comment is sent to keep the lead warm
		if redditLead.LeadMetadata.CommentScheduledAt != nil {
			newTime := redditLead.LeadMetadata.CommentScheduledAt.Add(5 * time.Minute)
			interactionDM.ScheduledAt = &newTime
			s.logger.Info("scheduling the DM right after 5 minutes of comment to keep the lead warm")
		}

		interaction, err := s.automatedInteractions.ScheduleDM(ctx, interactionDM)
		if err != nil {
			rollbackErr := s.state.RollbackCounter(ctx, redisKey, keyDMScheduledPerDay)
			if rollbackErr != nil {
				s.logger.Error("failed to rollback counter", zap.Error(rollbackErr))
			}
			return err
		}

		if interaction != nil {
			redditLead.LeadMetadata.DMScheduledAt = interaction.ScheduledAt
			return s.db.UpdateLeadStatus(ctx, redditLead)
		}
	}

	return nil
}

var ignoreOldEnoughChecksForOrgs = []string{"0d40bd4d-15ba-48d1-b3db-7d8dae22b7dd"}

func (s *redditKeywordTracker) sendAutomatedComment(ctx context.Context, org *models.Organization, redditConfig *models.RedditConfig, redditLead *models.Lead) error {
	if redditConfig == nil {
		return nil
	}
	isOldEnough := redditConfig.IsUserOldEnough(2)
	if utils.Contains(ignoreOldEnoughChecksForOrgs, org.ID) {
		isOldEnough = true
	}

	autoCommentEnabled := org.FeatureFlags.IsCommentAutomationEnabled()

	//// Case 1: User is old enough, but auto comment is currently disabled because of OrgActivityTypeCOMMENTDISABLEDACCOUNTAGENEW
	//// enable it
	//if isOldEnough && !autoCommentEnabled && org.FeatureFlags.ActivityExists(models.OrgActivityTypeCOMMENTDISABLEDACCOUNTAGENEW) {
	//	org.FeatureFlags.EnableAutoComment = true
	//	activity := models.OrgActivityTypeCOMMENTENABLEDWARMEDUP
	//	org.FeatureFlags.Activities = append(org.FeatureFlags.Activities, models.OrgActivity{
	//		ActivityType: activity,
	//		CreatedAt:    time.Now(),
	//	})
	//	if err := s.db.UpdateOrganization(ctx, org); err != nil {
	//		return fmt.Errorf("failed to enable auto comment: %w", err)
	//	}
	//	s.logger.Info("enabled auto comment", zap.String("org_name", org.Name), zap.String("activity", activity.String()))
	//}

	// Case 2: User is not old enough, but auto comment is currently enabled — disable it
	if !isOldEnough && autoCommentEnabled {
		org.FeatureFlags.EnableAutoComment = false
		org.FeatureFlags.Activities = append(org.FeatureFlags.Activities, models.OrgActivity{
			ActivityType: models.OrgActivityTypeCOMMENTDISABLEDACCOUNTAGENEW,
			CreatedAt:    time.Now(),
		})

		updates := map[string]any{
			psql.FEATURE_FLAG_DISABLE_AUTOMATED_COMMENT_PATH: false,
			psql.FEATURE_FLAG_ACTIVITIES_PATH:                org.FeatureFlags.Activities,
		}

		if err := s.db.UpdateOrganizationFeatureFlags(ctx, org.ID, updates); err != nil {
			return fmt.Errorf("failed to disable auto comment: %w", err)
		}

		s.logger.Info("disabled auto comment", zap.String("org_name", org.Name), zap.String("activity", models.OrgActivityTypeCOMMENTDISABLEDACCOUNTAGENEW.String()))

		go s.alertNotifier.SendAutoCommentDisabledEmail(context.Background(), org.ID, redditConfig.Name, "Your Reddit account is less than 2 weeks old and may be at risk of suspension.")
	}

	// Only proceed if commenting is enabled and user is old enough
	if !(isOldEnough && org.FeatureFlags.EnableAutoComment) {
		return nil // skip commenting
	}

	// Continue
	redisKey := dailyCounterKey(org.ID)
	shouldComment, err := s.state.CheckIfUnderLimitAndIncrement(ctx, redisKey, keyCommentScheduledPerDay, org.FeatureFlags.GetMaxCommentsPerDay(), 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to check if comment_scheduled under limit and increment: %w", err)
	}

	if shouldComment {
		interaction, err := s.automatedInteractions.ScheduleComment(ctx, &models.LeadInteraction{
			LeadID:    redditLead.ID,
			ProjectID: redditLead.ProjectID,
			From:      redditConfig.Name,
			To:        redditLead.PostID,
		})
		if err != nil {
			rollbackErr := s.state.RollbackCounter(ctx, redisKey, keyCommentScheduledPerDay)
			if rollbackErr != nil {
				s.logger.Error("failed to rollback counter", zap.Error(rollbackErr))
			}
			return err
		}

		if interaction != nil {
			redditLead.LeadMetadata.CommentScheduledAt = interaction.ScheduledAt
			return s.db.UpdateLeadStatus(ctx, redditLead)
		}
	}

	return nil
}

func dailyCounterKey(orgID string) string {
	return fmt.Sprintf("org:%s:counters:%s", orgID, time.Now().UTC().Format("2006-01-02"))
}

func (s *redditKeywordTracker) isMaxLeadLimitUnderLimit(ctx context.Context, org *models.Organization) (bool, error) {
	return s.state.CheckIfUnderLimitAndIncrement(ctx, dailyCounterKey(org.ID), keyRelevantLeadsPerDay, org.FeatureFlags.GetMaxLeadsPerDay(), 24*time.Hour)
}

func (s *redditKeywordTracker) isMaxLeadLimitReached(ctx context.Context, org *models.Organization) (bool, error) {
	dailyCounters, err := s.state.GetLeadAnalysisCounters(ctx, dailyCounterKey(org.ID))
	if err != nil {
		return false, err
	}
	if dailyCounters.RelevantPostsFound >= uint32(org.FeatureFlags.GetMaxLeadsPerDay()) {
		s.logger.Info("reached max leads per day",
			zap.Uint32("count", dailyCounters.RelevantPostsFound))
		return true, nil
	}

	if dailyCounters.PostsTracked >= maxPostsToTrackPerDay {
		s.logger.Info("reached max posts to track per day",
			zap.Uint32("count", dailyCounters.PostsTracked))
		return true, nil
	}

	return false, nil
}

const (
	minSelftextLength           = 30
	minTitleLength              = 5
	maxPostAgeInDays            = 5
	defaultRelevancyScoreGlobal = 90 // relevancy score to re-confirm with higher model and also max leads
	dailyPostsRelevancyScore    = 80
	minRelevancyScore           = 70
	defaultLLMFailedCount       = 3
	maxPostsToTrackPerDay       = 600
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
	daysAgo := time.Now().UTC().AddDate(0, 0, -maxPostAgeInDays).Unix()
	author := strings.TrimSpace(post.Author)

	if !isValidAuthor(author) {
		return false, "invalid or system author"
	}

	if post.SubRedditType == "user" || post.SubRedditType == "private" {
		return false, "not a public subreddit"
	}

	if (post.SubRedditType != "public" && post.SubRedditType != "restricted") || !strings.HasPrefix(post.SubRedditPrefixed, "r/") {
		return false, "not a visible subreddit post"
	}

	if len(strings.TrimSpace(post.Selftext)) < minSelftextLength || len(strings.TrimSpace(post.Title)) < minTitleLength {
		return false, "title or selftext is not big enough"
	}

	if int64(post.CreatedAt) < daysAgo || post.Archived {
		return false, fmt.Sprintf("post is older than %d days or has been archived", maxPostAgeInDays)
	}

	if isValid, reason := isValidPostDescription(post.Selftext); !isValid {
		return false, reason
	}

	return true, ""
}
