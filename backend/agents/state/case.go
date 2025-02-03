package state

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type customerCaseState struct {
	redisClient               *redis.Client
	customerCaseRunningTTL    time.Duration
	customerCaseRetryCooldown time.Duration
	logger                    *zap.Logger
}

func NewCustomerCaseState(redisAddr string, customerCaseTTL time.Duration, customerCaseRetryCooldown time.Duration, logger *zap.Logger) *customerCaseState {
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &customerCaseState{
		redisClient:               redisClient,
		customerCaseRunningTTL:    customerCaseTTL,
		customerCaseRetryCooldown: customerCaseRetryCooldown,
		logger:                    logger,
	}
}

func (r *customerCaseState) IsRunning(ctx context.Context, phone string) (bool, error) {
	key := callRunningKey(phone)
	_, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("get investigation state: %w", err)
	}

	return true, nil
}

func (r *customerCaseState) ActiveCount(ctx context.Context) (uint64, error) {
	iter := r.redisClient.Scan(ctx, 0, allCallKeyPattern()+"*", 0).Iterator()
	count := uint64(0)
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("scan keys: %w", err)
	}
	return count, nil
}

func (r *customerCaseState) KeepAlive(ctx context.Context, organizationID, phone string) error {
	key := callRunningKey(phone)
	value, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("get state: %w", err)
	}

	if err == redis.Nil {
		return r.start(ctx, organizationID, phone)
	}

	if cmd := r.redisClient.Set(ctx, key, value, r.customerCaseRunningTTL); cmd.Err() != nil {
		return fmt.Errorf("set case state: %w", cmd.Err())
	}

	return nil
}

func (r *customerCaseState) start(ctx context.Context, organizationID, phone string) error {
	caseState := &CustomerCaseState{
		Phone:          phone,
		OrganizationID: organizationID,
		StartedAt:      time.Now(),
	}

	key := callRunningKey(phone)
	if err := r.setKey(ctx, key, caseState, r.customerCaseRunningTTL); err != nil {
		return fmt.Errorf("set call running state: %w", err)
	}

	return nil
}

func (i *customerCaseState) setKey(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshalling : %w", err)
	}

	// TODO: what timeout do I take
	if cmd := i.redisClient.Set(ctx, key, data, ttl); cmd.Err() != nil {
		return fmt.Errorf("set key: %w", cmd.Err())
	}
	return nil
}
