package interactions

import (
	"context"
	"fmt"
	"github.com/shank318/doota/browser_automation"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
)

var accountsToWarmUp = []string{}

func (r redditInteractions) WarmUpAccounts(ctx context.Context) error {
	for _, account := range accountsToWarmUp {
		integrations, err := r.db.GetIntegrationsByReferenceId(ctx, account)
		if err != nil {
			r.logger.Error("failed to get integration", zap.Error(err), zap.String("account", account))
			continue
		}

		for _, integration := range integrations {
			if integration.Type != models.IntegrationTypeREDDITDMLOGIN {
				continue
			}

			config := integration.GetRedditDMLoginConfig()
			r.logger.Info("warming up account", zap.String("account", account))

			err = r.redditBrowserAutomation.DailyWarmup(ctx, browser_automation.DailyWarmParams{
				ID:          *integration.ReferenceID,
				Cookies:     config.Cookies,
				CountryCode: config.Alpha2CountryCode,
			})

			if err != nil {
				r.logger.Error("failed to warm up account", zap.Error(err), zap.String("account", account))
				go r.alertNotifier.SendInteractionError(context.Background(), *integration.ReferenceID, fmt.Errorf("failed to warm up account: %w", err))
			}
		}

	}

	return nil
}
