package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/stretchr/testify/assert"
)

const testToken = "token"

func TestTokenAuthFilterWithHeaderMissing(t *testing.T) {
	a := NewTokenAuthFilter(nil)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	rr := httptest.NewRecorder()
	handlerCalled := false
	a.GetHandler(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled)
}

func TestTokenAuthFilterWithHeaderNotBearer(t *testing.T) {
	a := NewTokenAuthFilter(nil)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	req.Header.Add("Authorization", "Digest foo")
	rr := httptest.NewRecorder()
	handlerCalled := false
	a.GetHandler(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled)
}

func TestTokenAuthFilterWithTokenInvalid(t *testing.T) {
	a := NewTokenAuthFilter(
		func(string) (*brignext.User, error) {
			return nil, nil
		},
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	req.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", "foo"),
	)
	rr := httptest.NewRecorder()
	handlerCalled := false
	a.GetHandler(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled)
}

func TestTokenAuthFilterWithTokenValid(t *testing.T) {
	a := NewTokenAuthFilter(
		func(string) (*brignext.User, error) {
			return &brignext.User{}, nil
		},
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	req.Header.Add(
		"Authorization",
		fmt.Sprintf("Bearer %s", testToken),
	)
	rr := httptest.NewRecorder()
	var handlerCalled bool
	var contextDecorated bool
	a.GetHandler(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		if UserFromContext(r.Context()) != nil &&
			UserTokenFromContext(r.Context()) != "" {
			contextDecorated = true
		}
	})(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	assert.True(t, contextDecorated)
}
