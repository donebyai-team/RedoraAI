package models

import "time"

type Project struct {
	ID                 string     `db:"id"`
	OrganizationID     string     `db:"organization_id"`
	Name               string     `db:"name"`
	ProductDescription string     `db:"description"`
	CustomerPersona    string     `db:"customer_persona"`
	EngagementGoals    string     `db:"goals"`
	WebsiteURL         string     `db:"website"`
	CreatedAt          time.Time  `db:"created_at"`
	UpdatedAt          *time.Time `db:"updated_at"`
}
