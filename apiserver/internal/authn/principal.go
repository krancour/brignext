package authn

import brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"

type Principal interface {
	Roles() []brignext.Role
}
