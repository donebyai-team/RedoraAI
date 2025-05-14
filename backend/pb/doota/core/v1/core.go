package pbcore

import (
	"fmt"
	"github.com/shank318/doota/models"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (i *SubscriptionStatus) FromModel(model models.SubscriptionStatus) {
	enum, found := SubscriptionStatus_value["SUBSCRIPTION_STATUS_"+strings.ToUpper(model.String())]
	if !found {
		panic(fmt.Errorf("unknown subscription status %q", model.String()))
	}
	*i = SubscriptionStatus(enum)
}

func (u *UsageLimit) FromModel(model *models.UsageLimits) *UsageLimit {
	u.PerDay = int32(model.PerDay)
	u.PerMonth = int32(model.PerMonth)
	return u
}

func (u *Subscription) FromModel(model *models.Subscription) *Subscription {
	u.Status.FromModel(model.GetStatus())
	u.CreatedAt = timestamppb.New(model.CreatedAt)
	u.ExpiresAt = timestamppb.New(model.ExpiresAt)
	u.MaxSources = int32(model.Metadata.MaxSources)
	u.MaxKeywords = int32(model.Metadata.MaxKeywords)
	u.Comments = new(UsageLimit).FromModel(&model.Metadata.Comments)
	u.Dm = new(UsageLimit).FromModel(&model.Metadata.DMs)
	return u
}

func (u *Source_RedditMetadata) FromModel(metadata *models.SubRedditMetadata) *Source_RedditMetadata {
	u.RedditMetadata = &SubRedditMetadata{
		Title:     metadata.Title,
		CreatedAt: timestamppb.New(metadata.CreatedAt),
	}
	return u
}

func (u *Source) FromModel(source *models.Source, details isSource_Details) *Source {
	u.Id = source.ID
	u.Name = source.Name
	u.Description = source.Description
	u.SourceType.FromModel(source.SourceType)
	u.Details = details
	return u
}

func (r *SourceType) FromModel(status models.SourceType) {
	enum, found := SourceType_value["SOURCE_TYPE_"+strings.ToUpper(string(status))]
	if !found {
		panic(fmt.Errorf("unknown source type %q", status))
	}
	*r = SourceType(enum)
}

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
