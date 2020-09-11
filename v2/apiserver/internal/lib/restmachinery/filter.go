package restmachinery

import "net/http"

// Filter is an interface to be implemented by components that can wrap a
// new http.HandlerFunc that handles authentication around another
// http.HandlerFunc.
type Filter interface {
	// Decorate decorates one http.HandlerFunc with another
	Decorate(http.HandlerFunc) http.HandlerFunc
}
