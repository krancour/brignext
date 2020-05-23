package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

type healthEndpoints struct {
	*baseEndpoints
}

func (h *healthEndpoints) register(router *mux.Router) {
	// Health check
	router.HandleFunc(
		"/healthz",
		h.check, // No filters applied to this request
	).Methods(http.MethodGet)
}

func (h *healthEndpoints) check(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			// TODO: Test that critical connections are healthy?
			return struct{}{}, nil
		},
		successCode: http.StatusOK,
	})
}
