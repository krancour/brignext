package authn

type observer struct{}

func (o *observer) Roles() []Role {
	return []Role{
		RoleObserver(),
	}
}

var Observer = &observer{}
