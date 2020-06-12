package brignext

import (
	"fmt"

	"github.com/krancour/brignext/v2/internal/pkg/meta"
)

type ErrAuthentication struct {
	meta.TypeMeta `json:",inline"`
	Reason        string `json:"reason"`
}

func NewErrAuthentication(reason string) *ErrAuthentication {
	return &ErrAuthentication{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "AuthenticationError",
		},
		Reason: reason,
	}
}

func (e *ErrAuthentication) Error() string {
	return fmt.Sprintf("Could not authenticate the request: %s", e.Reason)
}

type ErrAuthorization struct {
	meta.TypeMeta `json:",inline"`
}

func NewErrAuthorization() *ErrAuthorization {
	return &ErrAuthorization{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "AuthorizationError",
		},
	}
}

func (e *ErrAuthorization) Error() string {
	return "The request is not authorized."
}

type ErrBadRequest struct {
	meta.TypeMeta `json:",inline"`
	Reason        string   `json:"reason"`
	Details       []string `json:"details"`
}

func NewErrBadRequest(reason string, details ...string) *ErrBadRequest {
	return &ErrBadRequest{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "BadRequestError",
		},
		Reason:  reason,
		Details: details,
	}
}

func (e *ErrBadRequest) Error() string {
	if len(e.Details) == 0 {
		return fmt.Sprintf("Bad request: %s", e.Reason)
	}
	msg := fmt.Sprintf("Bad request: %s:", e.Reason)
	for i, detail := range e.Details {
		msg = fmt.Sprintf("%s\n  %d. %s", msg, i, detail)
	}
	return msg
}

type ErrNotFound struct {
	meta.TypeMeta `json:",inline"`
	Type          string `json:"type"`
	ID            string `json:"id"`
}

func NewErrNotFound(tipe, id string) *ErrNotFound {
	return &ErrNotFound{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "NotFoundError",
		},
		Type: tipe,
		ID:   id,
	}
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s %q not found.", e.Type, e.ID)
}

type ErrConflict struct {
	meta.TypeMeta `json:",inline"`
	Type          string `json:"type"`
	ID            string `json:"id"`
	Reason        string `json:"reason"`
}

func NewErrConflict(tipe string, id string, reason string) *ErrConflict {
	return &ErrConflict{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "ConflictError",
		},
		Type:   tipe,
		ID:     id,
		Reason: reason,
	}
}

func (e *ErrConflict) Error() string {
	return e.Reason
}

type ErrInternalServer struct {
	meta.TypeMeta `json:",inline"`
}

func NewErrInternalServer() *ErrInternalServer {
	return &ErrInternalServer{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "InternalServerError",
		},
	}
}

func (e *ErrInternalServer) Error() string {
	return "An internal server error occurred."
}

type ErrNotSupported struct {
	meta.TypeMeta `json:",inline"`
	Details       string `json:"reason"`
}

func NewErrNotSupported(details string) *ErrNotSupported {
	return &ErrNotSupported{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "NotSupportedError",
		},
		Details: details,
	}
}

func (e *ErrNotSupported) Error() string {
	return e.Details
}
