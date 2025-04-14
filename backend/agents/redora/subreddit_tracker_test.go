package redora

import (
	"github.com/shank318/doota/ai"
	"testing"
)

func TestSubRedditTracker(t *testing.T) {
	tracker := SubRedditTracker{
		gptModel:          ai.GPTModelGpt4O20240806,
		db:                nil,
		aiClient:          nil,
		logger:            nil,
		state:             nil,
		redditOauthClient: nil,
	}
}
