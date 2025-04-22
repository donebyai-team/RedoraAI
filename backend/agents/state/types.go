package state

import (
	"strings"
	"time"
)

//const namespace = "spool"
//const phonePrefix = "phone"
//const organizationPrefix = "org"

func callRunningKey(namespace, prefix, id string) string {
	return strings.Join([]string{namespace, prefix, id}, ":")
}

//	func orgRunningKey(orgID, phone string) string {
//		return strings.Join([]string{namespace, organizationPrefix, phonePrefix, orgID, phone}, ":")
//	}
//
//	func orgKeyPattern(organizationID string) string {
//		return strings.Join([]string{namespace, organizationPrefix, phonePrefix, organizationID}, ":")
//	}
func allCallKeyPattern(namespace, prefix string) string {
	return strings.Join([]string{namespace, prefix}, ":")
}

type CustomerCaseState struct {
	CaseID         string    `json:"cid"`
	Phone          string    `json:"ph"`
	OrganizationID string    `json:"oid"`
	StartedAt      time.Time `json:"sat"`
}
