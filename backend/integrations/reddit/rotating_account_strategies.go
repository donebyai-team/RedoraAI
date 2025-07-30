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
		var (
			activeRedditIntegrations []*models.Integration
			activeDMIntegrations     []*models.Integration
		)

		// Split into Reddit and DM
		for _, integ := range integrations {
			switch integ.Type {
			case models.IntegrationTypeREDDIT:
				activeRedditIntegrations = append(activeRedditIntegrations, integ)
			case models.IntegrationTypeREDDITDMLOGIN:
				activeDMIntegrations = append(activeDMIntegrations, integ)
			}
		}

		// Index DM by ReferenceID
		dmSet := make(map[string]struct{})
		for _, dm := range activeDMIntegrations {
			if dm.ReferenceID == nil {
				logger.Error("reddit DM reference id is nil", zap.String("integration_id", dm.ID))
				continue
			}
			dmSet[strings.ToLower(*dm.ReferenceID)] = struct{}{}
		}

		// Categorize Reddit integrations
		var bothTypesOld, bothTypesRecent, onlyRedditOld, onlyRedditRecent []*models.Integration
		for _, redditIntegration := range activeRedditIntegrations {
			if redditIntegration.ReferenceID == nil {
				logger.Error("reddit reference id is nil", zap.String("integration_id", redditIntegration.ID))
				continue
			}
			_, hasDM := dmSet[strings.ToLower(*redditIntegration.ReferenceID)]
			redditConfig := redditIntegration.GetRedditConfig()
			isOld := redditConfig.IsUserOldEnough(2)

			switch {
			case hasDM && isOld:
				bothTypesOld = append(bothTypesOld, redditIntegration)
			case !hasDM && isOld:
				onlyRedditOld = append(onlyRedditOld, redditIntegration)
			case hasDM && !isOld:
				bothTypesRecent = append(bothTypesRecent, redditIntegration)
			default:
				onlyRedditRecent = append(onlyRedditRecent, redditIntegration)
			}
		}

		// old accounts are giving the priority
		// then matching with DM

		var candidates []*models.Integration
		switch {
		case len(bothTypesOld) > 0:
			candidates = bothTypesOld // best case
		case len(onlyRedditOld) > 0:
			candidates = onlyRedditOld
		case len(bothTypesRecent) > 0:
			candidates = bothTypesRecent // should disable comment automation later
		default:
			candidates = onlyRedditRecent // should disable comment automation later
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
