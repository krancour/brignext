package authx

type Principal interface {
	Roles() []Role
}
