package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

func (s *server) buildCreate(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error reading body of create build request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	build := brignext.Build{}
	if err := json.Unmarshal(bodyBytes, &build); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of create build request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}
	build.ID = uuid.NewV4().String()

	// TODO: We should do some kind of validation!

	if err := s.projectStore.CreateBuild(build); err != nil {
		log.Println(errors.Wrap(err, "error creating new build"))
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	responseBytes, err := json.Marshal(
		struct {
			ID string `json:"id"`
		}{
			ID: build.ID,
		},
	)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create build response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusCreated, responseBytes)
	return
}
