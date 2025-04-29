package state

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
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

	_, err := redisClient.Ping().Result()
	if err != nil {
		logger.Error("Error connecting to Redis", zap.Error(err))
	}
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
		return fmt.Errorf("set auth  state: %w", cmd.Err())
	}
	return nil
}

const namespace = "auth"
const statePrefix = "state"

func stateKey(hash string) string {
	return strings.Join([]string{namespace, statePrefix, hash}, ":")
}
