package authn

const (
	// RoleNameAdmin is the name of a Role that enables principals to manage
	// Users, ServiceAccounts, and globally scoped permissions for Users and
	// ServiceAccounts.
	RoleNameAdmin = "ADMIN"
	// RoleNameEventCreator is the name of a Role that enables principals to
	// create Events.
	RoleNameEventCreator = "EVENT_CREATOR"
	// RoleNameProjectAdmin is the name of a Role that enables principals to
	// manage all aspects of Projects.
	RoleNameProjectAdmin = "PROJECT_ADMIN"
	// RoleNameProjectCreator is the name of a Role that enables principals to
	// create new Projects.
	RoleNameProjectCreator = "PROJECT_CREATOR"
	// RoleNameProjectDeveloper is the name of a Role that enables principals to
	// read and update Projects.
	RoleNameProjectDeveloper = "PROJECT_DEVELOPER"
	// RoleNameProjectUser is the name of a Role that enables principals to
	// read, create, and manage Events for a Project.
	RoleNameProjectUser = "PROJECT_USER"
	// RoleNameReader is the name of a Role that enables principals to
	// list and read Projects, Users, and Service Accounts.
	RoleNameReader = "READER"

	// Special roles
	//
	// These are reserved for use by system components and are assignable to Users
	// and ServiceAccounts.

	// RoleNameObserver is the name of a Role that enables principals to updates
	// Worker and Job status based on observation of the underlying workload
	// execution substrate. This Role is exclusively for the use of the Observer
	// component.
	RoleNameObserver = "OBSERVER"
	// RoleNameScheduler is the name of a Role that enables principals to initiate
	// execution of a Worker or Job on the underlying workload execution
	// substrate. This Role is execlusively for the use of the Scheduler
	// component.
	RoleNameScheduler = "SCHEDULER"
	// RoleNameWorker is the name of a Role that enables principals to create new
	// Jobs. This Role is exclusively for the use of Workers.
	RoleNameWorker = "WORKER"
)

// RoleScopeGlobal represents an unbounded scope.
const RoleScopeGlobal = "*"

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount.
type Role struct {
	// Name is the name of a Role and has domain-specific meaning.
	Name string `json:"name" bson:"name"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope" bson:"scope"`
}

// RoleAdmin returns a Role that enables a principal to manage Users,
// ServiceAccounts, and globally scoped permissions for Users and
// ServiceAccounts.
func RoleAdmin() Role {
	return Role{
		Name:  RoleNameAdmin,
		Scope: RoleScopeGlobal,
	}
}

// RoleEventCreator returns a Role that enables a principal to create new Events
// having a Source field whose value matches that of the Scope field. This Role
// is useful for ServiceAccounts used for gateways.
func RoleEventCreator(eventSource string) Role {
	return Role{
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
		Name:  RoleNameProjectAdmin,
		Scope: projectID,
	}
}

// RoleProjectCreator returns a Role that enables a principal to create new
// Projects.
func RoleProjectCreator() Role {
	return Role{
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
		Name:  RoleNameProjectUser,
		Scope: projectID,
	}
}

// RoleReader returns a Role that enables a principal to list and read Projects,
// Events, Users, and Service Accounts.
func RoleReader() Role {
	return Role{
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
		Name:  RoleNameObserver,
		Scope: RoleScopeGlobal,
	}
}

// RoleScheduler returns a Role that enables a principal to initiate execution
// of Workers and Jobs on the underlying workload execution substrate. This Role
// is exclusively for the use of the Scheduler component.
func RoleScheduler() Role {
	return Role{
		Name:  RoleNameScheduler,
		Scope: RoleScopeGlobal,
	}
}

// RoleWorker returns a Role that enables a principal to create Jobs for the
// Event whose ID matches the Scope. This Role is exclusively for the use of
// Workers.
func RoleWorker(eventID string) Role {
	return Role{
		Name:  RoleNameWorker,
		Scope: eventID,
	}
}
