package authx

import "context"

// PrincipalType is a type whose values can be used to disambiguate one type of
// principal from another. For instance, when assigning a Role to a pincipal
// via a RoleAssignment, a PrincipalType field is used to indicate whether the
// value of the PrincipalID field reflects a User ID or a ServiceAccount ID.
type PrincipalType string

const (
	// PrincipalTypeServiceAccount represents a principal that is a
	// ServiceAccount.
	PrincipalTypeServiceAccount PrincipalType = "SERVICE_ACCOUNT"
	// PrincipalTypeUser represents a principal that is a User.
	PrincipalTypeUser PrincipalType = "USER"
)

var (
	Observer  = &observer{}
	Root      = &root{}
	Scheduler = &scheduler{}
)

type Principal interface {
	Roles() []Role
}

type observer struct{}

func (o *observer) Roles() []Role {
	return []Role{
		RoleObserver(),
	}
}

type root struct{}

func (r *root) Roles() []Role {
	return []Role{
		RoleAdmin(),
		RoleEventCreator(RoleScopeGlobal),
		RoleProjectAdmin(RoleScopeGlobal),
		RoleProjectCreator(),
		RoleProjectDeveloper(RoleScopeGlobal),
		RoleProjectUser(RoleScopeGlobal),
		RoleReader(),
	}
}

type scheduler struct{}

func (s *scheduler) Roles() []Role {
	return []Role{
		RoleScheduler(),
	}
}

type worker struct {
	eventID string
}

func (w *worker) Roles() []Role {
	return []Role{
		RoleWorker(w.eventID),
	}
}

func Worker(eventID string) Principal {
	return &worker{
		eventID: eventID,
	}
}

type principalContextKey struct{}

func ContextWithPrincipal(
	ctx context.Context,
	principal Principal,
) context.Context {
	return context.WithValue(
		ctx,
		principalContextKey{},
		principal,
	)
}

func PincipalFromContext(ctx context.Context) Principal {
	return ctx.Value(principalContextKey{}).(Principal)
}
