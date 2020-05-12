package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

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
