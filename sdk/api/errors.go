package api

import (
	"encoding/json"
	"fmt"

	"github.com/krancour/brignext/v2/sdk/meta"
)

type ErrAuthentication struct {
	Reason string `json:"reason"`
}

func (e *ErrAuthentication) Error() string {
	return fmt.Sprintf("Could not authenticate the request: %s", e.Reason)
}

func (e ErrAuthentication) MarshalJSON() ([]byte, error) {
	type Alias ErrAuthentication
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "AuthenticationError",
			},
			Alias: (Alias)(e),
		},
	)
}

type ErrAuthorization struct{}

func (e *ErrAuthorization) Error() string {
	return "The request is not authorized."
}

func (e ErrAuthorization) MarshalJSON() ([]byte, error) {
	type Alias ErrAuthorization
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "AuthorizationError",
			},
			Alias: (Alias)(e),
		},
	)
}

type ErrBadRequest struct {
	Reason  string   `json:"reason"`
	Details []string `json:"details"`
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

func (e ErrBadRequest) MarshalJSON() ([]byte, error) {
	type Alias ErrBadRequest
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "BadRequestError",
			},
			Alias: (Alias)(e),
		},
	)
}

type ErrNotFound struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("%s %q not found.", e.Type, e.ID)
}

func (e ErrNotFound) MarshalJSON() ([]byte, error) {
	type Alias ErrNotFound
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "NotFoundError",
			},
			Alias: (Alias)(e),
		},
	)
}

type ErrConflict struct {
	Type   string `json:"type"`
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

func (e *ErrConflict) Error() string {
	return e.Reason
}

func (e ErrConflict) MarshalJSON() ([]byte, error) {
	type Alias ErrConflict
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ConflictError",
			},
			Alias: (Alias)(e),
		},
	)
}

type ErrInternalServer struct{}

func (e *ErrInternalServer) Error() string {
	return "An internal server error occurred."
}

func (e ErrInternalServer) MarshalJSON() ([]byte, error) {
	type Alias ErrInternalServer
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "InternalServerError",
			},
			Alias: (Alias)(e),
		},
	)
}

type ErrNotSupported struct {
	Details string `json:"reason"`
}

func (e *ErrNotSupported) Error() string {
	return e.Details
}

func (e ErrNotSupported) MarshalJSON() ([]byte, error) {
	type Alias ErrNotSupported
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "NotSupportedError",
			},
			Alias: (Alias)(e),
		},
	)
}
