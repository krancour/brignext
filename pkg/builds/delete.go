package builds

import (
	context "context"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/brigadecore/brigade/pkg/storage"
	oldStorage "github.com/brigadecore/brigade/pkg/storage"
	"github.com/pkg/errors"
)

func (b *buildsServer) DeleteBuild(
	ctx context.Context,
	req *DeleteBuildRequest,
) (*DeleteBuildResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &DeleteBuildResponse{}

	if err := b.projectStore.DeleteBuild(
		req.Id,
		storage.DeleteBuildOptions{
			SkipRunningBuilds: !req.Force,
		},
	); err != nil {
		return resp, errors.Wrapf(
			err,
			"error deleting build %q from new store",
			req.Id,
		)
	}

	if err := b.oldStore.DeleteBuild(
		req.Id,
		oldStorage.DeleteBuildOptions{
			SkipRunningBuilds: !req.Force,
		},
	); err != nil {
		return resp, errors.Wrapf(
			err,
			"error deleting build %q from old store",
			req.Id,
		)
	}

	return resp, nil
}

func (b *buildsServer) DeleteAllBuilds(
	ctx context.Context,
	req *DeleteAllBuildsRequest,
) (*DeleteAllBuildsResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &DeleteAllBuildsResponse{}

	projectID := brigade.ProjectID(req.ProjectName)
	builds, err := b.projectStore.GetProjectBuilds(projectID)
	if err != nil {
		return resp, errors.Wrapf(
			err,
			"error retrieving builds for project %q",
			req.ProjectName,
		)
	}

	for _, build := range builds {
		if err := b.projectStore.DeleteBuild(
			build.ID,
			storage.DeleteBuildOptions{
				SkipRunningBuilds: !req.Force,
			},
		); err != nil {
			return resp, errors.Wrapf(
				err,
				"error deleting build %q from new store",
				build.ID,
			)
		}
		if err := b.oldStore.DeleteBuild(
			build.ID,
			oldStorage.DeleteBuildOptions{
				SkipRunningBuilds: !req.Force,
			},
		); err != nil {
			return resp, errors.Wrapf(
				err,
				"error deleting build %q from old store",
				build.ID,
			)
		}
	}

	// TODO: Cascade delete to associated jobs

	return resp, nil
}
