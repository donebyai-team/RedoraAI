package models

import (
	"database/sql/driver"
	"time"
)

type LLMModel string

type LLMModelUsage struct {
	Model        LLMModel `json:"model"`
	Usage        int      `json:"usage"`
	RateLimitLow bool
}

type RuleEvaluationResult struct {
	ProductMentionAllowed bool     `json:"product_mention_allowed"` // true if it's okay to mention product in comments
	RiskLevel             string   `json:"risk_level"`              // e.g., "low", "medium", "high"
	ImportantGuidelines   []string `json:"important_guidelines"`    // key points to keep in mind while generating comments
	ChainOfThought        string   `json:"chain_of_thought"`        // short explanation referencing rules that influenced the decision
}

type RedditPostRelevanceResponse struct {
	ChainOfThoughtIsRelevant       string       `json:"chain_of_thought"`
	IsRelevantConfidenceScore      float64      `json:"relevant_confidence_score"`
	SuggestedDM                    string       `json:"suggested_dm"`
	Intents                        []PostIntent `json:"intents"`
	ChainOfThoughtSuggestedDM      string       `json:"chain_of_thought_suggested_dm"`
	SuggestedComment               string       `json:"suggested_comment"`
	ChainOfThoughtSuggestedComment string       `json:"chain_of_thought_suggested_comment"`
	AppliedRules                   []string     `json:"applied_rules"`
}

type CaseDecisionResponse struct {
	ChainOfThoughtCaseStatus           string             `json:"chain_of_thought_case_status"`
	CaseStatusReason                   CustomerCaseReason `json:"case_status"`
	CaseStatusConfidenceScore          float64            `json:"case_status_confidence_score"`
	NextCallScheduledAt                string             `json:"next_call_scheduled_at"`
	ChainOfThoughtNextCallScheduledAt  string             `json:"chain_of_thought_next_call_scheduled_at"`
	NextCallScheduledAtConfidenceScore float64            `json:"next_call_scheduled_at_confidence_score"`
	NextCallScheduledAtTime            *time.Time         `json:"-"`
}

func (b CaseDecisionResponse) Value() (driver.Value, error) {
	if b.IsEmpty() {
		return "{}", nil // Return empty JSON object if struct is empty
	}
	return valueAsJSON(b, "ai_decision")
}

func (b *CaseDecisionResponse) Scan(value interface{}) error {
	return scanFromJSON(value, b, "ai_decision")
}

func (b CaseDecisionResponse) IsEmpty() bool {
	return b.ChainOfThoughtCaseStatus == "" &&
		b.CaseStatusReason == "" &&
		b.CaseStatusConfidenceScore == 0 &&
		b.NextCallScheduledAt == "" &&
		b.ChainOfThoughtNextCallScheduledAt == "" &&
		b.NextCallScheduledAtConfidenceScore == 0
}
