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

	MinDescriptionChars = 10
	MaxDescriptionWords = 250

	MinTargetPersonaChars = 10
	MaxTargetPersonaWords = 100
)

func IsValidDescription(desc string) bool {
	length := len(desc)
	if length < MinDescriptionChars {
		return false
	}
	if len(strings.Fields(desc)) > MaxDescriptionWords {
		return false
	}
	return true
}

func IsValidTargetPersona(persona string) bool {
	length := len(persona)
	if length < MinTargetPersonaChars {
		return false
	}
	if len(strings.Fields(persona)) > MaxTargetPersonaWords {
		return false
	}
	return true
}

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
		"gmail.com":         {},
		"googlemail.com":    {},
		"outlook.com":       {},
		"hotmail.com":       {},
		"live.com":          {},
		"msn.com":           {},
		"yahoo.com":         {},
		"yahoo.co.uk":       {},
		"ymail.com":         {},
		"icloud.com":        {},
		"me.com":            {},
		"mac.com":           {},
		"aol.com":           {},
		"protonmail.com":    {},
		"tutanota.com":      {},
		"gmx.com":           {},
		"gmx.net":           {},
		"mail.com":          {},
		"zoho.com":          {},
		"yopmail.com":       {},
		"fastmail.com":      {},
		"hushmail.com":      {},
		"mailinator.com":    {},
		"trashmail.com":     {},
		"temp-mail.org":     {},
		"10minutemail.com":  {},
		"guerrillamail.com": {},
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
