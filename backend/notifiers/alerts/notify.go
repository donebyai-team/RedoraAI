package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/resend/resend-go/v2"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
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
	DailyCount             uint32
}

type AlertNotifier interface {
	SendLeadsSummary(ctx context.Context, summary LeadSummary) error
	SendTrackingError(ctx context.Context, trackingID, project string, err error)
	SendLeadsSummaryEmail(ctx context.Context, summary LeadSummary) error
	SendNewUserAlert(ctx context.Context, orgName string)
	SendUserActivity(ctx context.Context, activity, orgName, redditUsername string)
	SendNewProductAddedAlert(ctx context.Context, productName, website string)
	SendWelcomeEmail(ctx context.Context, orgID string)
}

type SlackNotifier struct {
	SlackClient  *http.Client
	ResendClient *resend.Client
	db           datastore.Repository
	logger       *zap.Logger
}

func NewSlackNotifier(resendAPIKey string, db datastore.Repository, logger *zap.Logger) AlertNotifier {
	return &SlackNotifier{
		db:           db,
		logger:       logger,
		SlackClient:  &http.Client{Timeout: 10 * time.Second},
		ResendClient: resend.NewClient(resendAPIKey),
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

func (s *SlackNotifier) SendNewProductAddedAlert(ctx context.Context, productName, website string) {
	msg := fmt.Sprintf(
		"*New Product Added*\n"+
			"*Product:* %s\n"+
			"*Website:* %s",
		productName, website,
	)

	if err := s.send(ctx, msg, redoraChannel); err != nil {
		s.logger.Error("failed to send new product alert to Slack", zap.Error(err))
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

func (s *SlackNotifier) SendLeadsSummaryEmail(ctx context.Context, summary LeadSummary) error {
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

	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<body style="font-family: Arial, sans-serif; background-color: #f7f9fc; padding: 20px;">
		  <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px;">
		    <h2>Daily Reddit Posts Summary — <strong>RedoraAI</strong></h2>
		    <p><strong>Product:</strong> %s</p>
		    <p><strong>Posts Analyzed:</strong> %d</p>
		    <p><strong>Automated Comments Scheduled:</strong> %d</p>
			<p><strong>Automated DM Scheduled:</strong> %d</p>
		    <p><strong>Relevant Posts Found:</strong> <strong>%d</strong></p>
		    <p>🔗 <a href="%s">View all leads in your dashboard</a></p>
		    <hr>
		    <footer style="font-size: 12px; color: #888;">
		      <p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
		      <p>Need help? <a href="mailto:shashank@donebyai.team">shashank@donebyai.team</a></p>
		    </footer>
		  </div>
		</body>
		</html>
	`, summary.ProjectName, summary.TotalPostsAnalysed, summary.TotalCommentsScheduled, summary.TotalDMScheduled, summary.DailyCount, "https://app.redoraai.com/dashboard/leads")

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <leads@alerts.redoraai.com>",
		To:      to,
		Subject: "📊Daily Lead Summary",
		Html:    htmlBody,
	}

	_, err = s.ResendClient.Emails.Send(params)
	return err
}

func (s *SlackNotifier) SendWelcomeEmail(ctx context.Context, orgID string) {
	users, err := s.db.GetUsersByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to send welcome email", zap.Error(err))
		return
	}

	// Only send it for the first one
	if len(users) == 0 || len(users) > 1 {
		return
	}

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
			<p>Need help? <a href="mailto:shashank@donebyai.team">shashank@donebyai.team</a></p>
		</footer>
	  </div>
	</body>
	</html>
`)

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <welcome@alerts.redoraai.com>",
		To:      []string{users[0].Email},
		Cc:      []string{"shashank@donebyai.team"},
		Subject: "🔥Welcome aboard — here’s what to do next",
		Html:    htmlBody,
	}

	_, err = s.ResendClient.Emails.Send(params)
	if err != nil {
		s.logger.Error("failed to send welcome email", zap.Error(err))
	}
}

func (s *SlackNotifier) SendLeadsSummary(ctx context.Context, summary LeadSummary) error {
	integration, err := s.db.GetIntegrationByOrgAndType(ctx, summary.OrgID, models.IntegrationTypeSLACKWEBHOOK)
	if err != nil && errors.Is(err, datastore.NotFound) {
		s.logger.Info("no integration configured for alerts, skipped")
	}

	leadsURL := "https://app.redoraai.com/dashboard/leads"

	msg := fmt.Sprintf(
		"*📊 Daily Reddit Posts Summary — RedoraAI*\n"+
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
		summary.DailyCount,
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

	if integration != nil {
		err := s.send(ctx, msg, integration.GetSlackWebhook().Webhook)
		if err != nil {
			return err
		} else {
			s.logger.Info("sent slack message to slack channel", zap.String("channel", integration.GetSlackWebhook().Channel))
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
