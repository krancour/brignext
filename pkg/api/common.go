package api

import (
	"log"
	"net/http"

	"github.com/pkg/errors"
)

func (s *server) writeResponse(
	w http.ResponseWriter,
	statusCode int,
	responseBody []byte,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if _, err := w.Write(responseBody); err != nil {
		log.Println(
			errors.Wrap(err, "api server error: error writing response"),
		)
	}
}
