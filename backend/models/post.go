package models

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"time"
)

//go:generate go-enum -f=$GOFILE

// ENUM(CREATED, SENT, FAILED, SCHEDULED)
type PostStatus string

type Post struct {
	ID          string       `db:"id"`
	ProjectID   string       `db:"project_id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	SourceID    string       `db:"source_id"`
	ReferenceID *string      `db:"reference_id"`
	PostID      *string      `db:"post_id"`
	Status      PostStatus   `db:"status"`
	Reason      string       `db:"reason"`
	ScheduleAt  *time.Time   `db:"schedule_at"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   *time.Time   `db:"updated_at"`
	DeletedAt   *time.Time   `db:"deleted_at"`
	Metadata    PostMetadata `db:"metadata"`
}

type AugmentedPost struct {
	ID          string       `db:"id"`
	ProjectID   string       `db:"project_id"`
	Title       string       `db:"title"`
	Description string       `db:"description"`
	SourceID    string       `db:"source_id"`
	Source      Source       `db:"source"`
	Status      PostStatus   `db:"status"`
	Reason      string       `db:"reason"`
	ReferenceID *string      `db:"reference_id"` // id of the insight
	PostID      *string      `db:"post_id"`
	ScheduleAt  *time.Time   `db:"schedule_at"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   *time.Time   `db:"updated_at"`
	DeletedAt   *time.Time   `db:"deleted_at"`
	Metadata    PostMetadata `db:"metadata"`
}

type PostMetadata struct {
	Author           string                    `json:"author"`
	Settings         PostSettings              `json:"settings"`
	History          []PostRegenerationHistory `json:"history"`
	PostRequirements PostRequirements          `json:"post_requirements"`
	Flairs           []Flair                   `json:"flairs"`
}

type PostRegenerationHistory struct {
	PostSettings PostSettings `json:"post_settings"`
	Title        string       `db:"title"`
	Description  string       `db:"description"`
}

func (b PostMetadata) Value() (driver.Value, error) {
	return valueAsJSON(b, "post metadata")
}

func (b *PostMetadata) Scan(value interface{}) error {
	return scanFromJSON(value, b, "post metadata")
}

type PostSettings struct {
	Topic       string  `json:"topic"`
	Context     string  `json:"context"`
	Goal        string  `json:"goal"`
	Tone        string  `json:"tone"`
	ReferenceID *string `json:"reference_id"`
	FlairID     *string `json:"flair_id"`
}

type PostRequirements struct {
	TitleRegexes            []string `json:"title_regexes"`
	BodyBlacklistedStrings  []string `json:"body_blacklisted_strings"`
	TitleBlacklistedStrings []string `json:"title_blacklisted_strings"`
	BodyTextMaxLength       *int     `json:"body_text_max_length"`
	TitleRequiredStrings    []string `json:"title_required_strings"`
	GuidelinesText          string   `json:"guidelines_text"`
	DomainBlacklist         []string `json:"domain_blacklist"`
	DomainWhitelist         []string `json:"domain_whitelist"`
	TitleTextMaxLength      *int     `json:"title_text_max_length"`
	BodyRestrictionPolicy   string   `json:"body_restriction_policy"`
	LinkRestrictionPolicy   string   `json:"link_restriction_policy"`
	GuidelinesDisplayPolicy string   `json:"guidelines_display_policy"`
	BodyRequiredStrings     []string `json:"body_required_strings"`
	TitleTextMinLength      *int     `json:"title_text_min_length"`
	IsFlairRequired         bool     `json:"is_flair_required"`
	BodyRegexes             []string `json:"body_regexes"`
	BodyTextMinLength       *int     `json:"body_text_min_length"`
}

type Flair struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Text    string `json:"text"`
	ModOnly bool   `json:"mod_only"`
}

func (rules PostRequirements) Validate(post Post) []string {
	var errs []string
	title := post.Title
	body := post.Description

	if rules.IsFlairRequired && strings.TrimSpace(*post.Metadata.Settings.FlairID) == "" {
		errs = append(errs, "flair is required")
	}

	if len(title) == 0 {
		errs = append(errs, "title can not be empty")
	}

	if len(body) == 0 {
		errs = append(errs, "body can not be empty")
	}

	if rules.TitleTextMinLength != nil && len(title) < *rules.TitleTextMinLength {
		errs = append(errs, fmt.Sprintf("title is too short (min %d characters)", *rules.TitleTextMinLength))
	}

	if rules.TitleTextMaxLength != nil && len(title) > *rules.TitleTextMaxLength {
		errs = append(errs, fmt.Sprintf("title is too long (max %d characters)", *rules.TitleTextMaxLength))
	}

	if rules.BodyTextMinLength != nil && len(body) < *rules.BodyTextMinLength {
		errs = append(errs, fmt.Sprintf("body is too short (min %d characters)", *rules.BodyTextMinLength))
	}
	if rules.BodyTextMaxLength != nil && len(body) > *rules.BodyTextMaxLength {
		errs = append(errs, fmt.Sprintf("body is too long (max %d characters)", *rules.BodyTextMaxLength))
	}

	for _, required := range rules.TitleRequiredStrings {
		if !strings.Contains(strings.ToLower(title), strings.ToLower(required)) {
			errs = append(errs, fmt.Sprintf("title must include: %q", required))
		}
	}
	for _, required := range rules.BodyRequiredStrings {
		if !strings.Contains(strings.ToLower(body), strings.ToLower(required)) {
			errs = append(errs, fmt.Sprintf("body must include: %q", required))
		}
	}

	for _, forbidden := range rules.TitleBlacklistedStrings {
		if strings.Contains(strings.ToLower(title), strings.ToLower(forbidden)) {
			errs = append(errs, fmt.Sprintf("title contains blacklisted phrase: %q", forbidden))
		}
	}
	for _, forbidden := range rules.BodyBlacklistedStrings {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			errs = append(errs, fmt.Sprintf("body contains blacklisted phrase: %q", forbidden))
		}
	}

	for _, pattern := range rules.TitleRegexes {
		matched, err := regexp.MatchString(pattern, title)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid title regex %q: %v", pattern, err))
		} else if !matched {
			errs = append(errs, fmt.Sprintf("title does not match required pattern: %q", pattern))
		}
	}
	for _, pattern := range rules.BodyRegexes {
		matched, err := regexp.MatchString(pattern, body)
		if err != nil {
			errs = append(errs, fmt.Sprintf("invalid body regex %q: %v", pattern, err))
		} else if !matched {
			errs = append(errs, fmt.Sprintf("body does not match required pattern: %q", pattern))
		}
	}

	if rules.BodyRestrictionPolicy == "required" && strings.TrimSpace(body) == "" {
		errs = append(errs, "body is required by subreddit policy")
	}

	return errs
}

// ToRules returns all human-readable rules for display or LLM prompt
func (rules PostRequirements) ToRules() []string {
	var result []string

	// Flair requirement
	if rules.IsFlairRequired {
		result = append(result, "Flair is required.")
	}

	// Title length
	if rules.TitleTextMinLength != nil && *rules.TitleTextMinLength != 0 {
		result = append(result, fmt.Sprintf("Title must be at least %d characters long.", *rules.TitleTextMinLength))
	}

	if rules.TitleTextMaxLength != nil && *rules.TitleTextMaxLength != 0 {
		result = append(result, fmt.Sprintf("Title cannot exceed %d characters.", *rules.TitleTextMaxLength))
	}

	// Body length
	if rules.BodyTextMinLength != nil && *rules.BodyTextMinLength != 0 {
		result = append(result, fmt.Sprintf("Body must be at least %d characters long.", *rules.BodyTextMinLength))
	}
	if rules.BodyTextMaxLength != nil && *rules.BodyTextMaxLength != 0 {
		result = append(result, fmt.Sprintf("Body cannot exceed %d characters.", *rules.BodyTextMaxLength))
	}

	// Required phrases
	if len(rules.TitleRequiredStrings) > 0 {
		result = append(result, fmt.Sprintf("Title must contain all of the phrases: %q", rules.TitleRequiredStrings))
	}
	if len(rules.BodyRequiredStrings) > 0 {
		result = append(result, fmt.Sprintf("Body must contain all of the phrases: %q", rules.BodyRequiredStrings))
	}

	// Blacklisted phrases
	if len(rules.TitleBlacklistedStrings) > 0 {
		result = append(result, fmt.Sprintf("Title must NOT contain any of the phrases: %q", rules.TitleBlacklistedStrings))
	}
	if len(rules.BodyBlacklistedStrings) > 0 {
		result = append(result, fmt.Sprintf("Body must NOT contain any of the phrases: %q", rules.BodyBlacklistedStrings))
	}

	// Regex requirements
	if len(rules.TitleRegexes) > 0 {
		result = append(result, fmt.Sprintf("Title must match all of the regex patterns: %q", rules.TitleRegexes))
	}
	if len(rules.BodyRegexes) > 0 {
		result = append(result, fmt.Sprintf("Body must match all of the regex patterns: %q", rules.BodyRegexes))
	}

	// Domain restrictions
	if len(rules.DomainWhitelist) > 0 {
		result = append(result, fmt.Sprintf("Links must be from one of these domains: %q", rules.DomainWhitelist))
	}
	if len(rules.DomainBlacklist) > 0 {
		result = append(result, fmt.Sprintf("Links must NOT be from any of these domains: %q", rules.DomainBlacklist))
	}

	// Policies
	if rules.BodyRestrictionPolicy == "required" {
		result = append(result, "Body is required.")
	}
	if rules.LinkRestrictionPolicy != "" && rules.LinkRestrictionPolicy != "none" {
		result = append(result, fmt.Sprintf("Link restriction policy: %s", rules.LinkRestrictionPolicy))
	}
	if rules.GuidelinesText != "" {
		result = append(result, fmt.Sprintf("Note: %s", rules.GuidelinesText))
	}

	return result
}
