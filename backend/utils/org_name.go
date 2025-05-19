package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"
)

const (
	MaxProductNameWords = 3
	MaxProductNameChars = 30
	MinProductNameChars = 3
)

func IsValidProductName(name string) bool {
	length := len(name)
	if length < MinProductNameChars || length > MaxProductNameChars {
		return false
	}
	if len(strings.Fields(name)) > MaxProductNameWords {
		return false
	}
	return true
}

func GetOrganizationName(email string) string {
	genericDomains := map[string]struct{}{
		"gmail.com":      {},
		"outlook.com":    {},
		"hotmail.com":    {},
		"yahoo.com":      {},
		"icloud.com":     {},
		"aol.com":        {},
		"protonmail.com": {},
		"gmx.com":        {},
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return generateUnique("user") // fallback if malformed email
	}

	localPart := parts[0]
	domain := strings.ToLower(parts[1])

	if _, isGeneric := genericDomains[domain]; isGeneric {
		return generateUnique(localPart)
	}

	// for custom domains, use domain prefix (e.g., "openai" from "openai.com")
	org := strings.Split(domain, ".")[0]
	return CapitalizeFirst(org)
}

func CapitalizeFirst(s string) string {
	if s == "" {
		return s
	}

	// Convert first rune to upper, append rest
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// generateUnique appends a short random string to the base for uniqueness
func generateUnique(base string) string {
	return fmt.Sprintf("%s-%s", base, shortUUID(6))
}

// shortUUID generates a short pseudo-UUID string (not cryptographically secure)
func shortUUID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
