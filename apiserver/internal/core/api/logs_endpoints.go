package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/lib/apimachinery"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

type logsEndpoints struct {
	*apimachinery.BaseEndpoints
	service core.EventsService
}

func NewLogsEndpoints(
	baseEndpoints *apimachinery.BaseEndpoints,
	service core.EventsService,
) apimachinery.Endpoints {
	// nolint: lll
	return &logsEndpoints{
		BaseEndpoints: baseEndpoints,
		service:       service,
	}
}

func (l *logsEndpoints) Register(router *mux.Router) {
	// Stream logs
	router.HandleFunc(
		"/v2/events/{id}/logs",
		l.TokenAuthFilter.Decorate(l.stream),
	).Methods(http.MethodGet)
}

func (l *logsEndpoints) stream(
	w http.ResponseWriter,
	r *http.Request,
) {
	id := mux.Vars(r)["id"]
	// nolint: errcheck
	follow, _ := strconv.ParseBool(r.URL.Query().Get("follow"))

	selector := core.LogsSelector{
		Job:       r.URL.Query().Get("job"),
		Container: r.URL.Query().Get("container"),
	}
	opts := core.LogStreamOptions{
		Follow: follow,
	}

	logEntryCh, err := l.service.StreamLogs(r.Context(), id, selector, opts)
	if err != nil {
		if _, ok := errors.Cause(err).(*meta.ErrNotFound); ok {
			l.WriteAPIResponse(w, http.StatusNotFound, errors.Cause(err))
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving log stream for event %q", id),
		)
		l.WriteAPIResponse(
			w,
			http.StatusInternalServerError,
			&meta.ErrInternalServer{},
		)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.(http.Flusher).Flush()
	for logEntry := range logEntryCh {
		logEntryBytes, err := json.Marshal(logEntry)
		if err != nil {
			log.Println(errors.Wrapf(err, "error marshaling log entry"))
			return
		}
		fmt.Fprint(w, string(logEntryBytes))
		w.(http.Flusher).Flush()
	}
}
