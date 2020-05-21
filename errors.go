package brignext

import "fmt"

type ErrAuthentication struct {
	TypeMeta `json:",inline"`
	Reason   string `json:"reason"`
}

func NewErrAuthentication(reason string) *ErrAuthentication {
	return &ErrAuthentication{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "AuthenticationError",
		},
		Reason: reason,
	}
}

func (e *ErrAuthentication) Error() string {
	return fmt.Sprintf("could not authenticate the request: %s", e.Reason)
}

type ErrAuthorization struct {
	TypeMeta `json:",inline"`
	Reason   string `json:"reason"`
}

func NewErrAuthorization(reason string) *ErrAuthorization {
	return &ErrAuthorization{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "AuthorizationError",
		},
		Reason: reason,
	}
}

func (e *ErrAuthorization) Error() string {
	return fmt.Sprintf("could not authorize the request: %s", e.Reason)
}

type ErrBadRequest struct {
	TypeMeta `json:",inline"`
	Reason   string   `json:"reason"`
	Details  []string `json:"details"`
}

func NewErrBadRequest(reason string, details ...string) *ErrBadRequest {
	return &ErrBadRequest{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "BadRequestError",
		},
		Reason:  reason,
		Details: details,
	}
}

func (e *ErrBadRequest) Error() string {
	if len(e.Details) == 0 {
		return fmt.Sprintf("bad request: %s", e.Reason)
	}
	msg := fmt.Sprintf("bad request: %s:", e.Reason)
	for i, detail := range e.Details {
		msg = fmt.Sprintf("%s\n  %d. %s", msg, i, detail)
	}
	return msg
}

type ErrNotFound struct {
	TypeMeta `json:",inline"`
	Type     string `json:"type"`
	ID       string `json:"id"`
}

func NewErrNotFound(tipe, id string) *ErrNotFound {
	return &ErrNotFound{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "NotFoundError",
		},
		Type: tipe,
		ID:   id,
	}
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s %q not found", e.Type, e.ID)
}

type ErrConflict struct {
	TypeMeta `json:",inline"`
	Type     string `json:"type"`
	ID       string `json:"id"`
}

func NewErrConflict(tipe string, id string) *ErrConflict {
	return &ErrConflict{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "IDConflictError",
		},
		Type: tipe,
		ID:   id,
	}
}

func (e *ErrConflict) Error() string {
	return fmt.Sprintf("a %s with the ID %q already exists", e.Type, e.ID)
}

type ErrInternalServer struct {
	TypeMeta `json:",inline"`
}

func NewErrInternalServer() *ErrInternalServer {
	return &ErrInternalServer{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "InternalServerError",
		},
	}
}

func (e *ErrInternalServer) Error() string {
	return "an internal server error occurred"
}
