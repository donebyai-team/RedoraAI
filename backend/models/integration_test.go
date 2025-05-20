package models

import (
	"testing"
	"time"
)

func TestIsUserOldEnough(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		createdAgo time.Duration
		thresholdW int
		want       bool
	}{
		{
			name:       "Exactly 2 weeks old",
			createdAgo: 14 * 24 * time.Hour,
			thresholdW: 2,
			want:       true,
		},
		{
			name:       "Older than threshold",
			createdAgo: 30 * 24 * time.Hour,
			thresholdW: 2,
			want:       true,
		},
		{
			name:       "Younger than threshold",
			createdAgo: 7 * 24 * time.Hour,
			thresholdW: 2,
			want:       false,
		},
		{
			name:       "Exactly on the boundary",
			createdAgo: time.Duration(2*7*24) * time.Hour,
			thresholdW: 2,
			want:       true,
		},
		{
			name:       "Zero age (just created)",
			createdAgo: 0,
			thresholdW: 2,
			want:       false,
		},
		{
			name:       "2 days ago",
			createdAgo: time.Duration(2*24) * time.Hour,
			thresholdW: 2,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := RedditConfig{
				CreatedUtc: float64(now.Add(-tt.createdAgo).Unix()),
			}
			got := r.IsUserOldEnough(tt.thresholdW)
			if got != tt.want {
				t.Errorf("IsUserOldEnough(%d) = %v; want %v", tt.thresholdW, got, tt.want)
			}
		})
	}
}
