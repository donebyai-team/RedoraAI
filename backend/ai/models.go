package ai

import (
	"github.com/shank318/doota/models"
	"time"
)

type CaseDecisionResponse struct {
	ChainOfThoughtCaseStatus           string                    `json:"chain_of_thought_case_status"`
	CaseStatusReason                   models.CustomerCaseReason `json:"case_status"`
	CaseStatusConfidenceScore          float64                   `json:"case_status_confidence_score"`
	NextCallScheduledAt                string                    `json:"next_call_scheduled_at"`
	ChainOfThoughtNextCallScheduledAt  string                    `json:"chain_of_thought_next_call_scheduled_at"`
	NextCallScheduledAtConfidenceScore float64                   `json:"next_call_scheduled_at_confidence_score"`
	Summary                            string                    `json:"summary"`
	NextCallScheduledAtTime            *time.Time                `json:"-"`
}
