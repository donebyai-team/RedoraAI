package state

import (
	"context"
	"encoding/json"
	"fmt"
	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
	"log"
	"net/url"
	"strconv"
	"strings"
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

	ttl := r.customerCaseRunningTTL
	if strings.Contains(uniqueID, "interactions") {
		ttl = 2 * time.Minute
	}

	ok, err := r.redisClient.SetNX(ctx, key, data, ttl).Result()
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

func (r *customerCaseState) Set(ctx context.Context, key string, data interface{}, ttl time.Duration) error {
	key = callRunningKey(r.namespace, r.prefix, key)
	if err := r.setKey(ctx, key, data, ttl); err != nil {
		return fmt.Errorf("set case state: %w", err)
	}

	return nil
}

func (r *customerCaseState) Get(ctx context.Context, key string) ([]byte, error) {
	key = callRunningKey(r.namespace, r.prefix, key)
	value, err := r.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return value, fmt.Errorf("get state: %w", err)
	}

	return value, nil
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

var checkAndIncrementLua = redis.NewScript(`
    local current = redis.call("HGET", KEYS[1], ARGV[1])
    if not current then current = 0 else current = tonumber(current) end

    if current < tonumber(ARGV[2]) then
        redis.call("HINCRBY", KEYS[1], ARGV[1], 1)
        return 1
    else
        return 0
    end
`)

func (r *customerCaseState) CheckIfUnderLimitAndIncrement(ctx context.Context, redisKey string, field string, limit int64, expiry time.Duration) (bool, error) {
	redisKey = callRunningKey(r.namespace, r.prefix, redisKey)
	// Set TTL only if not already set
	ttl, err := r.redisClient.TTL(ctx, redisKey).Result()
	if err == nil && ttl < 0 {
		r.redisClient.Expire(ctx, redisKey, expiry)
	}

	// Run Lua script atomically
	res, err := checkAndIncrementLua.Run(ctx, r.redisClient, []string{redisKey}, field, limit).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func (r *customerCaseState) RollbackCounter(ctx context.Context, redisKey, field string) error {
	redisKey = callRunningKey(r.namespace, r.prefix, redisKey)
	return r.redisClient.HIncrBy(ctx, redisKey, field, -1).Err()
}

func (r *customerCaseState) GetLeadAnalysisCounters(ctx context.Context, redisKey string) (*pbportal.LeadAnalysis, error) {
	redisKey = callRunningKey(r.namespace, r.prefix, redisKey)
	data, err := r.redisClient.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return nil, err
	}

	convert := func(val string) uint32 {
		if i, err := strconv.ParseUint(val, 10, 32); err == nil {
			return uint32(i)
		}
		return 0
	}

	return &pbportal.LeadAnalysis{
		PostsTracked:       convert(data["posts_tracked"]),
		RelevantPostsFound: convert(data["relevant_posts"]),
		CommentSent:        convert(data["comment_sent"]),
		CommentScheduled:   convert(data["comment_scheduled"]),
		DmSent:             convert(data["dm_sent"]),
		DmScheduled:        convert(data["dm_scheduled"]),
	}, nil
}
