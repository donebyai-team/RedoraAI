package errorx

import "fmt"

type LoginError struct {
	Reason string
}

func (e *LoginError) Error() string {
	return fmt.Sprintf("login failed: %s", e.Reason)
}

type RefreshTokenError struct {
	Reason string
}

func (e *RefreshTokenError) Error() string {
	return fmt.Sprintf("failed to refresh token: %s", e.Reason)
}
