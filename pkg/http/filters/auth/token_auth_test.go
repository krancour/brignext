package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/stretchr/testify/require"
)

func TestTokenAuthFilterWithHeaderMissing(t *testing.T) {
	a := NewTokenAuthFilter(nil, nil, false)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	rr := httptest.NewRecorder()
	handlerCalled := false
	a.Decorate(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithHeaderNotBearer(t *testing.T) {
	a := NewTokenAuthFilter(nil, nil, false)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Digest foo")
	rr := httptest.NewRecorder()
	handlerCalled := false
	a.Decorate(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithTokenInvalid(t *testing.T) {
	a := NewTokenAuthFilter(
		func(string) (*brignext.Session, error) {
			return nil, nil
		},
		nil,
		false,
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", "foo"),
	)
	rr := httptest.NewRecorder()
	handlerCalled := false
	a.Decorate(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithUnauthenticatedSession(t *testing.T) {
	a := NewTokenAuthFilter(
		func(string) (*brignext.Session, error) {
			return &brignext.Session{}, nil
		},
		func(string) (*brignext.User, error) {
			return &brignext.User{}, nil
		},
		false,
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer foobar")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	a.Decorate(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		if UserFromContext(r.Context()) != nil &&
			SessionIDFromContext(r.Context()) != "" {
		}
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

// TODO: Test expired token

// TODO: Test token for a root session when root sessions are not allowed

func TestTokenAuthFilterWithAuthenticatedSession(t *testing.T) {
	const sessionID = "foobar"
	a := NewTokenAuthFilter(
		func(string) (*brignext.Session, error) {
			return &brignext.Session{
				ID:            sessionID,
				Authenticated: true,
				ExpiresAt:     time.Now().Add(time.Minute),
			}, nil
		},
		func(string) (*brignext.User, error) {
			return &brignext.User{}, nil
		},
		false,
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer foobar")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	var contextDecorated bool
	a.Decorate(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		if UserFromContext(r.Context()) != nil &&
			SessionIDFromContext(r.Context()) == sessionID {
			contextDecorated = true
		}
	})(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, handlerCalled)
	require.True(t, contextDecorated)
}
