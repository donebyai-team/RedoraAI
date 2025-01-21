package utils

import (
	"fmt"
	"os"
	"testing"
)

func GetEnvTestReq(t *testing.T, key string) string {
	value := os.Getenv(key)
	if value == "" {
		t.Skipf(fmt.Sprintf("missing required env %s", key))
	}
	return value
}
