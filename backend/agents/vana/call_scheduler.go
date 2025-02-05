package vana

import (
	"github.com/shank318/doota/models"
	"time"
)

const (
	maxCallPerDay         = 2
	maxTotalAllowedCalls  = 5
	minCallHour           = 10
	maxCallHour           = 17
	callRetryDelayMinutes = 10
	istLocation           = "Asia/Kolkata"
)

func getCustomerCallStats(conversations []*models.Conversation) (int, int) {
	loc, _ := time.LoadLocation(istLocation)
	now := time.Now().In(loc)
	today := now.Format("2006-01-02")
	totalCalls := len(conversations)
	callsToday := 0

	for _, conv := range conversations {
		if conv.CallStatus == models.CallStatusENDED {
			totalCalls++
			if conv.CreatedAt.In(loc).Format("2006-01-02") == today {
				callsToday++
			}
		}
	}

	return callsToday, totalCalls
}

func getNextCallTime(callsToday int, lastCalledAt time.Time) *time.Time {
	loc, _ := time.LoadLocation(istLocation)
	now := lastCalledAt.In(loc)
	nextCallTime := now.Add(time.Minute * callRetryDelayMinutes)
	if nextCallTime.Hour() < minCallHour {
		nextCallTime = time.Date(nextCallTime.Year(), nextCallTime.Month(), nextCallTime.Day(), minCallHour, 0, 0, 0, loc)
	}

	if nextCallTime.Hour() >= maxCallHour || callsToday >= maxCallPerDay {
		nextCallTime = time.Date(nextCallTime.Year(), nextCallTime.Month(), nextCallTime.Day()+1, minCallHour, 0, 0, 0, loc)
	}

	return &nextCallTime
}
