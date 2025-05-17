package interactions

import (
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/shank318/doota/models"
	"golang.org/x/net/context"
	"sync"
	"time"
)

type RateLimitConfig struct {
	MaxDMs           int
	MaxComments      int
	CooldownDuration time.Duration
}

type RateLimiter interface {
	CanSend(ctx context.Context, projectID string, typ models.LeadInteractionType) bool
}

type InMemoryRateLimiter struct {
	mu           sync.Mutex
	lastReset    map[string]time.Time
	dmCount      map[string]int
	commentCount map[string]int
	config       RateLimitConfig
}

func NewInMemoryRateLimiter(config RateLimitConfig) *InMemoryRateLimiter {
	return &InMemoryRateLimiter{
		lastReset:    make(map[string]time.Time),
		dmCount:      make(map[string]int),
		commentCount: make(map[string]int),
		config:       config,
	}
}

func (r *InMemoryRateLimiter) CanSend(ctx context.Context, projectID string, typ models.LeadInteractionType) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	last := r.lastReset[projectID]
	if now.Sub(last) > r.config.CooldownDuration {
		r.lastReset[projectID] = now
		r.dmCount[projectID] = 0
		r.commentCount[projectID] = 0
	}

	switch typ {
	case models.LeadInteractionTypeDM:
		if r.dmCount[projectID] < r.config.MaxDMs {
			r.dmCount[projectID]++
			return true
		}
	case models.LeadInteractionTypeCOMMENT:
		if r.commentCount[projectID] < r.config.MaxComments {
			r.commentCount[projectID]++
			return true
		}
	}
	return false
}

type RedisRateLimiter struct {
	client *redis.Client
	config RateLimitConfig
	prefix string
}

func NewRedisRateLimiter(client *redis.Client, config RateLimitConfig, prefix string) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
		config: config,
		prefix: prefix,
	}
}

func (r *RedisRateLimiter) CanSend(ctx context.Context, projectID string, typ models.LeadInteractionType) bool {
	key := fmt.Sprintf("%s:%s:%s", r.prefix, projectID, typ)
	limit := 0

	switch typ {
	case models.LeadInteractionTypeDM:
		limit = r.config.MaxDMs
	case models.LeadInteractionTypeCOMMENT:
		limit = r.config.MaxComments
	}

	// Use INCR and EXPIRE to handle rate limiting atomically
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return false
	}

	if val == 1 {
		// First time set expiration
		r.client.Expire(ctx, key, r.config.CooldownDuration)
	}

	if int(val) > limit {
		return false
	}

	return true
}
