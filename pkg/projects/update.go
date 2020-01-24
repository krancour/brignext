package projects

import (
	context "context"

	"github.com/pkg/errors"
)

func (p *projectsServer) UpdateProject(
	ctx context.Context,
	req *UpdateProjectRequest,
) (*UpdateProjectResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &UpdateProjectResponse{}

	project := WireProjectToBrigadeProject(req.Project)

	if err := p.projectStore.UpdateProject(project); err != nil {
		return resp, errors.Wrapf(
			err,
			"error updating project %q in new store",
			project.Name,
		)
	}
	if err := p.oldStore.ReplaceProject(project); err != nil {
		return resp, errors.Wrapf(
			err,
			"error updating project %q in old store",
			project.Name,
		)
	}

	resp.Project = req.Project

	return resp, nil
}
