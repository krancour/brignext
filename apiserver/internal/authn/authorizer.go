package authn

import "context"

// TODO: Do we need a whole interface? Or can we just get by with a function?

type Authorizer interface {
	SubjectHas(ctx context.Context, role Role) (bool, error)
}

type alwaysAuthorizer struct{}

func NewAlwaysAuthorizer() Authorizer {
	return &alwaysAuthorizer{}
}

func (a *alwaysAuthorizer) SubjectHas(context.Context, Role) (bool, error) {
	return true, nil
}

type neverAuthorizer struct{}

func NewNeverAuthorizer() Authorizer {
	return &neverAuthorizer{}
}

func (n *neverAuthorizer) SubjectHas(context.Context, Role) (bool, error) {
	return false, nil
}

type authorizer struct{}

func NewAuthorizer() Authorizer {
	return &neverAuthorizer{}
}

// TODO: Implement this
func (a *authorizer) SubjectHas(context.Context, Role) (bool, error) {
	return false, nil
}
