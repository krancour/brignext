package authx

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

var Root = &root{}
