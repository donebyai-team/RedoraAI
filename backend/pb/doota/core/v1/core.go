package pbcore

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (x *TzTimestamp) FromTimePtr(t *time.Time) *TzTimestamp {
	if t == nil {
		return nil
	}
	return x.FromTime(*t)
}

func (x *TzTimestamp) FromTime(t time.Time) *TzTimestamp {
	x.Timestamp = timestamppb.New(t)
	_, offset := t.Zone()
	x.Offset = int32(offset / 3600)
	return x
}

func (x *TzTimestamp) ToTime() time.Time {
	timeInUTC := x.Timestamp.AsTime()
	// Define an offset in seconds (e.g., -5 hours for UTC-5)
	offset := int(x.Offset * 60 * 60)
	// Create a timezone with the offset
	timezone := time.FixedZone(fmt.Sprintf("UTC%d", x.Offset), offset)
	// Convert the time to the specified timezone
	return timeInUTC.In(timezone)
}

func (x *TzTimestamp) ToTimePtr() *time.Time {
	if x == nil {
		return nil
	}
	out := x.ToTime()
	return &out
}
