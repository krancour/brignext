package projects

import (
	context "context"

	"github.com/pkg/errors"
)

func (p *projectsServer) CreateProject(
	ctx context.Context,
	req *CreateProjectRequest,
) (*CreateProjectResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &CreateProjectResponse{}

	project := WireProjectToBrigadeProject(req.Project)
	if err := p.projectStore.CreateProject(project); err != nil {
		return resp, errors.Wrapf(
			err,
			"error storing new project %q in new store",
			project.Name,
		)
	}
	if err := p.oldStore.CreateProject(project); err != nil {
		return resp, errors.Wrapf(
			err,
			"error storing new project %q in old store",
			project.Name,
		)
	}

	resp.Project = req.Project

	return resp, nil
}
