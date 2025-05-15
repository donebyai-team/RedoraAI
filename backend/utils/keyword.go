package utils

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// List of stop words to reject
var stopWords = map[string]bool{
	"the": true, "and": true, "a": true, "an": true,
	"of": true, "in": true, "on": true, "for": true,
	"at": true, "to": true, "is": true, "it": true,
}

// Check if a string contains at least one letter or digit
func containsLetterOrDigit(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func ValidateKeyword(keyword string) error {
	keyword = strings.ToLower(strings.TrimSpace(keyword))

	if len(keyword) < 3 {
		return errors.New("keyword must be at least 3 characters long")
	}
	if stopWords[keyword] {
		return errors.New("keyword is too generic (stop word)")
	}
	if !containsLetterOrDigit(keyword) {
		return errors.New("keyword must contain at least one letter or digit")
	}
	if matched, _ := regexp.MatchString(`^[^a-zA-Z0-9]+$`, keyword); matched {
		return errors.New("keyword cannot be only special characters")
	}
	return nil
}
