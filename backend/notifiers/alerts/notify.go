package alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shank318/doota/datastore"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type LeadSummary struct {
	OrgID              string
	UserID             string
	ProjectName        string
	TotalPostsAnalysed uint32
	TotalCommentsSent  uint32
	DailyCount         uint32
}

type AlertNotifier interface {
	SendLeadsSummary(ctx context.Context, summary LeadSummary) error
}

type SlackNotifier struct {
	Client *http.Client
	db     datastore.Repository
	logger *zap.Logger
}

func NewSlackNotifier(db datastore.Repository, logger *zap.Logger) AlertNotifier {
	return &SlackNotifier{
		db:     db,
		logger: logger,
		Client: &http.Client{Timeout: 10 * time.Second},
	}
}

const redoraChannel = "https://hooks.slack.com/services/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"

func (s *SlackNotifier) SendLeadsSummary(ctx context.Context, summary LeadSummary) error {
	integration, err := s.db.GetIntegrationByOrgAndType(ctx, summary.OrgID, models.IntegrationTypeSLACKWEBHOOK)
	if err != nil && errors.Is(err, datastore.NotFound) {
		s.logger.Info("no integration configured for alerts, skipped")
	}

	leadsURL := "https://app.redoraai.com/dashboard/leads"

	msg := fmt.Sprintf(
		"*📊 Daily Lead Summary — RedoraAI*\n"+
			"*Product:* %s\n"+
			"*Posts Analyzed:* %d\n"+
			"*Automated Comments Posted:* %d\n"+
			"*Leads Found:* *%d*\n\n"+
			"🔗 <%s|View all leads in your dashboard>",
		summary.ProjectName,
		summary.TotalPostsAnalysed,
		summary.TotalCommentsSent,
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

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("send Slack message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("Slack returned non-2xx status: %d", resp.StatusCode)
	}
	return nil
}
