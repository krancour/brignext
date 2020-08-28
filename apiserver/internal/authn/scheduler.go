package authn

type scheduler struct{}

func (s *scheduler) Roles() []Role {
	return nil
}

var Scheduler = &scheduler{}
