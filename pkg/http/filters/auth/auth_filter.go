package auth

import "net/http"

// AuthFilter is an interface to be implemented by components that can wrap a
// new http.HandlerFunc that handles authentication around another
// http.HandlerFunc.
type AuthFilter interface {
	// Decorate decorates one http.HandlerFunc with another
	Decorate(http.HandlerFunc) http.HandlerFunc
}
