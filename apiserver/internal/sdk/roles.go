package sdk

type RoleName string

const (
	RoleNameEventCreator          RoleName = "EVENT_CREATOR"
	RoleNameProjectAdmin          RoleName = "PROJECT_ADMIN"
	RoleNameProjectCreator        RoleName = "PROJECT_CREATOR"
	RoleNameProjectDeveloper      RoleName = "PROJECT_DEVELOPER"
	RoleNameProjectReader         RoleName = "PROJECT_READER"
	RoleNameProjectUser           RoleName = "PROJECT_USER"
	RoleNameServiceAccountManager RoleName = "SERVICE_ACCOUNT_MANAGER"
	RoleNameUserManager           RoleName = "USER_MANAGER"
)

type RoleScope string

const RoleScopeGlobal RoleScope = "*"

type Role struct {
	Name  RoleName  `json:"name" bson:"name"`
	Scope RoleScope `json:"scope" bson:"scope"`
}

func RoleEventCreator(scope RoleScope) Role {
	return Role{
		Name:  RoleNameEventCreator,
		Scope: scope,
	}
}

func RoleProjectAdmin(scope RoleScope) Role {
	return Role{
		Name:  RoleNameProjectAdmin,
		Scope: scope,
	}
}

func RoleProjectCreator() Role {
	return Role{
		Name:  RoleNameProjectCreator,
		Scope: RoleScopeGlobal,
	}
}

func RoleProjectDeveloper(scope RoleScope) Role {
	return Role{
		Name:  RoleNameProjectDeveloper,
		Scope: scope,
	}
}

func RoleProjectReader(scope RoleScope) Role {
	return Role{
		Name:  RoleNameProjectReader,
		Scope: scope,
	}
}

func RoleProjectUser(scope RoleScope) Role {
	return Role{
		Name:  RoleNameProjectUser,
		Scope: scope,
	}
}

func RoleServiceAccountManager() Role {
	return Role{
		Name:  RoleNameServiceAccountManager,
		Scope: RoleScopeGlobal,
	}
}

func RoleUserManager() Role {
	return Role{
		Name:  RoleNameUserManager,
		Scope: RoleScopeGlobal,
	}
}
