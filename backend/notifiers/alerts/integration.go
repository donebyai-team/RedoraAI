package alerts

import (
	"context"
	"fmt"
	"github.com/resend/resend-go/v2"
)

func (s *SlackNotifier) SendIntegrationRevoked(ctx context.Context, orgID string, accountName string, reason string) {
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
					<h2>One of Your Connected Integrations Has Been Revoked</h2>
					<p>We noticed that one of your connected integrations, <strong>%s</strong>, has been revoked from your account.</p>
					<p style="margin: 20px 0; padding: 15px; background-color: #fee2e2; border-left: 4px solid #ef4444; border-radius: 4px;">
						Reason: %s
					</p>
					<p>This means features and automations using this integration will no longer work.  
					Other integrations remain active.  
					To restore functionality for this integration, please reconnect it from your RedoraAI dashboard or contact support.</p>
					<hr>
					<footer style="font-size: 12px; color: #888;">
						<p><strong>RedoraAI</strong> — AI for Intelligent Lead Generation</p>
						<p>Need help or have questions? <a href="mailto:adarsh@redoraai.com">adarsh@redoraai.com</a></p>
					</footer>
				</div>
			</body>
		</html>
	`, accountName, reason)

	params := &resend.SendEmailRequest{
		From:    "RedoraAI <leads@alerts.redoraai.com>",
		To:      to,
		Cc:      []string{"shashank@donebyai.team", "adarsh@redoraai.com"},
		Subject: "⚠️ Integration Revoked",
		Html:    htmlBody,
	}

	_, err = s.ResendClient.Emails.Send(params)
	return
}
