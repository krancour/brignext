package authx

import "context"

type PrincipalType string

const (
	PrincipalTypeUser           PrincipalType = "USER"
	PrincipalTypeServiceAccount PrincipalType = "SERVICE_ACCOUNT"
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