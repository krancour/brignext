package authn

import brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"

type root struct{}

func (r *root) Roles() []brignext.Role {
	return []brignext.Role{
		brignext.RoleEventCreator(brignext.RoleScopeGlobal),
		brignext.RoleProjectAdmin(brignext.RoleScopeGlobal),
		brignext.RoleProjectCreator(),
		brignext.RoleProjectDeveloper(brignext.RoleScopeGlobal),
		brignext.RoleProjectReader(brignext.RoleScopeGlobal),
		brignext.RoleProjectUser(brignext.RoleScopeGlobal),
		brignext.RoleServiceAccountManager(),
		brignext.RoleUserManager(),
	}
}

var Root = &root{}
