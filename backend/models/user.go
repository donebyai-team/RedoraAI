package models

import "time"

//go:generate go-enum -f=$GOFILE

// ENUM(USER, ADMIN, PLATFORM_ADMIN)
type UserRole string

// ENUM(PENDING, ACTIVE)
type UserState string

type User struct {
	ID             string     `db:"id"`
	Auth0ID        string     `db:"auth0_id"`
	Email          string     `db:"email"`
	EmailVerified  bool       `db:"email_verified"`
	OrganizationID string     `db:"organization_id"`
	Role           UserRole   `db:"role"`
	State          UserState  `db:"state"`
	CreatedAt      time.Time  `db:"created_at"`
	UpdatedAt      *time.Time `db:"updated_at"`
}

func (u *User) IsPlatformAdmin() bool {
	return u.Role == UserRolePLATFORMADMIN
}

func (u *User) IsAdmin() bool {
	return u.Role == UserRoleADMIN || u.IsPlatformAdmin()
}
