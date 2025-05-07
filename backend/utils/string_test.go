package utils

import (
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
			input: `I haven’t tried those browsers myself, but I’m super curious to hear what others think.\n\nIf you come across someone who’s used **BrowserBase**, **Airtop.ai**, **Browser Use**, **Hyperbrowser**, or **Anchor Browser**, tag me — I’d love to join the convo!\n\nI’m planning to test them out myself soon too.`,
			expected: `I haven’t tried those browsers myself, but I’m super curious to hear what others think.

If you come across someone who’s used **BrowserBase**, **Airtop.ai**, **Browser Use**, **Hyperbrowser**, or **Anchor Browser**, tag me — I’d love to join the convo!

I’m planning to test them out myself soon too.`,
		},
		{
			// Test with only one paragraph (no paragraph breaks)
			input: `I haven’t tried those browsers myself, but I’m super curious to hear what others think.\nI’m planning to test them out myself soon too.`,
			expected: `I haven’t tried those browsers myself, but I’m super curious to hear what others think.
I’m planning to test them out myself soon too.`,
		},
		{
			// Test with no placeholders, input should remain unchanged
			input:    `I haven’t tried those browsers myself, but I’m super curious to hear what others think.`,
			expected: `I haven’t tried those browsers myself, but I’m super curious to hear what others think.`,
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
