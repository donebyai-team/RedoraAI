package psql

import (
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"testing"
	"time"
)

func TestGetDateRange(t *testing.T) {
	loc := time.UTC
	fixedNow := time.Date(2025, 5, 19, 15, 0, 0, 0, loc)

	tests := []struct {
		name         string
		filter       pbportal.DateRangeFilter
		wantStartStr string
		wantEndStr   string
	}{
		{
			name:         "today",
			filter:       pbportal.DateRangeFilter_DATE_RANGE_TODAY,
			wantStartStr: "2025-05-19T00:00:00Z",
			wantEndStr:   "2025-05-20T00:00:00Z",
		},
		{
			name:         "yesterday",
			filter:       pbportal.DateRangeFilter_DATE_RANGE_YESTERDAY,
			wantStartStr: "2025-05-18T00:00:00Z",
			wantEndStr:   "2025-05-19T00:00:00Z",
		},
		{
			name:         "7_days",
			filter:       pbportal.DateRangeFilter_DATE_RANGE_7_DAYS,
			wantStartStr: "2025-05-13T00:00:00Z",
			wantEndStr:   "2025-05-20T00:00:00Z",
		},
		{
			name:         "invalid filter",
			filter:       pbportal.DateRangeFilter_DATE_RANGE_UNSPECIFIED,
			wantStartStr: "",
			wantEndStr:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := GetDateRange(tt.filter, fixedNow)

			if tt.wantStartStr != "" {
				wantStart, _ := time.Parse(time.RFC3339, tt.wantStartStr)
				if gotStart == nil || !gotStart.Equal(wantStart) {
					t.Errorf("expected start: %v, got: %v", wantStart, gotStart)
				}
			} else if gotStart != nil {
				t.Errorf("expected nil start, got: %v", gotStart)
			}

			if tt.wantEndStr != "" {
				wantEnd, _ := time.Parse(time.RFC3339, tt.wantEndStr)
				if gotEnd == nil || !gotEnd.Equal(wantEnd) {
					t.Errorf("expected end: %v, got: %v", wantEnd, gotEnd)
				}
			} else if gotEnd != nil {
				t.Errorf("expected nil end, got: %v", gotEnd)
			}
		})
	}
}
