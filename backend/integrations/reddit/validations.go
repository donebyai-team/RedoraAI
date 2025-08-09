package reddit

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shank318/doota/models"
)

func (rules ValidationRules) Validate(post models.Post) []string {
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
func (rules ValidationRules) ToRules() []string {
	var result []string

	// Flair requirement
	if rules.IsFlairRequired {
		result = append(result, "Flair is required.")
	}

	// Title length
	if rules.TitleTextMinLength != nil {
		result = append(result, fmt.Sprintf("Title must be at least %d characters long.", *rules.TitleTextMinLength))
	}

	if rules.TitleTextMaxLength != nil {
		result = append(result, fmt.Sprintf("Title cannot exceed %d characters.", *rules.TitleTextMaxLength))
	}

	// Body length
	if rules.BodyTextMinLength != nil {
		result = append(result, fmt.Sprintf("Body must be at least %d characters long.", *rules.BodyTextMinLength))
	}
	if rules.BodyTextMaxLength != nil {
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
