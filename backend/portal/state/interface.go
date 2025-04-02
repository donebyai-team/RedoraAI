package state

import (
	"errors"
	"strings"
	"time"

	pbportal "github.com/shank318/doota/pb/doota/portal/v1"
)

//go:generate go-enum -f=$GOFILE

// ENUM(integration, message_source)
type AuthStateContext string

type AuthStateStore interface {
	GetState(hash string) (*State, error)
	DelState(hash string) error
	SetState(s *State) error
}

type State struct {
	Hash            string                   `json:"hash"`
	Nonce           string                   `json:"nonce"`
	Context         AuthStateContext         `json:"context"`
	ContextId       *string                  `json:"context_id,omitempty"` // message_source_id
	IntegrationType pbportal.IntegrationType `json:"integrationType,omitempty"`
	RedirectUri     string                   `json:"redirect_uri"`
	ExpiresAt       time.Time                `json:"expires_at"`
}

func (s *State) HasExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

func orgNameFromEmail(email string) string {
	return strings.Split(email, "@")[0]
}

var NotFound = errors.New("state not found")
