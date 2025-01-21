package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_isAdminPath(t *testing.T) {

	assert.Equal(t, false, isAdminPath("/doota.portal.v1.PortalService/GetConfig"))
	assert.Equal(t, true, isAdminPath("/doota.portal.v1.AdminService/GetUser"))
}
