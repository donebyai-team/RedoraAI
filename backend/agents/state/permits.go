package state

import (
	"context"
	_ "embed"
	"fmt"
	"go.uber.org/zap"

	"github.com/redis/go-redis/v9"
)

type TrackerState interface {
	AcquireTracker(ctx context.Context, organizationID, trackerID string, maxConcurrent int) (bool, error)
	ReleaseTracker(ctx context.Context, organizationID, trackerID string) error
	IsRunningTracker(ctx context.Context, organizationID, trackerID string) (bool, error)
	KeepAliveTracker(ctx context.Context, organizationID, trackerID string) error
}

//go:embed scripts/acquire_project_semaphore.lua
var acquireLuaScript string

func (p *customerCaseState) generateKey(key string) string {
	return callRunningKey(p.namespace, p.prefix, key)
}

func (p *customerCaseState) AcquireTracker(ctx context.Context, organizationID, trackerID string, maxConcurrent int) (bool, error) {
	setKey := p.generateKey(fmt.Sprintf("org:%s:active_trackers", organizationID))
	heartbeatKey := p.generateKey(fmt.Sprintf("org:%s:tracker:%s", organizationID, trackerID))

	// Cleanup expired members in Go before Lua call
	members, err := p.redisClient.SMembers(ctx, setKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to read active set: %w", err)
	}

	p.logger.Info("members found while acquiring tracking",
		zap.String("set_key", setKey),
		zap.Int("count", len(members)))

	memberToClean := 0
	for _, id := range members {
		hbKey := p.generateKey(fmt.Sprintf("org:%s:tracker:%s", organizationID, id))
		exists, err := p.redisClient.Exists(ctx, hbKey).Result()
		if err != nil {
			return false, fmt.Errorf("failed to check tracker key: %w", err)
		}
		if exists == 0 {
			if p.redisClient.SRem(ctx, setKey, id).Err() == nil {
				memberToClean++
			}
		}
	}

	p.logger.Info("members cleaned while acquiring tracking",
		zap.String("set_key", setKey),
		zap.Int("members_removed", memberToClean))

	script := redis.NewScript(acquireLuaScript)

	ttlSeconds := int(p.customerCaseRunningTTL.Seconds())
	res, err := script.Run(ctx, p.redisClient, []string{setKey}, maxConcurrent, trackerID, ttlSeconds, heartbeatKey).Result()
	if err != nil {
		return false, fmt.Errorf("acquire script failed: %w", err)
	}

	switch result := res.(int64); result {
	case 1:
		return true, nil // Acquired successfully
	case 2:
		return false, nil // Already running — duplicate
	case 0:
		return false, nil // Concurrency limit reached
	default:
		return false, fmt.Errorf("unexpected acquire return: %v", result)
	}
}

func (p *customerCaseState) ReleaseTracker(ctx context.Context, organizationID, trackerID string) error {
	setKey := p.generateKey(fmt.Sprintf("org:%s:active_trackers", organizationID))
	heartbeatKey := p.generateKey(fmt.Sprintf("org:%s:tracker:%s", organizationID, trackerID))
	_ = p.redisClient.SRem(ctx, setKey, trackerID)
	_ = p.redisClient.Del(ctx, heartbeatKey)
	return nil
}

func (p *customerCaseState) IsRunningTracker(ctx context.Context, organizationID, trackerID string) (bool, error) {
	heartbeatKey := p.generateKey(fmt.Sprintf("org:%s:tracker:%s", organizationID, trackerID))
	exists, err := p.redisClient.Exists(ctx, heartbeatKey).Result()
	if err != nil {
		return false, fmt.Errorf("redis check failed: %w", err)
	}
	return exists == 1, nil
}

func (p *customerCaseState) KeepAliveTracker(ctx context.Context, organizationID, trackerID string) error {
	heartbeatKey := p.generateKey(fmt.Sprintf("org:%s:tracker:%s", organizationID, trackerID))

	exists, err := p.redisClient.Exists(ctx, heartbeatKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check existence: %w", err)
	}

	if exists == 0 {
		err := p.redisClient.Set(ctx, heartbeatKey, 1, p.customerCaseRunningTTL).Err()
		if err != nil {
			return fmt.Errorf("failed to recreate missing heartbeat key: %w", err)
		}
		return nil
	}

	return p.redisClient.Expire(ctx, heartbeatKey, p.customerCaseRunningTTL).Err()
}
