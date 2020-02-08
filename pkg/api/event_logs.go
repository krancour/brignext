package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/pkg/errors"
)

func (s *server) eventLogs(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	logEntryCh, err := s.logStore.StreamWorkerLogs(r.Context(), id)
	if err != nil {
		log.Println(
			errors.Wrapf(err, "error retrieving logs for event %q", id),
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
