package authn

import brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"

type observer struct{}

func (o *observer) Roles() []brignext.Role {
	return nil
}

var Observer = &observer{}
