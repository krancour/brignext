package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
)

func (s *server) buildDeleteAll(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	projectName := mux.Vars(r)["projectName"]
	forceDeleteStr := r.URL.Query().Get("force")
	var forceDelete bool
	if forceDeleteStr != "" {
		forceDelete, _ = strconv.ParseBool(forceDeleteStr) // nolint: errcheck
	}

	builds, err := s.projectStore.GetBuildsByProjectName(projectName)
	if err != nil {
		log.Println(
			errors.Wrap(err, "error retrieving builds for project"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	for _, build := range builds {
		if err := s.projectStore.DeleteBuild(
			build.ID,
			storage.DeleteBuildOptions{
				DeleteRunningBuilds: forceDelete,
			},
		); err != nil {
			log.Println(
				errors.Wrap(err, "error deleting build"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
	}

	// TODO: Cascade delete to associated jobs

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
