package authx

type worker struct {
	eventID string
}

func (w *worker) Roles() []Role {
	return []Role{
		RoleWorker(w.eventID),
	}
}

func Worker(eventID string) Principal {
	return &worker{
		eventID: eventID,
	}
}
