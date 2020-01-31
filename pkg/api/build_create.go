package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/pkg/errors"
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

	build := &brignext.Build{}
	if err := json.Unmarshal(bodyBytes, build); err != nil {
		log.Println(
			errors.Wrap(err, "error unmarshaling body of create build request"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	// TODO: We should do some kind of validation!

	brigadeBuild := brignext.BrigNextBuildToBrigadeBuild(build)

	if err := s.oldProjectStore.CreateBuild(brigadeBuild); err != nil {
		log.Println(
			errors.Wrap(err, "error storing new build in old store"),
		)
		s.writeResponse(w, http.StatusBadRequest, responseEmptyJSON)
		return
	}

	// We DON'T write to the new store here. Gateways all still write to the old
	// store only. We'll use a controller to intercept all those old-style build
	// creations and echo them to the new store. Which means we don't have to
	// write to the new store here-- in fact, if we did, we'd end up with a
	// duplicate.

	// This is how we'll be certain to return the build ID that is assigned
	// by the write to the old store.
	build = brignext.BrigadeBuildToBrigNextBuild(brigadeBuild, build.ProjectName)

	responseBytes, err := json.Marshal(build)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error marshaling create build response"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	s.writeResponse(w, http.StatusOK, responseBytes)
}
