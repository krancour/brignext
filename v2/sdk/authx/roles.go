package authx

import (
	"encoding/json"

	"github.com/brigadecore/brigade/v2/sdk/meta"
)

type RoleName string

const (
	// RoleNameAdmin is the name of a Role that enables principals to manage
	// Users, ServiceAccounts, and globally scoped permissions for Users and
	// ServiceAccounts.
	RoleNameAdmin RoleName = "ADMIN"
	// RoleNameEventCreator is the name of a Role that enables principals to
	// create Events.
	RoleNameEventCreator RoleName = "EVENT_CREATOR"
	// RoleNameProjectAdmin is the name of a Role that enables principals to
	// manage all aspects of Projects.
	RoleNameProjectAdmin RoleName = "PROJECT_ADMIN"
	// RoleNameProjectCreator is the name of a Role that enables principals to
	// create new Projects.
	RoleNameProjectCreator RoleName = "PROJECT_CREATOR"
	// RoleNameProjectDeveloper is the name of a Role that enables principals to
	// read and update Projects.
	RoleNameProjectDeveloper RoleName = "PROJECT_DEVELOPER"
	// RoleNameProjectUser is the name of a Role that enables principals to
	// read, create, and manage Events for a Project.
	RoleNameProjectUser RoleName = "PROJECT_USER"
	// RoleNameReader is the name of a Role that enables principals to
	// list and read Projects, Users, and Service Accounts.
	RoleNameReader RoleName = "READER"
)

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount.
type Role struct {
	Type string `json:"type,omitempty"`
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name,omitempty"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope,omitempty"`
}

type RoleAssignment struct {
	Role          RoleName      `json:"role,omitempty"`
	Scope         string        `json:"scope,omitempty"`
	PrincipalType PrincipalType `json:"principalType,omitempty"`
	PrincipalID   string        `json:"principalID,omitempty"`
}

// MarshalJSON amends ServiceAccountList instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (r RoleAssignment) MarshalJSON() ([]byte, error) {
	type Alias RoleAssignment
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "RoleAssignment",
			},
			Alias: (Alias)(r),
		},
	)
}
