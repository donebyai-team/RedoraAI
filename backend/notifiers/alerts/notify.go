package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/resend/resend-go/v2"
	"github.com/shank318/doota/agents/state"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"github.com/shank318/doota/notifiers/events"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type LeadSummary struct {
	OrgID                  string
	UserID                 string
	ProjectName            string
	TotalPostsAnalysed     uint32
	TotalCommentsScheduled uint32
	TotalDMScheduled       uint32
	RelevantPostsCount     uint32
}

type AlertNotifier interface {
	SendLeadsSummary(ctx context.Context, summary LeadSummary) error
	SendTrackingError(ctx context.Context, trackingID, project string, err error)
	SendLeadsSummaryEmail(ctx context.Context, summary LeadSummary, frequency models.NotificationFrequency) error
	SendNewUserAlert(ctx context.Context, orgName string)
	SendUserActivity(ctx context.Context, activity, orgName, redditUsername string)
	SendNewProductAddedAlert(ctx context.Context, project *models.Project)
	SendWelcomeEmail(ctx context.Context, email string)
	SendRedditChatConnectedAlert(ctx context.Context, email string)
	SendTrialExpiredEmail(ctx context.Context, orgID string, trialDays int) error
	SendInteractionError(ctx context.Context, interactionID string, err error)
	SendAutoCommentDisabledEmail(ctx context.Context, orgID string, redditUsername string, reason string)
	SendAutoDMDisabledEmail(ctx context.Context, orgID string, redditUsername string, reason string)
	SendIntegrationRevoked(ctx context.Context, orgID string, redditUsername string, reason string)
	SendSubscriptionCreatedEmail(ctx context.Context, orgID string)
	SendSubscriptionRenewedEmail(ctx context.Context, orgID string)
	SendSubscriptionCancelledEmail(ctx context.Context, orgID string)
}

type SlackNotifier struct {
	redisClient    state.ConversationState
	SlackClient    *http.Client
	ResendClient   *resend.Client
	db             datastore.Repository
	eventPublisher *events.EventPublisher
	logger         *zap.Logger
}

func NewSlackNotifier(
	resendAPIKey string,
	redisClient state.ConversationState,
	brevoIntegration *events.Brevo,
	db datastore.Repository,
	logger *zap.Logger) AlertNotifier {

	return &SlackNotifier{
		db:             db,
		logger:         logger,
		eventPublisher: events.NewEventPublisher(db, logger, brevoIntegration),
		redisClient:    redisClient,
		SlackClient:    &http.Client{Timeout: 10 * time.Second},
		ResendClient:   resend.NewClient(resendAPIKey),
	}
}

const redoraChannel = "https://hooks.slack.com/services/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"
const alertsChannel = "https://hooks.slack.com/services/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"

func (s *SlackNotifier) SendTrackingError(ctx context.Context, trackingID, project string, err error) {
	msg := fmt.Sprintf("*Tracking Error*\n "+
		"*Product:* %s\n"+
		"*TrackerID:* %s\n"+
		"*Error:* %s", project, trackingID, err.Error())
	err = s.send(ctx, msg, alertsChannel)
	if err != nil {
		s.logger.Error("failed to send error alert to redora channel", zap.Error(err))
		return
	}
}

func (s *SlackNotifier) SendInteractionError(ctx context.Context, interactionID string, err error) {
	msg := fmt.Sprintf("*Interaction Error*\n "+
		"*InteractionID:* %s\n"+
		"*Error:* %s", interactionID, err.Error())
	err = s.send(ctx, msg, alertsChannel)
	if err != nil {
		s.logger.Error("failed to send error alert to redora channel", zap.Error(err))
		return
	}
}

func (s *SlackNotifier) SendUserActivity(ctx context.Context, activity, orgName, redditUsername string) {
	redditURL := fmt.Sprintf("https://www.reddit.com/user/%s", redditUsername)

	msg := fmt.Sprintf(
		"*User Activity Recorded*\n"+
			"*Activity:* %s\n"+
			"*Organization:* %s\n"+
			"🔗 <%s|Reddit Account>",
		activity, orgName, redditURL,
	)

	if err := s.send(ctx, msg, redoraChannel); err != nil {
		s.logger.Error("failed to send user activity to Slack", zap.Error(err))
	}
}

func (s *SlackNotifier) SendAutoDMDisabledEmail(ctx context.Context, orgID string, redditUsername string, reason string) {
	// acquire lock before sending
	disableAutomatedDMAlertKey := fmt.Sprintf("disable_automated_dm:%s", orgID)
	// Check if a call is already running across organizations
	isRunning, err := s.redisClient.IsRunning(ctx, disableAutomatedDMAlertKey)
	if err != nil {
		s.logger.Error("failed to check if daily tracking summary is running", zap.Error(err))
		return
	}
	if isRunning {
		return
	}

	// Try to acquire the lock
	if err := s.redisClient.Acquire(ctx, orgID, disableAutomatedDMAlertKey); err != nil {
		s.logger.Warn("could not acquire lock for disableAutomatedDMAlertKey, skipped", zap.Error(err))
		return
	}

	users, err := s.db.GetUsersByOrgID(ctx, orgID)
	if err != nil {
		return
	}

	if len(users) == 0 {
		return
	}

	to := make([]string, 0, len(users))
	for _, user := range users {
		to = append(to, user.Email)
	}

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<body style="font-family: Arial, sans-serif; background-color: #f7f9fc; padding: 20px;">
  			<div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px;">
    			<h2>Automated DMs Disabled for Your Reddit Account</h2>
    			<p>We've detected an issue with your connected Reddit account <strong>u/%s</strong>:</p>
    			<p style="margin: 20px 0; padding: 15px; background-color: #fef3c7; border-left: 4px solid #facc15; border-radius: 4px;">
      			%s
    			</p>
    			<p>To continue using automation safely, please reconnect your Reddit account or reach out to us via in-app chat support. We've temporarily disabled automated DMs until this is resolved.</p>
    			<hr>
    			<footer style="font-size: 12px; color: #888;">
      				<p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
      				<p>Need help or have questions? <a href="mailto:adarsh@redoraai.com">adarsh@redoraai.com</a></p>
    			</footer>
  			</div>
			</body>
		</html>
`, redditUsername, reason)

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <leads@alerts.redoraai.com>",
		To:      to,
		Cc:      []string{"shashank@donebyai.team", "adarsh@redoraai.com"},
		Subject: "🚫 Auto DM Disabled",
		Html:    htmlBody,
	}

	_, err = s.ResendClient.Emails.Send(params)
	return
}

func (s *SlackNotifier) SendAutoCommentDisabledEmail(ctx context.Context, orgID string, redditUsername string, reason string) {
	// acquire lock before sending
	disableAutomatedCommentAlertKey := fmt.Sprintf("disable_automated_comment:%s", orgID)
	// Check if a call is already running across organizations
	isRunning, err := s.redisClient.IsRunning(ctx, disableAutomatedCommentAlertKey)
	if err != nil {
		s.logger.Error("failed to check if daily tracking summary is running", zap.Error(err))
		return
	}
	if isRunning {
		return
	}

	// Try to acquire the lock
	if err := s.redisClient.Acquire(ctx, orgID, disableAutomatedCommentAlertKey); err != nil {
		s.logger.Warn("could not acquire lock for disableAutomatedCommentAlertKey, skipped", zap.Error(err))
		return
	}

	users, err := s.db.GetUsersByOrgID(ctx, orgID)
	if err != nil {
		return
	}

	if len(users) == 0 {
		return
	}

	to := make([]string, 0, len(users))
	for _, user := range users {
		to = append(to, user.Email)
	}

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
			<body style="font-family: Arial, sans-serif; background-color: #f7f9fc; padding: 20px;">
  			<div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px;">
    			<h2>Automated Comments Disabled for Your Reddit Account</h2>
    			<p>We've detected an issue with your connected Reddit account <strong>u/%s</strong>:</p>
    			<p style="margin: 20px 0; padding: 15px; background-color: #fef3c7; border-left: 4px solid #facc15; border-radius: 4px;">
      			%s
    			</p>
    			<p>To continue using automation safely, please reconnect your Reddit account or reach out to us via in-app chat support. We've temporarily disabled automated comments until this is resolved.</p>
    			<hr>
    			<footer style="font-size: 12px; color: #888;">
      				<p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
      				<p>Need help or have questions? <a href="mailto:adarsh@redoraai.com">adarsh@redoraai.com</a></p>
    			</footer>
  			</div>
			</body>
		</html>
`, redditUsername, reason)

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <leads@alerts.redoraai.com>",
		To:      to,
		Cc:      []string{"shashank@donebyai.team", "adarsh@redoraai.com"},
		Subject: "🚫 Auto Commenting Disabled",
		Html:    htmlBody,
	}

	_, err = s.ResendClient.Emails.Send(params)
	return
}

func (s *SlackNotifier) SendNewProductAddedAlert(ctx context.Context, project *models.Project) {
	msg := fmt.Sprintf(
		"*New Product Added*\n"+
			"*Product:* %s\n"+
			"*Website:* %s",
		project.Name, project.WebsiteURL,
	)

	if err := s.send(ctx, msg, redoraChannel); err != nil {
		s.logger.Error("failed to send new product alert to Slack", zap.Error(err))
	}

	// update event
	err := s.eventPublisher.UpdateUsers(ctx, project.OrganizationID)
	if err != nil {
		s.logger.Error("failed to create user event", zap.Error(err))
	}
}

func (s *SlackNotifier) SendRedditChatConnectedAlert(ctx context.Context, email string) {
	msg := fmt.Sprintf(
		"*DM Automation Enabled*\n"+
			"*Email:* %s",
		email,
	)

	if err := s.send(ctx, msg, redoraChannel); err != nil {
		s.logger.Error("failed to send new user alert to Slack", zap.Error(err))
	}
}

func (s *SlackNotifier) SendNewUserAlert(ctx context.Context, email string) {
	msg := fmt.Sprintf(
		"*New User Onboarded*\n"+
			"*Email:* %s",
		email,
	)

	if err := s.send(ctx, msg, redoraChannel); err != nil {
		s.logger.Error("failed to send new user alert to Slack", zap.Error(err))
	}
}

// SendLeadsSummaryEmail sends a leads summary email based on the specified frequency.
func (s *SlackNotifier) SendLeadsSummaryEmail(ctx context.Context, summary LeadSummary, frequency models.NotificationFrequency) error {
	users, err := s.db.GetUsersByOrgID(ctx, summary.OrgID)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}

	to := make([]string, 0, len(users))
	for _, user := range users {
		to = append(to, user.Email)
	}

	var cc []string
	var subject string
	var htmlBody string
	var timePeriod string

	switch frequency {
	case models.NotificationFrequencyDAILY:
		timePeriod = "Today"
		subject = "📊 Daily Lead Summary"
	case models.NotificationFrequencyWEEKLY:
		timePeriod = "This Week"
		subject = "📈 Weekly Lead Summary"
	default:
		// Default to daily if an invalid frequency is provided
		timePeriod = "Today"
		subject = "📊 Daily Lead Summary"
	}

	// Determine CC recipients and HTML body based on relevant posts count
	if summary.RelevantPostsCount < 2 {
		cc = []string{"shashank@donebyai.team", "adarsh@redoraai.com"}
		htmlBody = fmt.Sprintf(`
           <!DOCTYPE html>
           <html>
           <body style="font-family: Arial, sans-serif; background-color: #f7f9fc; padding: 20px;">
             <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px;">
               <h2>We Didn't Find Many Relevant Posts %s</h2>
               <p><strong>Product:</strong> %s</p>
               <p><strong>Posts Analyzed:</strong> %d</p>
               <p><strong>Relevant Posts Found:</strong> <strong>%d</strong></p>
               <p>It looks like your current subreddits or keywords may not be returning enough relevant posts.</p>
               <p>👉 Consider updating them in your <a href="https://app.redoraai.com/dashboard">RedoraAI Dashboard</a> to improve your lead discovery.</p>
               <hr>
               <footer style="font-size: 12px; color: #888;">
                 <p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
                 <p>Need help? <a href="mailto:adarsh@redoraai.com">adarsh@redoraai.com</a></p>
               </footer>
             </div>
           </body>
           </html>
        `, timePeriod, summary.ProjectName, summary.TotalPostsAnalysed, summary.RelevantPostsCount)
	} else {
		htmlBody = fmt.Sprintf(`
           <!DOCTYPE html>
           <html>
           <body style="font-family: Arial, sans-serif; background-color: #f7f9fc; padding: 20px;">
             <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px;">
               <h2>%s Reddit Posts Summary</h2>
               <p><strong>Product:</strong> %s</p>
               <p><strong>Posts Analyzed:</strong> %d</p>
               <p><strong>Automated Comments Scheduled:</strong> %d</p>
              <p><strong>Automated DM Scheduled:</strong> %d</p>
               <p><strong>Relevant Posts Found:</strong> <strong>%d</strong></p>
               <p>🔗 <a href="%s">View all leads in your dashboard</a></p>
               <hr>
               <footer style="font-size: 12px; color: #888;">
                 <p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
                 <p>Need help? <a href="mailto:adarsh@redoraai.com">adarsh@redoraai.com</a></p>
               </footer>
             </div>
           </body>
           </html>
        `, timePeriod, summary.ProjectName, summary.TotalPostsAnalysed, summary.TotalCommentsScheduled, summary.TotalDMScheduled, summary.RelevantPostsCount, "https://app.redoraai.com/dashboard")
	}

	params := &resend.SendEmailRequest{ // Changed to SendEmailRequest from resend.SendEmailRequest for mock
		From:    "RedoraAI <leads@alerts.redoraai.com>",
		To:      to,
		Subject: subject,
		Html:    htmlBody,
	}

	if cc != nil && len(cc) > 0 {
		params.Cc = cc
	}

	_, err = s.ResendClient.Emails.Send(params)
	return err
}

func (s *SlackNotifier) SendTrialExpiredEmail(ctx context.Context, orgID string, trialDays int) error {
	users, err := s.db.GetUsersByOrgID(ctx, orgID)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}

	to := make([]string, 0, len(users))
	for _, user := range users {
		to = append(to, user.Email)
	}

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f7f9fc; padding: 20px;">
		  <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px;">
		    <h2>Your Free Trial Has Expired</h2>
		    <p>Your %d-day free trial on RedoraAI has ended. To continue receiving relevant Reddit leads and automated outreach, please upgrade your plan.</p>
		    <p>🚀 Unlock full access and keep your growth going!</p>
		    <div style="margin: 20px 0;">
		      <a href="https://app.redoraai.com" 
		         style="display: inline-block; padding: 12px 24px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 6px;">
		        Upgrade Your Plan
		      </a>
		    </div>
		    <hr>
		    <footer style="font-size: 12px; color: #888;">
		      <p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
		      <p>Need help or have questions? <a href="mailto:adarsh@redoraai.com">adarsh@redoraai.com</a></p>
		    </footer>
		  </div>
		</body>
		</html>
	`, trialDays)

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <leads@alerts.redoraai.com>",
		To:      to,
		Cc:      []string{"shashank@donebyai.team", "adarsh@redoraai.com"},
		Subject: "🚫 Your RedoraAI Trial Has Ended — Upgrade to Stay Live",
		Html:    htmlBody,
	}

	_, err = s.ResendClient.Emails.Send(params)
	return err
}

func (s *SlackNotifier) SendWelcomeEmail(ctx context.Context, email string) {
	//users, err := s.db.GetUsersByOrgID(ctx, orgID)
	//if err != nil {
	//	s.logger.Error("failed to send welcome email", zap.Error(err))
	//	return
	//}
	//
	//// Only send it for the first one
	//if len(users) == 0 || len(users) > 1 {
	//	return
	//}

	htmlBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<body style="font-family: Arial, sans-serif; background-color: #f7f9fc; padding: 20px;">
	  <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px;">
	    <h2>Welcome to <strong>RedoraAI</strong> 👋</h2>
	    <p>We're excited to have you onboard! RedoraAI helps you discover and engage with high-intent leads from Reddit — automatically.</p>
	    
	    <h3>🚀 Here’s how to get started:</h3>
	    <ol>
	      <li><strong>Tell us about your product</strong> — so we can tailor your outreach.</li>
	      <li><strong>Select keywords & subreddits</strong> — to track relevant discussions.</li>
	      <li><strong>Enable Redora Copilot</strong> — to automate intelligent comments and DMs to potential leads.</li>
	    </ol>
	    
	    <p>🔗 <a href="https://app.redoraai.com/onboarding" style="color: #3366cc;">Begin Onboarding Now</a></p>
	    
	    <hr>
	    <footer style="font-size: 12px; color: #888;">
			<p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
			<p>Need help? <a href="mailto:adarsh@redoraai.com">adarsh@redoraai.com</a></p>
		</footer>
	  </div>
	</body>
	</html>
`)

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <welcome@alerts.redoraai.com>",
		To:      []string{email},
		Cc:      []string{"shashank@donebyai.team", "adarsh@redoraai.com"},
		Subject: "🔥Welcome aboard — here’s what to do next",
		Html:    htmlBody,
	}

	_, err := s.ResendClient.Emails.Send(params)
	if err != nil {
		s.logger.Error("failed to send welcome email", zap.Error(err))
	}

	// update event
	err = s.eventPublisher.CreateUser(ctx, email)
	if err != nil {
		s.logger.Error("failed to create user event", zap.Error(err))
	}
}

func (s *SlackNotifier) SendLeadsSummary(ctx context.Context, summary LeadSummary) error {
	integrations, err := s.db.GetIntegrationByOrgAndType(ctx, summary.OrgID, models.IntegrationTypeSLACKWEBHOOK)
	if err != nil && errors.Is(err, datastore.NotFound) {
		s.logger.Info("no integration configured for alerts, skipped")
	}

	leadsURL := "https://app.redoraai.com/dashboard"

	msg := fmt.Sprintf(
		"*📊 Daily Reddit Posts Summary*\n"+
			"*Product:* %s\n"+
			"*Posts Analyzed:* %d\n"+
			"*Automated Comments Scheduled:* %d\n"+
			"*Automated DM Scheduled:* %d\n"+
			"*Relevant Posts Found:* *%d*\n\n"+
			"🔗 <%s|View all posts in your dashboard>",
		summary.ProjectName,
		summary.TotalPostsAnalysed,
		summary.TotalCommentsScheduled,
		summary.TotalDMScheduled,
		summary.RelevantPostsCount,
		leadsURL,
	)

	defer func() {
		err := s.send(ctx, msg, redoraChannel)
		if err != nil {
			s.logger.Error("failed to send slack message to redora channel", zap.Error(err))
		} else {
			s.logger.Info("sent slack message to redora channel")
		}
	}()

	if len(integrations) != 0 {
		err := s.send(ctx, msg, integrations[0].GetSlackWebhook().Webhook)
		if err != nil {
			return err
		} else {
			s.logger.Info("sent slack message to slack channel", zap.String("channel", integrations[0].GetSlackWebhook().Channel))
		}
	}
	return nil
}

func (s *SlackNotifier) send(ctx context.Context, message string, webhook string) error {
	payload := map[string]string{
		"text": message,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal Slack payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhook, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("create Slack request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.SlackClient.Do(req)
	if err != nil {
		return fmt.Errorf("send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Slack returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}
