package api

import (
	"log"
	"net/http"
	"strconv"

	"github.com/brigadecore/brigade/pkg/storage"
	oldStorage "github.com/brigadecore/brigade/pkg/storage"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

func (s *server) buildDelete(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close() // nolint: errcheck

	buildID := mux.Vars(r)["id"]
	forceDeleteStr := r.URL.Query().Get("force")
	var forceDelete bool
	if forceDeleteStr != "" {
		forceDelete, _ = strconv.ParseBool(forceDeleteStr) // nolint: errcheck
	}

	if err := s.projectStore.DeleteBuild(
		buildID,
		storage.DeleteBuildOptions{
			SkipRunningBuilds: !forceDelete,
		},
	); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting build"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	if err := s.oldProjectStore.DeleteBuild(
		buildID,
		oldStorage.DeleteBuildOptions{
			SkipRunningBuilds: !forceDelete,
		},
	); err != nil {
		log.Println(
			errors.Wrap(err, "error deleting build from old store"),
		)
		s.writeResponse(w, http.StatusInternalServerError, responseEmptyJSON)
		return
	}

	// TODO: Cascade delete to associated jobs

	s.writeResponse(w, http.StatusOK, responseEmptyJSON)
}
