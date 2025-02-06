package state

import "context"

type ConversationState interface {
	// IsRunning returns true if a case is currently running
	IsRunning(ctx context.Context, phone string) (bool, error)

	// ActiveCount returns the number of active cases across organizations
	ActiveCount(ctx context.Context) (uint64, error)

	// KeepAlive signals that the case is still being processed, this should be called periodically
	KeepAlive(ctx context.Context, organizationID, phone string) error

	Release(ctx context.Context, phone string) error
}
