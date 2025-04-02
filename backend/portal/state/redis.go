package state

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

var _ AuthStateStore = (*redisStore)(nil)

type redisStore struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

func NewRedisStore(redisAddr string, logger *zap.Logger) *redisStore {
	redisClient := redis.NewClient(&redis.Options{Addr: redisAddr})
	return &redisStore{
		redisClient: redisClient,
		logger:      logger,
	}
}

// DelState implements AuthStateStore.
func (r *redisStore) DelState(hash string) error {
	key := stateKey(hash)
	if cmd := r.redisClient.Del(key); cmd.Err() != nil {
		return fmt.Errorf("del auth state: %w", cmd.Err())
	}
	return nil
}

// GetState implements AuthStateStore.
func (r *redisStore) GetState(hash string) (*State, error) {
	key := stateKey(hash)
	value, err := r.redisClient.Get(key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, NotFound
		}
		return nil, fmt.Errorf("get auth state: %w", err)
	}

	s := &State{}
	if err := json.Unmarshal(value, s); err != nil {
		return nil, fmt.Errorf("unmarhsal state: %w", err)
	}
	return s, nil
}

// SetState implements AuthStateStore.
func (r *redisStore) SetState(s *State) error {
	key := stateKey(s.Hash)
	value, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if cmd := r.redisClient.Set(key, value, 5*time.Minute); cmd.Err() != nil {
		return fmt.Errorf("set auth state: %w", cmd.Err())
	}
	return nil
}

const namespace = "auth"
const statePrefix = "state"

func stateKey(hash string) string {
	return strings.Join([]string{namespace, statePrefix, hash}, ":")
}
