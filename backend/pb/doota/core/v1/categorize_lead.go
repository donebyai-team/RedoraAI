package pbcore

import "github.com/shank318/doota/models"

type OutcomeTag string

const (
	ProbableLead      OutcomeTag = "PROBABLE_LEAD"
	BestForEngagement OutcomeTag = "BEST_FOR_ENGAGEMENT"
)

// mapping of OutcomeTags to related PostIntents
var outcomeIntentMap = map[OutcomeTag][]models.PostIntent{
	ProbableLead: {
		models.PostIntentSEEKINGRECOMMENDATIONS,
		models.PostIntentEXPRESSINGPAIN,
		models.PostIntentASKINGFORSOLUTIONS,
		models.PostIntentSHARINGRECOMMENDATION,
		models.PostIntentCOMPETITORMENTION,
	},
	BestForEngagement: {
		models.PostIntentEXPRESSINGGOAL,
		models.PostIntentBUILDINGINPUBLIC,
		models.PostIntentASKINGFORFEEDBACK,
		models.PostIntentDESCRIBINGCURRENTSTACK,
		models.PostIntentGENERALDISCUSSION,
	},
}

// CategorizePost maps post intents to their corresponding high-level outcome tags
func CategorizePost(intents []models.PostIntent) []OutcomeTag {
	if containsUnknown(intents) {
		return []OutcomeTag{}
	}

	tagSet := make(map[OutcomeTag]bool)
	for tag, relatedIntents := range outcomeIntentMap {
		for _, intent := range intents {
			if containsIntent(relatedIntents, intent) {
				tagSet[tag] = true
			}
		}
	}

	var tags []OutcomeTag
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags
}

// GetPostIntentsForOutcome returns the PostIntents associated with a given OutcomeTag
func GetPostIntentsForOutcome(tag OutcomeTag) []models.PostIntent {
	return outcomeIntentMap[tag]
}

// containsUnknown returns true if any intent is UNKNOWN
func containsUnknown(intents []models.PostIntent) bool {
	for _, intent := range intents {
		if intent == models.PostIntentUNKNOWN {
			return true
		}
	}
	return false
}

// containsIntent checks if a slice of PostIntent contains a specific intent
func containsIntent(intents []models.PostIntent, target models.PostIntent) bool {
	for _, intent := range intents {
		if intent == target {
			return true
		}
	}
	return false
}
