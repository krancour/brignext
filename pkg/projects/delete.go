package projects

import (
	context "context"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/pkg/errors"
)

func (p *projectsServer) DeleteProject(
	ctx context.Context,
	req *DeleteProjectRequest,
) (*DeleteProjectResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &DeleteProjectResponse{}

	projectID := brigade.ProjectID(req.ProjectName)
	if err := p.projectStore.DeleteProject(projectID); err != nil {
		return resp, errors.Wrapf(
			err,
			"error deleting project %q from new store",
			req.ProjectName,
		)
	}
	if err := p.oldStore.DeleteProject(
		brigade.ProjectID(req.ProjectName),
	); err != nil {
		return resp, errors.Wrapf(
			err,
			"error deleting project %q from old store",
			req.ProjectName,
		)
	}

	// TODO: Cascade delete to associated builds & jobs

	return &DeleteProjectResponse{}, nil
}
