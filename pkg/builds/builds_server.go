package builds

import (
	oldStorage "github.com/brigadecore/brigade/pkg/storage"
	"github.com/krancour/brignext/pkg/storage"
)

type buildsServer struct {
	oldStore     oldStorage.Store
	projectStore storage.ProjectStore
	logStore     storage.LogStore
}

func NewServer(
	oldStore oldStorage.Store,
	projectStore storage.ProjectStore,
	logStore storage.LogStore,
) BuildsServer {
	return &buildsServer{
		oldStore:     oldStore,
		projectStore: projectStore,
		logStore:     logStore,
	}
}
