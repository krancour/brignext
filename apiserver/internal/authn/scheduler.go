package authn

import brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"

type scheduler struct{}

func (s *scheduler) Roles() []brignext.Role {
	return nil
}

var Scheduler = &scheduler{}
