package authn

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/stretchr/testify/require"
)

const testSchedulerToken = "foooooooooooooooooooo"
const testObserverToken = "baaaaaaaaaaaaaaaaaaaar"

func TestTokenAuthFilterWithHeaderMissing(t *testing.T) {
	a := NewTokenAuthFilter(
		nil,
		nil,
		nil,
		nil,
		false,
		testSchedulerToken,
		testObserverToken,
	)
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
	a := NewTokenAuthFilter(
		nil,
		nil,
		nil,
		nil,
		false,
		testSchedulerToken,
		testObserverToken,
	)
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
		func(context.Context, string) (authx.Session, error) {
			return authx.Session{}, &meta.ErrNotFound{}
		},
		func(ctx context.Context, token string) (core.Event, error) {
			return core.Event{}, &meta.ErrNotFound{}
		},
		nil,
		func(ctx context.Context, token string) (authx.ServiceAccount, error) {
			return authx.ServiceAccount{}, &meta.ErrNotFound{}
		},
		false,
		testSchedulerToken,
		testObserverToken,
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
		func(context.Context, string) (authx.Session, error) {
			return authx.Session{}, nil
		},
		func(ctx context.Context, token string) (core.Event, error) {
			return core.Event{}, &meta.ErrNotFound{}
		},
		func(context.Context, string) (authx.User, error) {
			return authx.User{}, nil
		},
		func(ctx context.Context, token string) (authx.ServiceAccount, error) {
			return authx.ServiceAccount{}, &meta.ErrNotFound{}
		},
		false,
		testSchedulerToken,
		testObserverToken,
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer foobar")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	a.Decorate(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		require.Nil(t, authx.PincipalFromContext(r.Context()))
		require.Empty(t, authx.SessionIDFromContext(r.Context()))
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithAuthenticatedSession(t *testing.T) {
	const sessionID = "foobar"
	a := NewTokenAuthFilter(
		func(context.Context, string) (authx.Session, error) {
			now := time.Now()
			expiry := now.Add(time.Minute)
			return authx.Session{
				ObjectMeta: meta.ObjectMeta{
					ID: sessionID,
				},
				Authenticated: &now,
				Expires:       &expiry,
			}, nil
		},
		func(ctx context.Context, token string) (core.Event, error) {
			return core.Event{}, &meta.ErrNotFound{}
		},
		func(context.Context, string) (authx.User, error) {
			return authx.User{}, nil
		},
		func(ctx context.Context, token string) (authx.ServiceAccount, error) {
			return authx.ServiceAccount{}, &meta.ErrNotFound{}
		},
		false,
		testSchedulerToken,
		testObserverToken,
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", "Bearer foobar")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	a.Decorate(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		require.NotNil(t, authx.PincipalFromContext(r.Context()))
		require.Equal(t, sessionID, authx.SessionIDFromContext(r.Context()))
	})(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, handlerCalled)
}
