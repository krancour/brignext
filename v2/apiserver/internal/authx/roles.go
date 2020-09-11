package authx

import "context"

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

	// Special roles
	//
	// These are reserved for use by system components and are assignable to Users
	// and ServiceAccounts.

	// RoleNameObserver is the name of a Role that enables principals to updates
	// Worker and Job status based on observation of the underlying workload
	// execution substrate. This Role is exclusively for the use of the Observer
	// component.
	RoleNameObserver RoleName = "OBSERVER"
	// RoleNameScheduler is the name of a Role that enables principals to initiate
	// execution of a Worker or Job on the underlying workload execution
	// substrate. This Role is execlusively for the use of the Scheduler
	// component.
	RoleNameScheduler RoleName = "SCHEDULER"
	// RoleNameWorker is the name of a Role that enables principals to create new
	// Jobs. This Role is exclusively for the use of Workers.
	RoleNameWorker RoleName = "WORKER"
)

// RoleScopeGlobal represents an unbounded scope.
const RoleScopeGlobal = "*"

const RoleTypeSystem = "SYSTEM"
const RoleTypeProject = "PROJECT"

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount.
type Role struct {
	Type string `json:"type" bson:"type"`
	// Name is the name of a Role and has domain-specific meaning.
	Name RoleName `json:"name" bson:"name"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope" bson:"scope"`
}

type RoleAssignment struct {
	Role          RoleName      `json:"role"`
	PrincipalType PrincipalType `json:"principalType"`
	PrincipalID   string        `json:"principalID"`
}

// RoleAdmin returns a Role that enables a principal to manage Users,
// ServiceAccounts, and globally scoped permissions for Users and
// ServiceAccounts.
func RoleAdmin() Role {
	return Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameAdmin,
		Scope: RoleScopeGlobal,
	}
}

// RoleEventCreator returns a Role that enables a principal to create new Events
// having a Source field whose value matches that of the Scope field. This Role
// is useful for ServiceAccounts used for gateways.
func RoleEventCreator(eventSource string) Role {
	return Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameEventCreator,
		Scope: eventSource,
	}
}

// RoleProjectAdmin returns a Role that enables a principal to manage a Project
// having an ID field whose value matches that of the Scope field. If the value
// of the Scope field is RoleScopeGlobal ("*"), then the Role is unbounded and
// enables a principal to manage all Projects.
func RoleProjectAdmin(projectID string) Role {
	return Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectAdmin,
		Scope: projectID,
	}
}

// RoleProjectCreator returns a Role that enables a principal to create new
// Projects.
func RoleProjectCreator() Role {
	return Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameProjectCreator,
		Scope: RoleScopeGlobal,
	}
}

// RoleProjectDeveloper returns a Role that enables a principal to read and
// update a Project having an ID field whose value matches that of the Scope
// field. If the value of the Scope field is RoleScopeGlobal ("*"), then the
// Role is unbounded and enables a principal to read and update all Projects.
func RoleProjectDeveloper(projectID string) Role {
	return Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectDeveloper,
		Scope: projectID,
	}
}

// RoleProjectUser returns a Role that enables a principal read, create, and
// manage Events for a Project having an ID field whose value matches the Scope
// field. If the value of the Scope field is RoleScopeGlobal ("*"), then the
// Role is unbounded and enables a principal to read, create, and manage Events
// for all Projects.
func RoleProjectUser(projectID string) Role {
	return Role{
		Type:  RoleTypeProject,
		Name:  RoleNameProjectUser,
		Scope: projectID,
	}
}

// RoleReader returns a Role that enables a principal to list and read Projects,
// Events, Users, and Service Accounts.
func RoleReader() Role {
	return Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameReader,
		Scope: RoleScopeGlobal,
	}
}

// Special roles
//
// These are reserved for use by system components and are assignable to Users
// and ServiceAccounts.

// RoleObserver returns a Role that enables a principal to update Worker and Job
// statuses based on observations of the underlying workload execution
// substrate. This Role is exclusively for the use of the Observer component.
func RoleObserver() Role {
	return Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameObserver,
		Scope: RoleScopeGlobal,
	}
}

// RoleScheduler returns a Role that enables a principal to initiate execution
// of Workers and Jobs on the underlying workload execution substrate. This Role
// is exclusively for the use of the Scheduler component.
func RoleScheduler() Role {
	return Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameScheduler,
		Scope: RoleScopeGlobal,
	}
}

// RoleWorker returns a Role that enables a principal to create Jobs for the
// Event whose ID matches the Scope. This Role is exclusively for the use of
// Workers.
func RoleWorker(eventID string) Role {
	return Role{
		Type:  RoleTypeSystem,
		Name:  RoleNameWorker,
		Scope: eventID,
	}
}

type RolesStore interface {
	GrantRole(
		ctx context.Context,
		principalType PrincipalType,
		principalID string,
		roles ...Role,
	) error
	RevokeRole(
		ctx context.Context,
		principalType PrincipalType,
		principalID string,
		roles ...Role,
	) error
}