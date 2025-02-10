package models

import (
	"database/sql/driver"
	"time"
)

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
	return valueAsJSON(b, "ai_decision")
}

func (b *CaseDecisionResponse) Scan(value interface{}) error {
	return scanFromJSON(value, b, "ai_decision")
}
