package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/internal/apiserver/service"
)

type healthEndpoints struct {
	*baseEndpoints
	service service.Service
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
			if err := h.service.CheckHealth(r.Context()); err != nil {
				return nil, err
			}
			return struct{}{}, nil
		},
		successCode: http.StatusOK,
	})
}
