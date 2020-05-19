package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

func (s *server) eventCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error reading body of create event request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	var validationResult *gojsonschema.Result
	if validationResult, err = gojsonschema.Validate(
		s.eventSchemaLoader,
		gojsonschema.NewBytesLoader(bodyBytes),
	); err != nil {
		log.Println(errors.Wrap(err, "error validating create event request"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !validationResult.Valid() {
		// TODO: Remove this and return validation errors to the client instead
		fmt.Println("-------------------------------------------------------------")
		for i, verr := range validationResult.Errors() {
			fmt.Printf("%d. %s\n", i, verr)
		}
		fmt.Println("-------------------------------------------------------------")
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	event := brignext.Event{}
	if err = json.Unmarshal(bodyBytes, &event); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of create event request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	eventRefList, err := s.service.CreateEvent(r.Context(), event)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(errors.Wrap(err, "error creating new event"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(eventRefList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create event response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseBytes)
	return
}

func (s *server) eventList(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	var eventList brignext.EventList
	var err error
	if projectID := r.URL.Query().Get("projectID"); projectID != "" {
		eventList, err = s.service.GetEventsByProject(r.Context(), projectID)
	} else {
		eventList, err = s.service.GetEvents(r.Context())
	}
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(errors.Wrap(err, "error retrieving events"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(eventList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling list events response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}

func (s *server) eventGet(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	id := mux.Vars(r)["id"]

	event, err := s.service.GetEvent(r.Context(), id)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error retrieving event %q", id),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(event)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling get event response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}

func (s *server) eventsCancel(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["id"]
	projectID := mux.Vars(r)["projectID"]

	cancelRunningStr := r.URL.Query().Get("cancelRunning")
	var cancelRunning bool
	if cancelRunningStr != "" {
		cancelRunning, _ =
			strconv.ParseBool(cancelRunningStr) // nolint: errcheck
	}

	if eventID != "" {
		eventRefList, err := s.service.CancelEvent(
			r.Context(),
			eventID,
			cancelRunning,
		)
		if err != nil {
			if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
				s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
				return
			}
			log.Println(
				errors.Wrapf(err, "error canceling event %q", eventID),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(eventRefList)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling cancel event response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusOK, responseBytes)
		return
	}

	eventRefList, err := s.service.CancelEventsByProject(
		r.Context(),
		projectID,
		cancelRunning,
	)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error canceling events for project %q", projectID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(eventRefList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling cancel event response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}

func (s *server) eventsDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["id"]
	projectID := mux.Vars(r)["projectID"]

	deletePendingStr := r.URL.Query().Get("deletePending")
	var deletePending bool
	if deletePendingStr != "" {
		deletePending, _ = strconv.ParseBool(deletePendingStr) // nolint: errcheck
	}

	deleteRunningStr := r.URL.Query().Get("deleteRunning")
	var deleteRunning bool
	if deleteRunningStr != "" {
		deleteRunning, _ =
			strconv.ParseBool(deleteRunningStr) // nolint: errcheck
	}

	if eventID != "" {
		eventRefList, err := s.service.DeleteEvent(
			r.Context(),
			eventID,
			deletePending,
			deleteRunning,
		)
		if err != nil {
			if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
				s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
				return
			}
			log.Println(
				errors.Wrapf(err, "error deleting event %q", eventID),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(eventRefList)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling delete event response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusOK, responseBytes)
		return
	}

	eventRefList, err := s.service.DeleteEventsByProject(
		r.Context(),
		projectID,
		deletePending,
		deleteRunning,
	)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrProjectNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(err, "error deleting events for project %q", projectID),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(eventRefList)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling delete event response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}

func (s *server) workerUpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["eventID"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(
				err,
				"error reading body of update event worker status request",
			),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if validationResult, err := gojsonschema.Validate(
		s.workerStatusSchemaLoader,
		gojsonschema.NewBytesLoader(bodyBytes),
	); err != nil {
		log.Println(
			errors.Wrap(err, "error validating update worker status request"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !validationResult.Valid() {
		// TODO: Remove this and return validation errors to the client instead
		fmt.Println("-------------------------------------------------------------")
		for i, verr := range validationResult.Errors() {
			fmt.Printf("%d. %s\n", i, verr)
		}
		fmt.Println("-------------------------------------------------------------")
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	status := brignext.WorkerStatus{}
	if err := json.Unmarshal(bodyBytes, &status); err != nil {
		log.Println(
			errors.Wrap(
				err,
				"error unmarshaling body of update event worker status request",
			),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if err :=
		s.service.UpdateWorkerStatus(
			r.Context(),
			eventID,
			status,
		); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error updating status on event %q worker",
				eventID,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)

}

func (s *server) workerLogs(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]

	streamStr := r.URL.Query().Get("stream")
	var stream bool
	if streamStr != "" {
		stream, _ = strconv.ParseBool(streamStr) // nolint: errcheck
	}

	initStr := r.URL.Query().Get("init")
	var init bool
	if initStr != "" {
		init, _ = strconv.ParseBool(initStr) // nolint: errcheck
	}

	if !stream {
		var logEntriesList brignext.LogEntryList
		var err error
		if init {
			logEntriesList, err = s.service.GetWorkerInitLogs(
				r.Context(),
				eventID,
			)
		} else {
			logEntriesList, err = s.service.GetWorkerLogs(
				r.Context(),
				eventID,
			)
		}
		if err != nil {
			if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
				s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
				return
			}
			log.Println(
				errors.Wrapf(
					err,
					"error retrieving event %q worker logs",
					eventID,
				),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(logEntriesList)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling get worker logs response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusOK, responseBytes)
		return
	}

	var logEntryCh <-chan brignext.LogEntry
	var err error
	if init {
		logEntryCh, err = s.service.StreamWorkerInitLogs(
			r.Context(),
			eventID,
		)
	} else {
		logEntryCh, err = s.service.StreamWorkerLogs(
			r.Context(),
			eventID,
		)
	}
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving log stream for event %q worker",
				eventID,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	for logEntry := range logEntryCh {
		logEntryBytes, err := json.Marshal(logEntry)
		if err != nil {
			log.Println(errors.Wrapf(err, "error unmarshaling log entry"))
			return
		}
		fmt.Fprint(w, string(logEntryBytes))
		w.(http.Flusher).Flush()
	}
}

func (s *server) jobUpdateStatus(
	w http.ResponseWriter,
	r *http.Request,
) {
	defer r.Body.Close() // nolint: errcheck

	eventID := mux.Vars(r)["eventID"]

	jobName := mux.Vars(r)["jobName"]

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(
				err,
				"error reading body of update job status request",
			),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if validationResult, err := gojsonschema.Validate(
		s.jobStatusSchemaLoader,
		gojsonschema.NewBytesLoader(bodyBytes),
	); err != nil {
		log.Println(errors.Wrap(err, "error validating update job status request"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !validationResult.Valid() {
		// TODO: Remove this and return validation errors to the client instead
		fmt.Println("-------------------------------------------------------------")
		for i, verr := range validationResult.Errors() {
			fmt.Printf("%d. %s\n", i, verr)
		}
		fmt.Println("-------------------------------------------------------------")
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	status := brignext.JobStatus{}
	if err := json.Unmarshal(bodyBytes, &status); err != nil {
		log.Println(
			errors.Wrap(
				err,
				"error unmarshaling body of update job status request",
			),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	if err :=
		s.service.UpdateJobStatus(
			r.Context(),
			eventID,
			jobName,
			status,
		); err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error updating status on event %q worker job %q",
				eventID,
				jobName,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)

}

func (s *server) jobLogs(w http.ResponseWriter, r *http.Request) {
	eventID := mux.Vars(r)["eventID"]
	jobName := mux.Vars(r)["jobName"]

	streamStr := r.URL.Query().Get("stream")
	var stream bool
	if streamStr != "" {
		stream, _ = strconv.ParseBool(streamStr) // nolint: errcheck
	}

	initStr := r.URL.Query().Get("init")
	var init bool
	if initStr != "" {
		init, _ = strconv.ParseBool(initStr) // nolint: errcheck
	}

	if !stream {
		var logEntriesList brignext.LogEntryList
		var err error
		if init {
			logEntriesList, err = s.service.GetJobInitLogs(
				r.Context(),
				eventID,
				jobName,
			)
		} else {
			logEntriesList, err = s.service.GetJobLogs(
				r.Context(),
				eventID,
				jobName,
			)
		}
		if err != nil {
			if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
				s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
				return
			} else if _, ok := errors.Cause(err).(*brignext.ErrJobNotFound); ok {
				s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
				return
			}
			log.Println(
				errors.Wrapf(
					err,
					"error retrieving event %q worker job %q logs",
					eventID,
					jobName,
				),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		responseBytes, err := json.Marshal(logEntriesList)
		if err != nil {
			log.Println(
				errors.Wrap(err, "error marshaling get job logs response"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}

		s.writeResponse(w, http.StatusOK, responseBytes)
		return
	}

	logEntryCh, err := s.service.StreamJobLogs(
		r.Context(),
		eventID,
		jobName,
	)
	if err != nil {
		if _, ok := errors.Cause(err).(*brignext.ErrEventNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		} else if _, ok := errors.Cause(err).(*brignext.ErrJobNotFound); ok {
			s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
			return
		}
		log.Println(
			errors.Wrapf(
				err,
				"error retrieving log stream for event %q worker job %q",
				eventID,
				jobName,
			),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	for logEntry := range logEntryCh {
		logEntryBytes, err := json.Marshal(logEntry)
		if err != nil {
			log.Println(errors.Wrapf(err, "error unmarshaling log entry"))
			return
		}
		fmt.Fprint(w, string(logEntryBytes))
		w.(http.Flusher).Flush()
	}
}
