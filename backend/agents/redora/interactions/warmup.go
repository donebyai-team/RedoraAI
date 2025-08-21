package interactions

import (
	"context"
	"fmt"
	"github.com/shank318/doota/browser_automation"
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

var accountsToWarmUp = []string{""}

func (r redditInteractions) WarmUpAccounts(ctx context.Context) error {
	now := time.Now()

	for _, account := range accountsToWarmUp {
		integrations, err := r.db.GetIntegrationsByReferenceId(ctx, account)
		if err != nil {
			r.logger.Error("failed to get integration",
				zap.Error(err),
				zap.String("account", account))
			continue
		}

		for _, integration := range integrations {
			if integration.Type != models.IntegrationTypeREDDITDMLOGIN {
				continue
			}

			// Skip if warmup was done less than (24h + random jitter) ago
			last := integration.Metadata.WarmUpData.LastWarmUpAt
			// add randomness: between 24h and 26h
			jitter := time.Duration(24+rand.Intn(3)) * time.Hour
			if !last.IsZero() && now.Sub(last) < jitter {
				continue
			}

			config := integration.GetRedditDMLoginConfig()
			r.logger.Info("warming up account",
				zap.String("account", account),
				zap.String("referenceID", *integration.ReferenceID))

			err = r.redditBrowserAutomation.DailyWarmup(ctx, browser_automation.DailyWarmParams{
				ID:          *integration.ReferenceID,
				Cookies:     config.Cookies,
				CountryCode: config.Alpha2CountryCode,
			})

			if err != nil {
				r.logger.Error("failed to warm up account",
					zap.Error(err),
					zap.String("account", account))

				go r.alertNotifier.SendInteractionError(
					context.Background(),
					*integration.ReferenceID,
					fmt.Errorf("failed to warm up account: %w", err),
				)
				continue
			}

			// update warmup metadata
			integration.Metadata.WarmUpData.LastWarmUpAt = now
			integration.Metadata.WarmUpData.Count++
			if _, err = r.db.UpsertIntegration(ctx, integration); err != nil {
				return fmt.Errorf("failed to update integration: %w", err)
			}
		}
	}

	return nil
}
