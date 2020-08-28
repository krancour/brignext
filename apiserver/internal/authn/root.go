package authn

type root struct{}

func (r *root) Roles() []Role {
	return []Role{
		RoleEventCreator(RoleScopeGlobal),
		RoleProjectAdmin(RoleScopeGlobal),
		RoleProjectCreator(),
		RoleProjectDeveloper(RoleScopeGlobal),
		RoleProjectReader(RoleScopeGlobal),
		RoleProjectUser(RoleScopeGlobal),
		RoleServiceAccountManager(),
		RoleUserManager(),
	}
}

var Root = &root{}
