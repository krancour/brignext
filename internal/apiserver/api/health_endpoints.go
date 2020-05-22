package api

import "net/http"

func (s *server) healthCheck(
	w http.ResponseWriter,
	r *http.Request,
) {
	s.serveAPIRequest(apiRequest{
		w: w,
		r: r,
		endpointLogic: func() (interface{}, error) {
			// TODO: Test that critical connections are healthy?
			return struct{}{}, nil
		},
		successCode: http.StatusOK,
	})
}
