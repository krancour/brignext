package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/pkg/errors"
)

func (s *server) buildLogs(w http.ResponseWriter, r *http.Request) {
	buildID := mux.Vars(r)["id"]

	logEntryCh, err := s.logStore.StreamWorkerLogs(r.Context(), buildID)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error retrieving logs for build %q", buildID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")

	for logEntry := range logEntryCh {
		fmt.Fprint(w, logEntry.Message)
		w.(http.Flusher).Flush()
	}
}
