package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_realIPFromHeader(t *testing.T) {

	cases := []struct {
		name          string
		xforwardedFor []string
		remoteAddr    string
		expectedIP    string
	}{
		{
			name:          "sunny path",
			xforwardedFor: []string{"12.34.56.78, 23.45.67.89"},
			expectedIP:    "12.34.56.78",
		},
		{
			name:          "more then 2 ips",
			xforwardedFor: []string{"8.8.8.8,12.34.56.78, 23.45.67.89"},
			expectedIP:    "12.34.56.78",
		},
		{
			name:          "more then 2 ips as different headers",
			xforwardedFor: []string{"8.8.8.8", "12.34.56.78", "23.45.67.89"},
			expectedIP:    "12.34.56.78",
		},
		{
			name:          "more then 2 ips as different headers, mixed",
			xforwardedFor: []string{"8.8.8.8", "12.34.56.78, 23.45.67.89"},
			expectedIP:    "12.34.56.78",
		},
		{
			name:          "single ip",
			xforwardedFor: []string{"12.34.56.78"},
			expectedIP:    "12.34.56.78",
		},
		{
			name:          "no ip",
			xforwardedFor: []string{""},
			expectedIP:    "0.0.0.0",
		},
		{
			name:          "with junk",
			xforwardedFor: []string{"foo bar, 12.34.56.78, 23.45.67.89"},
			expectedIP:    "12.34.56.78",
		},
		{
			name:       "from remote addr",
			remoteAddr: "12.34.56.78",
			expectedIP: "12.34.56.78",
		},
		{
			name:       "from remote addr with port",
			remoteAddr: "12.34.56.78:54321",
			expectedIP: "12.34.56.78",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := &http.Request{
				Header: map[string][]string{"X-Forwarded-For": c.xforwardedFor},
			}
			req.RemoteAddr = c.remoteAddr
			ip := RealIP(c.remoteAddr, req.Header)
			assert.Equal(t, c.expectedIP, ip)
		})
	}
}
