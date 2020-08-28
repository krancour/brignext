package authn

type observer struct{}

func (o *observer) Roles() []Role {
	return nil
}

var Observer = &observer{}
