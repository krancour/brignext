package api

import "net/http"

func (s *server) healthCheck(
	w http.ResponseWriter,
	_ *http.Request,
) {
	// TODO: Test that critical connections are healthy?
	s.writeResponse(w, http.StatusOK, struct{}{})
}
