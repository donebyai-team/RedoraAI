package alerts

import (
	"context"
	"fmt"
	"github.com/shank318/doota/datastore"
	"go.uber.org/zap"
)
import "github.com/resend/resend-go/v2"

type ResendNotifier struct {
	Client *resend.Client
	db     datastore.Repository
}

func NewResendNotifier(apiKey string, db datastore.Repository, logger *zap.Logger) *ResendNotifier {
	return &ResendNotifier{Client: resend.NewClient(apiKey), db: db}
}

func (r ResendNotifier) SendLeadsSummary(ctx context.Context, summary LeadSummary) error {
	users, err := r.db.GetUsersByOrgID(ctx, summary.OrgID)
	if err != nil {
		return err
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
		    <h2>Daily Lead Summary — <strong>RedoraAI</strong></h2>
		    <p><strong>Product:</strong> %s</p>
		    <p><strong>Posts Analyzed:</strong> %d</p>
		    <p><strong>Automated Comments Posted:</strong> %d</p>
		    <p><strong>Leads Found:</strong> <strong>%d</strong></p>
		    <p>🔗 <a href="%s">View all leads in your dashboard</a></p>
		    <hr>
		    <footer style="font-size: 12px; color: #888;">
		      <p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
		      <p>Need help? <a href="mailto:shashank@donebyai.team">shashank@donebyai.team</a></p>
		    </footer>
		  </div>
		</body>
		</html>
	`, summary.ProjectName, summary.TotalPostsAnalysed, summary.TotalCommentsSent, summary.DailyCount, "https://app.redoraai.com/dashboard/leads")

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <leads@alerts.redoraai.com>",
		To:      to,
		Subject: "📊Daily Lead Summary",
		Html:    htmlBody,
	}

	_, err = r.Client.Emails.Send(params)
	return err
}
