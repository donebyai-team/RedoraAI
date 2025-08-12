package interactions

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/datastore/psql"
	"github.com/shank318/doota/integrations/reddit"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/notifiers/alerts"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"github.com/shank318/doota/utils"
	"go.uber.org/zap"
)

type AutomatedInteractions interface {
	Authenticate(ctx context.Context, orgID string, cookieJSON string) (string, loginCallback, error)
	SendDM(ctx context.Context, interaction *models.LeadInteraction) error
	ScheduleComment(ctx context.Context, leadInteraction *models.LeadInteraction) (*models.LeadInteraction, error)
	ScheduleDM(ctx context.Context, leadInteraction *models.LeadInteraction) (*models.LeadInteraction, error)
	SendComment(ctx context.Context, interaction *models.LeadInteraction) (err error)
	GetInteractions(ctx context.Context, projectID string, status models.LeadInteractionStatus, dateRange pbportal.DateRangeFilter) ([]*models.LeadInteraction, error)
	ProcessScheduledPost(ctx context.Context, post *models.Post) error
}

type redditInteractions struct {
	db                datastore.Repository
	alertNotifier     alerts.AlertNotifier
	browserLessClient *browserless
	redditOauthClient *reddit.OauthClient
	logger            *zap.Logger
}

func (r redditInteractions) GetInteractions(ctx context.Context, projectID string, status models.LeadInteractionStatus, dateRange pbportal.DateRangeFilter) ([]*models.LeadInteraction, error) {
	return r.db.GetLeadInteractions(ctx, projectID, status, dateRange)
}

func NewRedditInteractions(db datastore.Repository, alertNotifier alerts.AlertNotifier, browserLessClient *browserless, redditOauthClient *reddit.OauthClient, logger *zap.Logger) AutomatedInteractions {
	return &redditInteractions{alertNotifier: alertNotifier, redditOauthClient: redditOauthClient, browserLessClient: browserLessClient, db: db, logger: logger}
}

func NewSimpleRedditInteractions(db datastore.Repository, logger *zap.Logger) AutomatedInteractions {
	return &redditInteractions{db: db, logger: logger}
}

func (r redditInteractions) ProcessScheduledPost(ctx context.Context, post *models.Post) (err error) {
	r.logger.Info("processing scheduled post", zap.String("id", post.ID))

	project, err := r.db.GetProject(ctx, post.ProjectID)
	if err != nil {
		r.logger.Error("failed to fetch project for post", zap.String("id", post.ID), zap.Error(err))
		return err
	}

	source, err := r.db.GetSourceByID(ctx, post.SourceID)
	if err != nil {
		r.logger.Error("failed to fetch source for post", zap.String("id", post.ID), zap.Error(err))
		return err
	}

	defer func() {
		if updateErr := r.db.UpdatePost(ctx, post); updateErr != nil {
			r.logger.Error("failed to update post in defer", zap.String("id", post.ID), zap.Error(updateErr))
			if err == nil {
				err = fmt.Errorf("post update failed: %w", updateErr)
			}
		}
	}()

	if !project.IsActive {
		post.Status = models.PostStatusFAILED
		post.Reason = "Project is not active"
		return nil
	}

	err = r.redditOauthClient.WithRotatingAPIClient(ctx, project.OrganizationID, func(client *reddit.Client) error {
		config := client.GetConfig()

		subredditName := utils.CleanSubredditName(source.Name)
		err = client.JoinSubreddit(ctx, subredditName)
		if err != nil && !strings.Contains(err.Error(), "403") {
			post.Status = models.PostStatusFAILED
			post.Reason = fmt.Sprintf("Reason: failed to join subreddit %v", err)
			return err
		}

		post.Metadata.Author = config.Name

		redditPost, err := client.CreatePost(ctx, subredditName, post)
		if err != nil {
			r.logger.Error("failed to post to Reddit", zap.String("id", post.ID), zap.Error(err))
			post.Status = models.PostStatusFAILED
			post.Reason = fmt.Sprintf("Failed to Post: %v", err)
			return err
		}

		post.PostID = &redditPost.ID
		post.Status = models.PostStatusSENT
		post.Reason = ""

		r.logger.Info("successfully posted to Reddit",
			zap.String("id", post.ID),
			zap.String("reddit_post_id", redditPost.ID),
		)
		return nil
	}, reddit.MostQualifiedAccountStrategy(r.logger))

	if err != nil {
		post.Status = models.PostStatusFAILED
		// if the reason is not set then set it to the error message
		if post.Reason == "" {
			post.Reason = err.Error()
		}
	}

	return err
}

func (r redditInteractions) SendComment(ctx context.Context, interaction *models.LeadInteraction) (err error) {
	if interaction.Type != models.LeadInteractionTypeCOMMENT {
		return fmt.Errorf("interaction type is not comment")
	}
	r.logger.Info("sending reddit comment",
		zap.String("interaction_id", interaction.ID),
		zap.String("from", interaction.From))

	// reset reason in case we retry
	interaction.Reason = ""

	project, err := r.db.GetProject(ctx, interaction.ProjectID)
	if err != nil {
		return err
	}

	redditLead, err := r.db.GetLeadByID(ctx, interaction.ProjectID, interaction.LeadID)
	if err != nil {
		return err
	}

	defer func() {
		// Always update interaction at the end
		updateErr := r.db.UpdateLeadInteraction(ctx, interaction)
		if updateErr != nil && err == nil {
			err = fmt.Errorf("failed to update interaction: %w", updateErr)
		}

		redditLead.LeadMetadata.CommentScheduledAt = nil
		updateError := r.db.UpdateLeadStatus(ctx, redditLead)
		if updateError != nil {
			r.logger.Warn("failed to update lead status for automated comment", zap.Error(err), zap.String("lead_id", redditLead.ID))
		}
	}()

	if !interaction.Organization.FeatureFlags.IsSubscriptionActive() {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "subscription has expired or not active"
		return nil
	}

	if !project.IsActive {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "project is not active"
		return nil
	}

	if redditLead.Status == models.LeadStatusNOTRELEVANT {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "Skipped, as user marked it as not relevant"
		return nil
	}

	if redditLead.Status == models.LeadStatusCOMPLETED {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "Skipped, as user has marked it responded manually"
		return nil
	}

	if strings.TrimSpace(utils.FormatComment(redditLead.LeadMetadata.SuggestedComment)) == "" {
		err := fmt.Errorf("no comment message found")
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = err.Error()
		return err
	}

	// case: if auto comment disabled
	if !interaction.Organization.FeatureFlags.EnableAutoComment {
		interaction.Status = models.LeadInteractionStatusFAILED
		interaction.Reason = "auto comment is disabled for this organization"
		return nil
	}

	err = r.redditOauthClient.WithRotatingAPIClient(ctx, interaction.Organization.ID, func(client *reddit.Client) error {
		interaction.From = client.GetConfig().Name

		//err = r.db.SetLeadInteractionStatusProcessing(ctx, interaction.ID)
		//if err != nil {
		//	return err
		//}

		subRedditName := utils.CleanSubredditName(redditLead.LeadMetadata.SubRedditPrefixed)
		if err = client.JoinSubreddit(ctx, subRedditName); err != nil {
			interaction.Reason = fmt.Sprintf("Failed to join subreddit: %v", err)
			interaction.Status = models.LeadInteractionStatusFAILED
			return err
		}

		var comment *reddit.Comment
		if comment, err = client.PostComment(ctx, fmt.Sprintf("t3_%s", interaction.To), utils.FormatComment(redditLead.LeadMetadata.SuggestedComment)); err != nil {
			interaction.Reason = fmt.Sprintf("Failed to post comment: %v", err)
			interaction.Status = models.LeadInteractionStatusFAILED
			return err
		}

		if comment == nil {
			interaction.Status = models.LeadInteractionStatusFAILED
			interaction.Reason = "comment is nil"
		} else {
			interaction.Status = models.LeadInteractionStatusSENT
			interaction.Reason = ""
			interaction.Metadata.ReferenceID = comment.ID
			interaction.Metadata.Permalink = fmt.Sprintf("r/%s/comments/%s/comment/%s", subRedditName, interaction.To, comment.ID)

			redditLead.LeadMetadata.AutomatedCommentURL = fmt.Sprintf("https://www.reddit.com/%s", interaction.Metadata.Permalink)
			redditLead.Status = models.LeadStatusAIRESPONDED
		}

		r.logger.Info("successfully sent reddit comment",
			zap.String("interaction_id", interaction.ID),
			zap.String("from", interaction.From))

		return nil
	}, reddit.PreferSpecificAccountStrategy(interaction.From))

	if err != nil {
		interaction.Status = models.LeadInteractionStatusFAILED
		// if the reason is not set then set it to the error message
		if interaction.Reason == "" {
			interaction.Reason = err.Error()
		}
		if errors.Is(err, reddit.AllAccountBanned) {
			r.disableAutomation(ctx, interaction, reddit.AllAccountBanned.Error())
		} else if errors.Is(err, reddit.AllAccountNotEstablished) {
			r.disableAutomation(ctx, interaction, reddit.AllAccountNotEstablished.Error())
		}
	}

	return err
}

func (r redditInteractions) disableAutomation(ctx context.Context, interaction *models.LeadInteraction, reason string) {
	if r.alertNotifier == nil {
		r.logger.Warn("alert notifier is not configured, skipping disable automation")
		return
	}

	if interaction.Type == models.LeadInteractionTypeCOMMENT {
		interaction.Organization.FeatureFlags.EnableAutoComment = false
		interaction.Organization.FeatureFlags.Activities = append(interaction.Organization.FeatureFlags.Activities, models.OrgActivity{
			ActivityType: models.OrgActivityTypeCOMMENTDISABLEDBYSYSTEM,
			CreatedAt:    time.Now(),
		})
		updates := map[string]any{
			psql.FEATURE_FLAG_DISABLE_AUTOMATED_COMMENT_PATH: false,
			psql.FEATURE_FLAG_ACTIVITIES_PATH:                interaction.Organization.FeatureFlags.Activities,
		}

		if err := r.db.UpdateOrganizationFeatureFlags(ctx, interaction.Organization.ID, updates); err != nil {
			r.logger.Error("failed to update organization feature flags", zap.Error(err))
		}

		go r.alertNotifier.SendAutoCommentDisabledEmail(context.Background(), interaction.Organization.ID, interaction.From, reason)
	} else if interaction.Type == models.LeadInteractionTypeDM {
		interaction.Organization.FeatureFlags.EnableAutoDM = false
		interaction.Organization.FeatureFlags.Activities = append(interaction.Organization.FeatureFlags.Activities, models.OrgActivity{
			ActivityType: models.OrgActivityTypeCOMMENTDISABLEDBYSYSTEM,
			CreatedAt:    time.Now(),
		})
		updates := map[string]any{
			psql.FEATURE_FLAG_DISABLE_AUTOMATED_DM_PATH: false,
			psql.FEATURE_FLAG_ACTIVITIES_PATH:           interaction.Organization.FeatureFlags.Activities,
		}

		if err := r.db.UpdateOrganizationFeatureFlags(ctx, interaction.Organization.ID, updates); err != nil {
			r.logger.Error("failed to update organization feature flags", zap.Error(err))
		}

		go r.alertNotifier.SendAutoDMDisabledEmail(context.Background(), interaction.Organization.ID, interaction.From, reason)
	}

	r.logger.Info("successfully disabled automation",
		zap.String("interaction_id", interaction.ID),
		zap.String("interaction_type", interaction.Type.String()),
		zap.String("org_id", interaction.Organization.ID),
		zap.String("reason", reason))
}

func (r redditInteractions) ScheduleComment(ctx context.Context, info *models.LeadInteraction) (*models.LeadInteraction, error) {
	r.logger.Info("creating interaction",
		zap.String("type", models.LeadInteractionTypeCOMMENT.String()),
		zap.String("thing_id", info.To),
	)
	info.Type = models.LeadInteractionTypeCOMMENT

	interactions, err := r.GetInteractions(ctx, info.ProjectID, models.LeadInteractionStatusCREATED, pbportal.DateRangeFilter_DATE_RANGE_TODAY)
	if err != nil {
		return nil, err
	}
	// Check if daily limits are reached

	scheduledAt, err := getNextAvailableScheduleTimeRandomBucket(time.Now().UTC(), interactions, 5*time.Minute, 1)
	if err != nil {
		return nil, err
	}

	info.ScheduledAt = utils.Ptr(scheduledAt)

	return r.db.CreateLeadInteraction(ctx, info)
}

func (r redditInteractions) ScheduleDM(ctx context.Context, info *models.LeadInteraction) (*models.LeadInteraction, error) {
	r.logger.Info("creating interaction",
		zap.String("type", models.LeadInteractionTypeDM.String()),
		zap.String("thing_id", info.To),
	)
	info.Type = models.LeadInteractionTypeDM

	if info.ScheduledAt == nil {
		interactions, err := r.GetInteractions(ctx, info.ProjectID, models.LeadInteractionStatusCREATED, pbportal.DateRangeFilter_DATE_RANGE_TODAY)
		if err != nil {
			return nil, err
		}

		scheduledAt, err := getNextAvailableScheduleTimeRandomBucket(time.Now().UTC(), interactions, 5*time.Minute, 1)
		if err != nil {
			return nil, err
		}
		info.ScheduledAt = utils.Ptr(scheduledAt)
	}

	return r.db.CreateLeadInteraction(ctx, info)
}

func getNextAvailableScheduleTimeRandomBucket(
	now time.Time,
	existingScheduled []*models.LeadInteraction,
	bucketSize time.Duration,
	maxPerBucket int,
) (time.Time, error) {
	if bucketSize <= 0 {
		return time.Time{}, errors.New("bucket size must be > 0")
	}

	rand.Seed(time.Now().UnixNano())

	for dayOffset := 0; dayOffset < 30; dayOffset++ {
		day := now.AddDate(0, 0, dayOffset)
		startOfDay := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)

		bucketCount := int(24 * time.Hour / bucketSize)
		buckets := make([]int, bucketCount)

		// Fill buckets for that day
		for _, i := range existingScheduled {
			if i.ScheduledAt == nil {
				continue
			}
			t := i.ScheduledAt.UTC()
			if t.Year() == day.Year() && t.YearDay() == day.YearDay() {
				offset := t.Sub(startOfDay)
				bucketIndex := int(offset / bucketSize)
				if bucketIndex >= 0 && bucketIndex < bucketCount {
					buckets[bucketIndex]++
				}
			}
		}

		startBucket := 0
		if dayOffset == 0 {
			startBucket = int(now.Sub(startOfDay) / bucketSize)
		}

		// Collect all available buckets >= startBucket
		availableBuckets := []int{}
		for b := startBucket; b < bucketCount; b++ {
			if buckets[b] < maxPerBucket {
				availableBuckets = append(availableBuckets, b)
			}
		}

		if len(availableBuckets) > 0 {
			// Pick one bucket randomly
			bucketIndex := availableBuckets[rand.Intn(len(availableBuckets))]
			bucketStart := startOfDay.Add(time.Duration(bucketIndex) * bucketSize)
			randomOffset := time.Duration(rand.Int63n(int64(bucketSize)))
			scheduledTime := bucketStart.Add(randomOffset)

			if scheduledTime.Before(now) {
				scheduledTime = now.Add(time.Minute)
			}

			return scheduledTime, nil
		}
	}

	return time.Time{}, errors.New("no available slots for scheduling within 30 days")
}
