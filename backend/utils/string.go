package utils

import (
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// HumanizeString humanizes a string by converting it to title case and replacing underscores with spaces
func HumanizeString(in string) string {
	return cases.Title(language.AmericanEnglish).String(strings.ReplaceAll(strings.ToLower(in), "_", " "))
}

// TruncateString truncates a string to a specified length
func TruncateString(in string, length int) string {
	if len(in) > length {
		return in[:length] + "..."
	}
	return in
}

// CleanUTF8String removes invalid UTF-8 characters from a string
func CleanUTF8String(in string) string {
	return strings.ToValidUTF8(in, "")
}

// removeExtraSpaces removes extra spaces from a string, it will convert multiple spaces into a single space
func RemoveExtraSpaces(s string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(s), " "))
}
func EString(str string) *string {
	if str == "" {
		return nil
	}
	return &str
}

func IsEmpty(s *string) bool {
	return s == nil || *s == ""
}

// RemoveDuplicateStrings - remove duplicates case-insensitive
func RemoveDuplicateStrings(items []string) []string {
	seen := make(map[string]struct{})
	var uniqueItems []string
	for _, item := range items {
		lowerItem := strings.ToLower(item)
		if _, exists := seen[lowerItem]; !exists {
			uniqueItems = append(uniqueItems, item)
			seen[lowerItem] = struct{}{}
		}
	}
	return uniqueItems
}

func CleanSubredditName(input string) string {
	input = strings.ToLower(input)
	input = strings.TrimPrefix(input, "/r/")
	input = strings.TrimPrefix(input, "r/")
	return input
}

func SanitizeKeyword(input string) string {
	// Trim leading/trailing whitespace
	input = strings.TrimSpace(input)

	// Replace all sequences of whitespace (spaces, tabs, newlines) with a single space
	whitespaceNormalizer := regexp.MustCompile(`\s+`)
	input = whitespaceNormalizer.ReplaceAllString(input, " ")

	// Remove all non-alphanumeric characters except spaces
	safeChars := regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
	input = safeChars.ReplaceAllString(input, "")

	// Convert to lowercase (optional depending on case sensitivity)
	input = strings.ToLower(input)

	return input
}
