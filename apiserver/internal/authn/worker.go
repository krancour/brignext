package authn

type worker struct{}

func (w *worker) Roles() []Role {
	return nil
}

var Worker = &worker{}
