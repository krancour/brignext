package brignext

import "fmt"

type ErrUserNotFound struct {
	ID string
}

func (e *ErrUserNotFound) Error() string {
	return fmt.Sprintf("user %q not found", e.ID)
}

type ErrUserIDConflict struct {
	ID string
}

func (e *ErrUserIDConflict) Error() string {
	return fmt.Sprintf("a user with the ID %q already exists", e.ID)
}

type ErrServiceAccountNotFound struct {
	ID string
}

func (e *ErrServiceAccountNotFound) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("service account %q not found", e.ID)
	}
	return "service account not found"
}

type ErrServiceAccountIDConflict struct {
	ID string
}

func (e *ErrServiceAccountIDConflict) Error() string {
	return fmt.Sprintf("a service account with the ID %q already exists", e.ID)
}

type ErrSessionNotFound struct {
	ID string
}

func (e *ErrSessionNotFound) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("session %q not found", e.ID)
	}
	return "session not found"
}

type ErrProjectNotFound struct {
	ID string
}

func (e *ErrProjectNotFound) Error() string {
	return fmt.Sprintf("project %q not found", e.ID)
}

type ErrProjectIDConflict struct {
	ID string
}

func (e *ErrProjectIDConflict) Error() string {
	return fmt.Sprintf("a project with the ID %q already exists", e.ID)
}

type ErrEventNotFound struct {
	ID string
}

func (e *ErrEventNotFound) Error() string {
	return fmt.Sprintf("event %q not found", e.ID)
}

type ErrWorkerNotFound struct {
	EventID    string
	WorkerName string
}

func (e *ErrWorkerNotFound) Error() string {
	return fmt.Sprintf(
		"worker %q not found for event %q",
		e.WorkerName,
		e.EventID,
	)
}

type ErrJobNotFound struct {
	EventID    string
	WorkerName string
	JobName    string
}

func (e *ErrJobNotFound) Error() string {
	return fmt.Sprintf(
		"worker %q job %q not found for event %q",
		e.WorkerName,
		e.JobName,
		e.EventID,
	)
}
