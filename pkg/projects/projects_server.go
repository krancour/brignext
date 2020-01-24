package projects

import (
	oldStorage "github.com/brigadecore/brigade/pkg/storage"
	"github.com/krancour/brignext/pkg/storage"
)

type projectsServer struct {
	oldStore     oldStorage.Store
	projectStore storage.ProjectStore
}

func NewServer(
	oldStore oldStorage.Store,
	projectStore storage.ProjectStore,
) ProjectsServer {
	return &projectsServer{
		oldStore:     oldStore,
		projectStore: projectStore,
	}
}
