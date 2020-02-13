package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
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

	if validationResult, err := gojsonschema.Validate(
		s.eventSchemaLoader,
		gojsonschema.NewBytesLoader(bodyBytes),
	); err != nil {
		log.Println(errors.Wrap(err, "error validating request"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !validationResult.Valid() {
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	event := brignext.Event{}
	if err := json.Unmarshal(bodyBytes, &event); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of create event request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	id, ok, err := s.service.CreateEvent(r.Context(), event)
	if err != nil {
		log.Println(errors.Wrap(err, "error creating new event"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	} else if !ok {
		s.writeResponse(w, http.StatusNotFound, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			ID string `json:"id"`
		}{
			ID: id,
		},
	)
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
