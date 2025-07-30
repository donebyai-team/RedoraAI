package reddit

import (
	"github.com/shank318/doota/models"
	"go.uber.org/zap"
	"math/rand"
	"strings"
)

type IntegrationSelectionStrategy func([]*models.Integration) []*models.Integration

func RandomStrategy(integrations []*models.Integration) []*models.Integration {
	shuffled := make([]*models.Integration, len(integrations))
	copy(shuffled, integrations)
	rand.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}

func PreferSpecificAccountStrategy(refID string) IntegrationSelectionStrategy {
	return func(integrations []*models.Integration) []*models.Integration {
		var prioritized *models.Integration
		var rest []*models.Integration

		for _, i := range integrations {
			if i.ReferenceID != nil && strings.EqualFold(*i.ReferenceID, refID) {
				prioritized = i
			} else {
				rest = append(rest, i)
			}
		}

		rand.Shuffle(len(rest), func(i, j int) {
			rest[i], rest[j] = rest[j], rest[i]
		})

		if prioritized != nil {
			return append([]*models.Integration{prioritized}, rest...)
		}
		return rest
	}
}

func MostQualifiedAccountStrategy(logger *zap.Logger) IntegrationSelectionStrategy {
	return func(integrations []*models.Integration) []*models.Integration {
		var oldAccounts, recentAccounts []*models.Integration

		for _, integ := range integrations {
			if integ.Type != models.IntegrationTypeREDDIT {
				continue // skip non-Reddit types
			}
			if integ.ReferenceID == nil {
				logger.Error("reddit reference id is nil", zap.String("integration_id", integ.ID))
				continue
			}

			cfg := integ.GetRedditConfig()
			if cfg.IsUserOldEnough(2) {
				oldAccounts = append(oldAccounts, integ)
			} else {
				recentAccounts = append(recentAccounts, integ)
			}
		}

		var candidates []*models.Integration
		if len(oldAccounts) > 0 {
			candidates = oldAccounts
		} else {
			candidates = recentAccounts
		}

		if len(candidates) == 0 {
			return nil
		}

		rand.Shuffle(len(candidates), func(i, j int) {
			candidates[i], candidates[j] = candidates[j], candidates[i]
		})

		return candidates
	}
}
