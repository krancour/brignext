package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"
	oldStorage "github.com/brigadecore/brigade/pkg/storage"
	"github.com/gorilla/mux"
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

	projectID := brigade.ProjectID(projectName)
	builds, err := s.projectStore.GetProjectBuilds(projectID)
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
				SkipRunningBuilds: !forceDelete,
			},
		); err != nil {
			log.Println(
				errors.Wrap(err, "error deleting build from new store"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
		if err := s.oldProjectStore.DeleteBuild(
			build.ID,
			oldStorage.DeleteBuildOptions{
				SkipRunningBuilds: !forceDelete,
			},
		); err != nil {
			log.Println(
				errors.Wrap(err, "error deleting build %q from old store"),
			)
			s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
			return
		}
	}

	// TODO: Cascade delete to associated jobs

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
