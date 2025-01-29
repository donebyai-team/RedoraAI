package models

import "time"

type Customer struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	Phone     string     `db:"phone"`
	OrgID     string     `db:"organization_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

//go:generate go-enum -f=$GOFILE

// ENUM(IN_PROGRESS, QUEUED, AI_ENDED, CUSTOMER_ENDED)
type ConversationStatus string

type CustomerSession struct {
	ID         string             `db:"id"`
	PromptType string             `db:"prompt_type"`
	OrgID      string             `db:"organization_id"`
	CustomerID string             `db:"customer_id"`
	ExternalID string             `db:"external_id"`
	DueDate    time.Time          `db:"due_date"`
	Status     ConversationStatus `db:"status"`
	CreatedAt  time.Time          `db:"created_at"`
	UpdatedAt  *time.Time         `db:"updated_at"`
}

// ENUM(VOICE_MILLIS, VOICE_VAPI)
type Provider string

type Conversation struct {
	ID                string             `db:"id"`
	CustomerSessionID string             `db:"customer_session_id"`
	FromPhone         string             `db:"from_phone"`
	Status            ConversationStatus `db:"status"`
	Summary           string             `db:"summary"`
	Provider          Provider           `db:"provider"`
	ExternalID        string             `db:"external_id"`
	CallDuration      uint32             `db:"call_duration"` // in milliseconds
	RecordingURL      string             `db:"recording_url"`
	CreatedAt         time.Time          `db:"created_at"`
	UpdatedAt         *time.Time         `db:"updated_at"`
}
