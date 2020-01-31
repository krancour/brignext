package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/stretchr/testify/assert"
)

func TestBasicAuthFilterWithHeaderMissing(t *testing.T) {
	a := NewBasicAuthFilter(nil)
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

func TestBasicAuthFilterWithHeaderNotBasic(t *testing.T) {
	a := NewBasicAuthFilter(nil)
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

func TestBasicAuthFilterWithUsernamePasswordNotBase64(t *testing.T) {
	a := NewBasicAuthFilter(nil)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	req.Header.Add("Authorization", "Basic foo")
	rr := httptest.NewRecorder()
	handlerCalled := false
	a.GetHandler(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled)
}

func TestBasicAuthFilterWithUsernamePasswordInvalid(t *testing.T) {
	a := NewBasicAuthFilter(
		func(string, string) (*brignext.User, error) {
			return nil, nil
		},
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	req.SetBasicAuth("username", "password")
	rr := httptest.NewRecorder()
	handlerCalled := false
	a.GetHandler(func(http.ResponseWriter, *http.Request) {
		handlerCalled = true
	})(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.False(t, handlerCalled)
}

func TestBasicAuthFilterWithUsernamePasswordValid(t *testing.T) {
	a := NewBasicAuthFilter(
		func(string, string) (*brignext.User, error) {
			return &brignext.User{}, nil
		},
	)
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	assert.Nil(t, err)
	req.SetBasicAuth("username", "password")
	rr := httptest.NewRecorder()
	var handlerCalled bool
	var contextDecorated bool
	a.GetHandler(func(_ http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		if UserFromContext(r.Context()) != nil {
			contextDecorated = true
		}
	})(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.True(t, handlerCalled)
	assert.True(t, contextDecorated)
}
