package utils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcess_HTMLToText(t *testing.T) {
	tests := []struct {
		name     string
		htmlFile string
		expect   string
	}{
		{
			name:     "basic html",
			htmlFile: "basic.html",
			expect:   "hello world",
		},
		{
			name:     "plain text",
			htmlFile: "no_tag.html",
			expect:   "hello world",
		},
		{
			name:     "complex html",
			htmlFile: "complex.html",
			expect:   "html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/" + tt.htmlFile)
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, HTMLToText(string(data)))
		})
	}
}
