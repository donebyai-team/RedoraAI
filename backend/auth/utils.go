package auth

import (
	"strings"
)

func isAdminPath(in string) bool {
	chunks := strings.Split(in, "/")
	if len(chunks) != 3 {
		return false
	}
	return chunks[1] == "doota.portal.v1.AdminService"
}

func getHeader(headers map[string][]string, key string) string {
	if headers == nil {
		return ""
	}
	if _, ok := headers[key]; !ok {
		return ""
	}
	return headers[key][0]
}
