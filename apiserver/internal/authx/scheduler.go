package authx

type scheduler struct{}

func (s *scheduler) Roles() []Role {
	return []Role{
		RoleScheduler(),
	}
}

var Scheduler = &scheduler{}
