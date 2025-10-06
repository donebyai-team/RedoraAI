package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type UsageLimits struct {
	PerDay int64 `json:"per_day"`
}

type AutomationLimits struct {
	Comment bool `json:"comment"`
	DM      bool `json:"dm"`
}

// ENUM(KEYWORD, SOURCE)
type AddOnType string

type SubscriptionPlanMetadata struct {
	AutomationLimits *AutomationLimits `json:"automation_limits"`
	Comments         UsageLimits       `json:"comments"`
	DMs              UsageLimits       `json:"dms"`
	RelevantPosts    UsageLimits       `json:"relevant_posts"`
	MaxKeywords      int               `json:"max_keywords"`
	MaxSources       int               `json:"max_sources"`
	AddOns           map[AddOnType]int `json:"add_ons"`
}

//go:generate go-enum -f=$GOFILE

// ENUM(FREE, FOUNDER, PRO, STARTER)
type SubscriptionPlanType string

type SubscriptionPlan struct {
	PlanType    SubscriptionPlanType     `json:"plan_type"`
	Description string                   `json:"description"`
	Price       float64                  `json:"price"`
	Interval    int                      `json:"interval"`
	Metadata    SubscriptionPlanMetadata `json:"metadata"`
}

var RedoraPlans = map[SubscriptionPlanType]*SubscriptionPlan{
	SubscriptionPlanTypeFREE: {
		PlanType:    SubscriptionPlanTypeFREE,
		Description: "Free plan with limited usage to try out the platform",
		Price:       0.0,
		Interval:    3, // 3 days
		Metadata: SubscriptionPlanMetadata{
			Comments: UsageLimits{
				PerDay: 25,
			},
			DMs: UsageLimits{
				PerDay: 25,
			},
			RelevantPosts: UsageLimits{
				PerDay: 25,
			},
			MaxSources:  7,
			MaxKeywords: 7,
		},
	},
	SubscriptionPlanTypeSTARTER: {
		PlanType:    SubscriptionPlanTypeSTARTER,
		Description: "Starter plan with limited usage to try out the platform",
		Price:       12.0,
		Interval:    30,
		Metadata: SubscriptionPlanMetadata{
			AutomationLimits: &AutomationLimits{
				Comment: false,
				DM:      false,
			},
			Comments: UsageLimits{
				PerDay: 25,
			},
			DMs: UsageLimits{
				PerDay: 25,
			},
			RelevantPosts: UsageLimits{
				PerDay: 25,
			},
			MaxSources:  5,
			MaxKeywords: 5,
		},
	},
	SubscriptionPlanTypeFOUNDER: {
		PlanType:    SubscriptionPlanTypeFOUNDER,
		Description: "Perfect for individual founders reaching out to niche communities",
		Price:       39.99,
		Interval:    30,
		Metadata: SubscriptionPlanMetadata{
			Comments: UsageLimits{
				PerDay: 25,
			},
			DMs: UsageLimits{
				PerDay: 25,
			},
			RelevantPosts: UsageLimits{
				PerDay: 25,
			},
			MaxSources:  7,
			MaxKeywords: 7,
		},
	},
	SubscriptionPlanTypePRO: {
		PlanType:    SubscriptionPlanTypePRO,
		Description: "Best for marketing agencies managing multiple clients",
		Price:       99.99,
		Interval:    30,
		Metadata: SubscriptionPlanMetadata{
			Comments: UsageLimits{
				PerDay: 50,
			},
			DMs: UsageLimits{
				PerDay: 50,
			},
			RelevantPosts: UsageLimits{
				PerDay: 50,
			},
			MaxSources:  20,
			MaxKeywords: 20,
		},
	},
}

// ENUM(CREATED, CANCELLED, EXPIRED, ACTIVE, FAILED)
type SubscriptionStatus string

type Subscription struct {
	ID             string                   `db:"id" json:"id"`
	OrganizationID string                   `db:"organization_id" json:"organization_id"`
	Amount         float64                  `db:"amount" json:"amount"`
	PlanID         SubscriptionPlanType     `db:"plan_id" json:"plan_id"`
	Status         SubscriptionStatus       `db:"status" json:"status"`
	Metadata       SubscriptionPlanMetadata `db:"metadata" json:"metadata"`
	ExternalID     *string                  `db:"external_id" json:"external_id,omitempty"`
	CreatedAt      time.Time                `db:"created_at" json:"created_at"`
	ExpiresAt      time.Time                `db:"expires_at" json:"expires_at"`
	UpdatedAt      *time.Time               `db:"updated_at" json:"updated_at,omitempty"`
	PaymentLink    string                   `db:"-" json:"payment_link"`
}

func (b SubscriptionPlanMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "subscription metadata")
}

func (b *SubscriptionPlanMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "subscription metadata")
}

func (s *Subscription) MarshalJSON() ([]byte, error) {
	// Check expiration and override status if needed
	if s.IsExpired() {
		s.Status = SubscriptionStatusEXPIRED
	}

	// Define alias to avoid infinite recursion
	type Alias Subscription
	return json.Marshal((*Alias)(s))
}

func (s *Subscription) IsExpired() bool {
	margin := 24 * time.Hour
	return time.Now().After(s.ExpiresAt.Add(margin))
}

func (s *Subscription) GetStatus() SubscriptionStatus {
	if s.IsExpired() {
		return SubscriptionStatusEXPIRED
	}
	return s.Status
}
