package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_HumanizeString(t *testing.T) {
	tests := []struct {
		in     string
		expect string
	}{
		{"booking_number", "Booking Number"},
		{"port_of_loading", "Port Of Loading"},
		{"pickup_location", "Pickup Location"},
		{"return_location", "Return Location"},
		{"vessel", "Vessel"},
		{"earliest_receiving_date", "Earliest Receiving Date"},
		{"estimated_time_of_departure", "Estimated Time Of Departure"},
		{"estimated_time_of_arrival", "Estimated Time Of Arrival"},
		{"port_ramp_cut_off", "Port Ramp Cut Off"},
		{"shipping_instruction_cut_off", "Shipping Instruction Cut Off"},
		{"transhipment_port", "Transhipment Port"},
		{"quantity", "Quantity"},
		{"container_size", "Container Size"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.expect, HumanizeString(tt.in))
		})
	}
}

func Test_removeExtraSpaces(t *testing.T) {
	tests := []struct {
		in     string
		expect string
	}{
		{"  hello  world  ", "hello world"},
		{"hello  world", "hello world"},
		{"hello world", "hello world"},
		{" hello world ", "hello world"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.expect, RemoveExtraSpaces(tt.in))
		})
	}
}

func TestService_isEmpty(t *testing.T) {
	tests := []struct {
		name   string
		in     *string
		expect bool
	}{
		{"nil", nil, true},
		{"empty", Ptr(""), true},
		{"non-empty", Ptr("foo"), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expect, IsEmpty(test.in))
		})
	}

}

func TestFormatComment(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			// Test with multiple paragraphs
			input: `Really appreciate the detailed breakdown of your options! \n\nI've seen local service businesses get good results blending a lean city-specific site with a strong content hub that shows off niche expertise. Sometimes that knowledge attracts referrals beyond your main market, which isn't a bad thing. \n\nCurious if you've started drafting any FAQ posts yet, or are you still mapping strategy?`,
			expected: `Really appreciate the detailed breakdown of your options! 
I've seen local service businesses get good results blending a lean city-specific site with a strong content hub that shows off niche expertise. Sometimes that knowledge attracts referrals beyond your main market, which isn't a bad thing. 
Curious if you've started drafting any FAQ posts yet, or are you still mapping strategy?`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Format the input comment
			actual := FormatComment(tt.input)

			// Check if the actual output matches the expected result
			if actual != tt.expected {
				t.Errorf("FormatComment() = %v; expected %v", actual, tt.expected)
			}
		})
	}
}

func TestGetOrganizationName(t *testing.T) {
	tests := []struct {
		email        string
		expectPrefix string // since generateUnique() adds random suffix
		description  string
	}{
		{
			email:        "john@openai.com",
			expectPrefix: "Openai",
			description:  "custom domain should return capitalized domain prefix",
		},
		{
			email:        "jane.doe@gmail.com",
			expectPrefix: "jane.doe-",
			description:  "generic domain should return local part with suffix",
		},
		{
			email:        "user@unknowncustomdomain.org",
			expectPrefix: "Unknowncustomdomain",
			description:  "custom domain with uncommon TLD",
		},
		{
			email:        "invalid-email",
			expectPrefix: "user-",
			description:  "malformed email should fallback to user prefix",
		},
		{
			email:        "",
			expectPrefix: "user-",
			description:  "empty email should fallback to user prefix",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			orgName := GetOrganizationName(test.email)

			if !strings.HasPrefix(orgName, test.expectPrefix) {
				t.Errorf("expected prefix %q, got %q", test.expectPrefix, orgName)
			}
		})
	}
}

func TestIsValidProductName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid short name",
			input:    "Notion",
			expected: true,
		},
		{
			name:     "Valid multi-word name",
			input:    "Pixel Global",
			expected: true,
		},
		{
			name:     "Too many words",
			input:    "Affordable Professional SEO Services",
			expected: false,
		},
		{
			name:     "Too many characters",
			input:    "SuperLongMarketingProductWithWayTooManyCharacters",
			expected: false,
		},
		{
			name:     "Exactly at word and character limit",
			input:    "Best Pixel App", // 3 words, under 30 chars
			expected: true,
		},
		{
			name:     "Too many words but under char limit",
			input:    "Smart Productivity Organizer Tool",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Below minimum character length",
			input:    "AI",
			expected: false,
		},
		{
			name:     "Exactly minimum character length",
			input:    "Bot",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsValidProductName(tt.input)
			if actual != tt.expected {
				t.Errorf("isValidProductName(%q) = %v; expected %v", tt.input, actual, tt.expected)
			}
		})
	}
}
