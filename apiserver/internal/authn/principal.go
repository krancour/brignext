package authn

type Principal interface {
	Roles() []Role
}
