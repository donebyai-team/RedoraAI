package state

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type customerCaseState struct {
	redisClient            *redis.Client
	customerCaseRunningTTL time.Duration
	logger                 *zap.Logger
	namespace              string
	prefix                 string
}

func NewCustomerCaseState(redisAddr string, customerCaseTTL time.Duration, logger *zap.Logger, namespace, prefix string) *customerCaseState {
	var redisClient *redis.Client
	// Check if redisAddr starts with the redis:// scheme
	if len(redisAddr) > 6 && redisAddr[:6] == "redis:" {
		// Parse the Redis URL
		parsedURL, err := url.Parse(redisAddr)
		if err != nil {
			log.Fatalf("Error parsing Redis URL: %v", err)
		}

		// Extracting user and password from the URL
		password, _ := parsedURL.User.Password()

		// Extracting the host and port
		host := parsedURL.Hostname()
		port := parsedURL.Port()

		// Set up Redis client options
		options := &redis.Options{
			Addr:     fmt.Sprintf("%s:%s", host, port),
			Password: password, // Password from the URL
		}
		redisClient = redis.NewClient(options)
	} else {
		// Use the simple address like localhost:6379
		redisClient = redis.NewClient(&redis.Options{
			Addr: redisAddr,
		})
	}

	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		logger.Error("Error connecting to Redis", zap.Error(err))
	}

	return &customerCaseState{
		redisClient:            redisClient,
		customerCaseRunningTTL: customerCaseTTL,
		logger:                 logger,
		namespace:              namespace,
		prefix:                 prefix,
	}
}

func (r *customerCaseState) Acquire(ctx context.Context, organizationID, uniqueID string) error {
	key := callRunningKey(r.namespace, r.prefix, uniqueID)
	value := &CustomerCaseState{
		Phone:          uniqueID,
		OrganizationID: organizationID,
		StartedAt:      time.Now(),
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshalling : %w", err)
	}

	ok, err := r.redisClient.SetNX(ctx, key, data, r.customerCaseRunningTTL).Result()
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}

	if !ok {
		return fmt.Errorf("lock already held")
	}

	return nil
}

func (r *customerCaseState) Release(ctx context.Context, phone string) error {
	key := callRunningKey(r.namespace, r.prefix, phone)
	if cmd := r.redisClient.Del(ctx, key); cmd.Err() != nil {
		return fmt.Errorf("release case state: %w", cmd.Err())
	}
	return nil
}

func (r *customerCaseState) IsRunning(ctx context.Context, phone string) (bool, error) {
	key := callRunningKey(r.namespace, r.prefix, phone)
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
	iter := r.redisClient.Scan(ctx, 0, allCallKeyPattern(r.namespace, r.prefix)+"*", 0).Iterator()
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
	key := callRunningKey(r.namespace, r.prefix, phone)
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

	key := callRunningKey(r.namespace, r.prefix, phone)
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
