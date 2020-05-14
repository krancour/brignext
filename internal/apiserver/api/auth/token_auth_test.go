package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/stretchr/testify/require"
)

const testToken = "foooooooooooooooooooo"

func TestTokenAuthFilterWithHeaderMissing(t *testing.T) {
	a := NewTokenAuthFilter(nil, nil, false, testToken)
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
	a := NewTokenAuthFilter(nil, nil, false, testToken)
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
		func(context.Context, string) (Session, error) {
			return Session{}, &brignext.ErrSessionNotFound{}
		},
		nil,
		false,
		testToken,
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
		func(context.Context, string) (Session, error) {
			return Session{}, nil
		},
		func(context.Context, string) (brignext.User, error) {
			return brignext.User{}, nil
		},
		false,
		testToken,
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer foobar")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	a.Decorate(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		require.Nil(t, PincipalFromContext(r.Context()))
		require.Empty(t, SessionIDFromContext(r.Context()))
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

// TODO: Test expired token

// TODO: Test token for a root session when root sessions are not allowed

func TestTokenAuthFilterWithAuthenticatedSession(t *testing.T) {
	const sessionID = "foobar"
	a := NewTokenAuthFilter(
		func(context.Context, string) (Session, error) {
			return Session{
				ID:            sessionID,
				Authenticated: true,
				Expires:       time.Now().Add(time.Minute),
			}, nil
		},
		func(context.Context, string) (brignext.User, error) {
			return brignext.User{}, nil
		},
		false,
		testToken,
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer foobar")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	a.Decorate(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		require.NotNil(t, PincipalFromContext(r.Context()))
		require.Equal(t, sessionID, SessionIDFromContext(r.Context()))
	})(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, handlerCalled)
}
