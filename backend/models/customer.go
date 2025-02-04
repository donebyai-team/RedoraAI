package models

import "time"

type Customer struct {
	ID        string     `db:"id"`
	FirstName string     `db:"first_name"`
	LastName  string     `db:"last_name"`
	Phone     string     `db:"phone"`
	OrgID     string     `db:"organization_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

//go:generate go-enum -f=$GOFILE

// ENUM(ASSISTANT_ERROR, ASSISTANT_ENDED, ASSISTANT_FORWARDED, CUSTOMER_BUSY, CUSTOMER_ENDED, MAX_CALL_LIMIT_REACHED)
type CallEndedReason string

// ENUM(UNKNOWN, CREATED, QUEUED, IN_PROGRESS, ENDED, FORWARDED)
type CallStatus string

// ENUM(CREATED, PENDING, PARTIALLY_PAID, PAID, CLOSED)
type CustomerCaseStatus string

type CustomerCase struct {
	ID              string             `db:"id"`
	PromptType      string             `db:"prompt_type"`
	OrgID           string             `db:"organization_id"`
	CustomerID      string             `db:"customer_id"`
	DueDate         time.Time          `db:"due_date"`
	Status          CustomerCaseStatus `db:"status"`
	Summary         string             `db:"summary"`
	LastCallStatus  CallStatus         `db:"last_call_status"`
	NextScheduledAt *time.Time         `db:"next_scheduled_at"`
	CreatedAt       time.Time          `db:"created_at"`
	UpdatedAt       *time.Time         `db:"updated_at"`
}

type Conversation struct {
	ID              string          `db:"id"`
	CustomerCaseID  string          `db:"customer_case_id"`
	FromPhone       string          `db:"from_phone"`
	CallStatus      CallStatus      `db:"call_status"`
	NextScheduledAt *time.Time      `db:"next_scheduled_at"`
	Summary         string          `db:"summary"`
	Provider        IntegrationType `db:"provider"`
	ExternalID      string          `db:"external_id"`
	CallDuration    uint32          `db:"call_duration"` // in milliseconds
	RecordingURL    string          `db:"recording_url"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       *time.Time      `db:"updated_at"`
	CallEndedReason CallEndedReason `db:"call_ended_reason"`

	// Non-db
	CustomerCaseStatus CustomerCaseStatus
	CaseSummary        string
}

type AugmentedCustomerCase struct {
	CustomerCase  *CustomerCase
	Customer      *Customer
	Conversations []*Conversation
}

type AugmentedConversation struct {
	CustomerCase *CustomerCase
	Customer     *Customer
	Conversation *Conversation
}
