package utils

import (
	"google.golang.org/protobuf/types/known/timestamppb"
	"strings"
	"time"

	"go.uber.org/zap"
)

var ESTTimezone *time.Location

var TZOffset = map[string]string{
	"EST": "-0500",
	"CST": "-0600",
	"MST": "-0700",
}

//"shipping-instruction-cut-off": "28 Feb 2024 15:00(CST)",

func init() {
	var err error
	ESTTimezone, err = time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
}

func MapTZ(in string) string {
	for code, offset := range TZOffset {
		in = strings.Replace(in, code, offset, -1)
	}
	return in
}

func ToTimezone(in string, tz *time.Location, logger *zap.Logger) *time.Time {
	if in == "" {
		return nil
	}

	t, err := time.Parse("2006-01-02T15:04:05", in)
	if err != nil {
		logger.Warn("failed to parse time", zap.String("time", in), zap.Error(err))
		return nil
	}
	if tz != nil {
		t = t.In(tz)
	}
	return &t
}

// Helper: Convert *time.Time to *timestamppb.Timestamp
func TimestamppbOrNil(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}
