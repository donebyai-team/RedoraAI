package state

import (
	"context"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"time"
)

type ConversationState interface {
	IsRunning(ctx context.Context, key string) (bool, error)
	KeepAlive(ctx context.Context, organizationID, key string) error
	Release(ctx context.Context, key string) error
	Acquire(ctx context.Context, organizationID, uniqueID string) error

	// ActiveCount returns the number of active cases across organizations
	ActiveCount(ctx context.Context) (uint64, error)
	Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
	CheckIfUnderLimitAndIncrement(ctx context.Context, redisKey string, field string, limit int64, expiry time.Duration) (bool, error)
	RollbackCounter(ctx context.Context, redisKey, field string) error
	GetLeadAnalysisCounters(ctx context.Context, redisKey string) (*pbportal.LeadAnalysis, error)

	TrackerState
}
