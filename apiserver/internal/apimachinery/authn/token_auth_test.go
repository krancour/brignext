package authn

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authn"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/stretchr/testify/require"
)

const testSchedulerToken = "foooooooooooooooooooo"
const testObserverToken = "baaaaaaaaaaaaaaaaaaaar"

func TestTokenAuthFilterWithHeaderMissing(t *testing.T) {
	a := NewTokenAuthFilter(
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
		func(context.Context, string) (authn.Session, error) {
			return authn.Session{}, &brignext.ErrNotFound{}
		},
		nil,
		nil,
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
		func(context.Context, string) (authn.Session, error) {
			return authn.Session{}, nil
		},
		nil,
		func(context.Context, string) (authn.User, error) {
			return authn.User{}, nil
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
		require.Nil(t, authn.PincipalFromContext(r.Context()))
		require.Empty(t, authn.SessionIDFromContext(r.Context()))
	})(rr, req)
	require.Equal(t, http.StatusUnauthorized, rr.Code)
	require.False(t, handlerCalled)
}

func TestTokenAuthFilterWithAuthenticatedSession(t *testing.T) {
	const sessionID = "foobar"
	a := NewTokenAuthFilter(
		func(context.Context, string) (authn.Session, error) {
			now := time.Now()
			expiry := now.Add(time.Minute)
			return authn.Session{
				ObjectMeta: meta.ObjectMeta{
					ID: sessionID,
				},
				Authenticated: &now,
				Expires:       &expiry,
			}, nil
		},
		nil,
		func(context.Context, string) (authn.User, error) {
			return authn.User{}, nil
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
		require.NotNil(t, authn.PincipalFromContext(r.Context()))
		require.Equal(t, sessionID, authn.SessionIDFromContext(r.Context()))
	})(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)
	require.True(t, handlerCalled)
}
