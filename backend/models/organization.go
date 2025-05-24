package models

import (
	"database/sql/driver"
	"time"

	"go.uber.org/zap/zapcore"
)

type Organization struct {
	ID           string                   `db:"id"`
	Name         string                   `db:"name"`
	FeatureFlags OrganizationFeatureFlags `db:"feature_flags"`
	CreatedAt    time.Time                `db:"created_at"`
	UpdatedAt    *time.Time               `db:"updated_at"`
}

func (o *Organization) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("id", o.ID)
	encoder.AddString("name", o.Name)
	return nil
}

// TODO: Move it to a better place
type OrganizationFeatureFlags struct {
	EnableAutoDM     bool    `json:"enable_auto_dm"`
	RelevancyScoreDM float64 `json:"relevancy_score_dm"`
	MaxDMsPerDay     int64   `json:"max_dms_per_day"` // specified by user, max can be based on the plan subscribed

	EnableAutoComment     bool    `json:"enable_auto_comment"`
	RelevancyScoreComment float64 `json:"relevancy_score_comment"`
	MaxCommentsPerDay     int64   `json:"max_comments_per_day"` // specified by user, max can be based on the plan subscribed

	CommentLLMModel   LLMModel      `json:"comment_llm_model"`
	DMLLMModel        LLMModel      `json:"dm_llm_model"`
	RelevancyLLMModel LLMModel      `json:"relevancy_llm_model"`
	Subscription      *Subscription `json:"subscription"` // storing here for faster access
	Activities        []OrgActivity `json:"activities"`
}

func (f OrganizationFeatureFlags) GetSubscriptionPlanMetadata() SubscriptionPlanMetadata {
	//if f.Subscription == nil {
	//	return RedoraPlans[SubscriptionPlanTypeFREE].Metadata
	//}
	return RedoraPlans[SubscriptionPlanTypeFREE].Metadata
}

// Defined by redora global or max allowed by plan, whichever is higher
func (f OrganizationFeatureFlags) GetMaxLeadsPerDay() int64 {
	return f.GetSubscriptionPlanMetadata().RelevantPosts.PerDay
}

// Defined by user or max allowed by plan, whichever is higher
func (f OrganizationFeatureFlags) GetMaxDMsPerDay() int64 {
	if f.MaxDMsPerDay == 0 {
		return f.GetSubscriptionPlanMetadata().DMs.PerDay
	}
	return f.MaxDMsPerDay
}

// Defined by user or max allowed by plan, whichever is higher
func (f OrganizationFeatureFlags) GetMaxCommentsPerDay() int64 {
	if f.MaxCommentsPerDay == 0 {
		return f.GetSubscriptionPlanMetadata().Comments.PerDay
	}
	return f.MaxCommentsPerDay
}

const defaultMinRelevancyScoreForAutomatedCommentsAndDM = 90

// Defined by user or max allowed by plan, whichever is higher
func (f OrganizationFeatureFlags) GetRelevancyScoreDM() float64 {
	if f.RelevancyScoreDM == 0 {
		return defaultMinRelevancyScoreForAutomatedCommentsAndDM
	}
	return f.RelevancyScoreDM
}

// Defined by user or max allowed by plan, whichever is higher
func (f OrganizationFeatureFlags) GetRelevancyScoreComment() float64 {
	if f.RelevancyScoreComment == 0 {
		return defaultMinRelevancyScoreForAutomatedCommentsAndDM
	}
	return f.RelevancyScoreComment
}

type OrgActivity struct {
	ActivityType OrgActivityType `json:"activity_type"`
	CreatedAt    time.Time       `json:"created_at"`
}

//go:generate go-enum -f=$GOFILE

// ENUM(COMMENT_DISABLED_ACCOUNT_AGE_NEW, COMMENT_DISABLED_LOW_KARMA, COMMENT_ENABLED_WARMED_UP)
type OrgActivityType string

func (b OrganizationFeatureFlags) IsSubscriptionExpired() bool {
	if b.Subscription == nil {
		return false
	}
	return b.Subscription.IsExpired()
}

func (b OrganizationFeatureFlags) Value() (driver.Value, error) {
	return valueAsJSON(b, "organization feature flags")
}

func (b *OrganizationFeatureFlags) Scan(value interface{}) error {
	return scanFromJSON(value, b, "organization feature flags")
}
