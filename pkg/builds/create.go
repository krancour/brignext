package builds

import (
	context "context"

	"github.com/pkg/errors"
)

func (b *buildsServer) CreateBuild(
	ctx context.Context,
	req *CreateBuildRequest,
) (*CreateBuildResponse, error) {
	// TODO: We should do some kind of validation!

	resp := &CreateBuildResponse{}

	brigadeBuild := WireBuildToBrigadeBuild(req.Build)

	if err := b.oldStore.CreateBuild(brigadeBuild); err != nil {
		return resp, errors.Wrapf(
			err,
			"error storing new build in old store %s",
			brigadeBuild.ID,
		)
	}

	// We DON'T write to the new store here. Gateways all still write to the old
	// store only. We'll use a controller to intercept all those old-style build
	// creations and echo them to the new store. Which means we don't have to
	// write to the new store here-- in fact, if we did, we'd end up with a
	// duplicate.

	// This is how we'll be certain to return the build ID that is assigned
	// by the write to the old store.
	resp.Build = BrigadeBuildToWireBuild(brigadeBuild, req.Build.ProjectName)

	return resp, nil
}
