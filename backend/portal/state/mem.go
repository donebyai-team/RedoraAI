package state

import (
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ AuthStateStore = (*memStore)(nil)

type memStore struct {
	store  map[string][]byte
	logger *zap.Logger
}

func NewMemStore(logger *zap.Logger) *memStore {
	return &memStore{
		store:  map[string][]byte{},
		logger: logger.Named("state_store"),
	}
}

func (m *memStore) SetState(s *State) error {
	cnt, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	return m.save(s.Hash, cnt)
}

func (m *memStore) GetState(hash string) (*State, error) {
	cnt, err := m.get(hash)
	if err != nil {
		return nil, err
	}
	s := &State{}
	if err := json.Unmarshal(cnt, s); err != nil {
		return nil, fmt.Errorf("unmarhsal state: %w", err)
	}
	return s, nil
}

func (m *memStore) DelState(hash string) error {
	m.del(hash)
	return nil
}

func (m *memStore) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddInt("count", len(m.store))
	return nil
}

func (m *memStore) save(key string, cnt []byte) error {
	m.store[key] = cnt
	m.logger.Debug("state saved", zap.String("key", key), zap.Int("count", len(m.store)))
	return nil
}

func (m *memStore) get(key string) ([]byte, error) {
	cnt, found := m.store[key]
	if !found {
		return nil, NotFound
	}
	return cnt, nil

}

func (m *memStore) del(key string) {
	delete(m.store, key)
	m.logger.Debug("state deleted", zap.String("key", key), zap.Int("count", len(m.store)))
}
