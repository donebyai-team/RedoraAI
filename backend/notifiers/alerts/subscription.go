package alerts

import (
	"context"
	"fmt"
	"github.com/resend/resend-go/v2"
	"go.uber.org/zap"
)

func (s *SlackNotifier) SendSubscriptionCreatedEmail(ctx context.Context, orgID string) {
	s.sendSubscriptionEmail(ctx, orgID, "🎉 You're now subscribed to RedoraAI!", `
	    <h2>Thanks for subscribing to <strong>RedoraAI</strong> 🎉</h2>
	    <p>You’ve just upgraded your plan and unlocked premium features to supercharge your lead generation from Reddit.</p>
	    
	    <h3>Here’s what you can do now:</h3>
	    <ul>
	      <li>🚀 Monitor more keywords and subreddits</li>
	      <li>🤖 Use Redora Copilot to automate replies and DMs</li>
	      <li>📊 Access lead analytics and campaign insights</li>
	    </ul>

	    <p>🔗 <a href="https://app.redoraai.com/dashboard" style="color: #3366cc;">Go to Dashboard</a></p> 		
	`)
}

func (s *SlackNotifier) SendSubscriptionRenewedEmail(ctx context.Context, orgID string) {
	s.sendSubscriptionEmail(ctx, orgID, "🔁 Your RedoraAI subscription has been renewed", `
	    <h2>Your <strong>RedoraAI</strong> subscription was renewed ✅</h2>
	    <p>Your payment was successful and your access to premium features continues without interruption.</p>
	    
	    <p>No action is needed. If you have any questions or feedback, just reply to this email.</p>

	    <p>🔗 <a href="https://app.redoraai.com/billing" style="color: #3366cc;">Manage Subscription</a></p>
	`)
}

func (s *SlackNotifier) SendSubscriptionCancelledEmail(ctx context.Context, orgID string) {
	s.sendSubscriptionEmail(ctx, orgID, "❌ Your RedoraAI subscription has been cancelled", `
	    <h2>Your <strong>RedoraAI</strong> subscription is now cancelled</h2>
	    <p>Your access to premium features will remain active until the end of your current billing cycle.</p>

	    <p>If you changed your mind, you can re-subscribe at any time from the billing page.</p>

	    <p>🔗 <a href="https://app.redoraai.com/billing" style="color: #3366cc;">Manage Subscription</a></p>

	    <p>Thanks for trying RedoraAI — we’d love to hear your feedback or help with anything.</p>
	    <p>Just reply to this email or reach us at <a href="mailto:support@redoraai.com">support@redoraai.com</a>.</p>
	`)
}

func (s *SlackNotifier) sendSubscriptionEmail(ctx context.Context, orgID, subject, bodyHTML string) {
	users, err := s.db.GetUsersByOrgID(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get users for subscription email", zap.Error(err))
		return
	}

	if len(users) == 0 || len(users) > 1 {
		return
	}

	htmlBody := fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<body style="font-family: Arial, sans-serif; background-color: #f7f9fc; padding: 20px;">
	  <div style="max-width: 600px; margin: auto; background-color: #ffffff; padding: 30px; border-radius: 8px;">
	    %s
	    <hr>
	    <footer style="font-size: 12px; color: #888;">
			<p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
			<p>Need help? <a href="mailto:adarsh@redoraai.com">adarsh@redoraai.com</a></p>
		</footer>
	  </div>
	</body>
	</html>
	`, bodyHTML)

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <welcome@alerts.redoraai.com>",
		To:      []string{users[0].Email},
		Cc:      []string{"shashank@donebyai.team", "adarsh@redoraai.com"},
		Subject: subject,
		Html:    htmlBody,
	}

	_, err = s.ResendClient.Emails.Send(params)
	if err != nil {
		s.logger.Error("failed to send subscription email", zap.Error(err))
	}
}
