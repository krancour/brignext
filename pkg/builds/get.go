package builds

import (
	context "context"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/pkg/errors"
)

func (b *buildsServer) GetBuilds(
	ctx context.Context,
	req *GetBuildsRequest,
) (*GetBuildsResponse, error) {
	resp := &GetBuildsResponse{}

	var builds []*brigade.Build
	var err error
	if req.ProjectName == "" {
		if builds, err = b.projectStore.GetBuilds(); err != nil {
			return resp, errors.Wrap(err, "error retrieving all builds")
		}
	} else {
		projectID := brigade.ProjectID(req.ProjectName)
		if builds, err = b.projectStore.GetProjectBuilds(projectID); err != nil {
			return resp, errors.Wrapf(
				err,
				"error retrieving builds for project %q",
				req.ProjectName,
			)
		}
	}

	projectNamesByID := map[string]string{}
	resp.Builds = make([]*Build, len(builds))
	for i, build := range builds {
		if _, ok := projectNamesByID[build.ProjectID]; !ok {
			project, err := b.projectStore.GetProject(build.ProjectID)
			if err != nil {
				return resp, errors.Wrapf(
					err,
					"error retrieving project with id %q",
					build.ProjectID,
				)
			}
			if project == nil {
				return resp, errors.Errorf(
					"could not find project with id %q",
					build.ProjectID,
				)
			}
			projectNamesByID[build.ProjectID] = project.Name
		}
		resp.Builds[i] = BrigadeBuildToWireBuild(
			build,
			projectNamesByID[build.ProjectID],
		)
	}

	return resp, nil
}

func (b *buildsServer) GetBuild(
	ctx context.Context,
	req *GetBuildRequest,
) (*GetBuildResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &GetBuildResponse{}

	if build, err := b.projectStore.GetBuild(req.Id); err != nil {
		return nil, errors.Wrapf(err, "error retrieving build %q", req.Id)
	} else if build != nil {
		project, err := b.projectStore.GetProject(build.ProjectID)
		if err != nil {
			return resp, errors.Wrapf(
				err,
				"error retrieving project with id %q",
				build.ProjectID,
			)
		}
		if project == nil {
			return resp, errors.Errorf(
				"could not find project with id %q",
				build.ProjectID,
			)
		}
		resp.Build = BrigadeBuildToWireBuild(build, project.Name)
	}

	return resp, nil
}
