package projects

import (
	context "context"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/pkg/errors"
)

func (p *projectsServer) GetProjects(
	ctx context.Context,
	req *GetProjectsRequest,
) (*GetProjectsResponse, error) {
	resp := &GetProjectsResponse{}

	projects, err := p.projectStore.GetProjects()
	if err != nil {
		return resp, errors.Wrap(err, "error retrieving all projects")
	}

	resp.Projects = make([]*Project, len(projects))
	for i, project := range projects {
		resp.Projects[i] = BrigadeProjectToWireProject(project)
	}

	return resp, nil
}

func (p *projectsServer) GetProject(
	ctx context.Context,
	req *GetProjectRequest,
) (*GetProjectResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &GetProjectResponse{}

	projectID := brigade.ProjectID(req.ProjectName)
	project, err := p.projectStore.GetProject(projectID)
	if err != nil {
		return resp, errors.Wrapf(
			err,
			"error retrieving project %q",
			req.ProjectName,
		)
	} else if project != nil {
		resp.Project = BrigadeProjectToWireProject(project)
	}

	return resp, nil
}
