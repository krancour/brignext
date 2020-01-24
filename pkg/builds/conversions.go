package builds

import (
	"time"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/krancour/brignext/pkg/brignext"
)

func WireBuildToBrigadeBuild(wireBuild *Build) *brigade.Build {
	brigadeBuild := &brigade.Build{
		ID:         wireBuild.Id,
		ProjectID:  brigade.ProjectID(wireBuild.ProjectName),
		Type:       wireBuild.Type,
		Provider:   wireBuild.Provider,
		ShortTitle: wireBuild.ShortTitle,
		LongTitle:  wireBuild.LongTitle,
		CloneURL:   wireBuild.CloneURL,
		Payload:    wireBuild.Payload,
		Script:     wireBuild.Script,
		Config:     wireBuild.Config,
		LogLevel:   wireBuild.LogLevel,
	}

	if wireBuild.Revision != nil {
		brigadeBuild.Revision = &brigade.Revision{
			Commit: wireBuild.Revision.Commit,
			Ref:    wireBuild.Revision.Ref,
		}
	}

	if wireBuild.Worker != nil {
		brigadeBuild.Worker = &brigade.Worker{
			ID:        wireBuild.Worker.Id,
			BuildID:   wireBuild.Worker.BuildID,
			ProjectID: wireBuild.Worker.ProjectID,
			StartTime: time.Unix(0, wireBuild.Worker.StartTime),
			EndTime:   time.Unix(0, wireBuild.Worker.EndTime),
			ExitCode:  wireBuild.Worker.ExitCode,
			Status:    brigade.JobStatus(wireBuild.Worker.Status),
		}
	}

	return brigadeBuild
}

func BrigadeBuildToWireBuild(
	brigadeBuild *brigade.Build,
	projectName string,
) *Build {
	wireBuild := &Build{
		Id:          brigadeBuild.ID,
		ProjectName: projectName,
		Type:        brigadeBuild.Type,
		Provider:    brigadeBuild.Provider,
		ShortTitle:  brigadeBuild.ShortTitle,
		LongTitle:   brigadeBuild.LongTitle,
		CloneURL:    brigadeBuild.CloneURL,
		Payload:     brigadeBuild.Payload,
		Script:      brigadeBuild.Script,
		Config:      brigadeBuild.Config,
		LogLevel:    brigadeBuild.LogLevel,
	}

	if brigadeBuild.Revision != nil {
		wireBuild.Revision = &Revision{
			Commit: brigadeBuild.Revision.Commit,
			Ref:    brigadeBuild.Revision.Ref,
		}
	}

	if brigadeBuild.Worker != nil {
		wireBuild.Worker = &Worker{
			Id:        brigadeBuild.Worker.ID,
			BuildID:   brigadeBuild.Worker.BuildID,
			ProjectID: brigadeBuild.Worker.ProjectID,
			StartTime: brigadeBuild.Worker.StartTime.UnixNano(),
			EndTime:   brigadeBuild.Worker.EndTime.UnixNano(),
			ExitCode:  brigadeBuild.Worker.ExitCode,
			Status:    string(brigadeBuild.Worker.Status),
		}
	}

	return wireBuild
}

func WireBuildToBrignextBuild(wireBuild *Build) *brignext.Build {
	brignextBuild := &brignext.Build{
		ID:          wireBuild.Id,
		ProjectName: wireBuild.ProjectName,
		Type:        wireBuild.Type,
		Provider:    wireBuild.Provider,
		ShortTitle:  wireBuild.ShortTitle,
		LongTitle:   wireBuild.LongTitle,
		CloneURL:    wireBuild.CloneURL,
		Payload:     wireBuild.Payload,
		Script:      wireBuild.Script,
		Config:      wireBuild.Config,
		LogLevel:    wireBuild.LogLevel,
	}

	if wireBuild.Revision != nil {
		brignextBuild.Revision = &brignext.Revision{
			Commit: wireBuild.Revision.Commit,
			Ref:    wireBuild.Revision.Ref,
		}
	}

	if wireBuild.Worker != nil {
		brignextBuild.Worker = &brignext.Worker{
			ID:        wireBuild.Worker.Id,
			BuildID:   wireBuild.Worker.BuildID,
			ProjectID: wireBuild.Worker.ProjectID,
			StartTime: time.Unix(0, wireBuild.Worker.StartTime),
			EndTime:   time.Unix(0, wireBuild.Worker.EndTime),
			ExitCode:  wireBuild.Worker.ExitCode,
			Status:    brignext.JobStatus(wireBuild.Worker.Status),
		}
	}

	return brignextBuild
}

func BrignextBuildToWireBuild(brignextBuild *brignext.Build) *Build {
	wireBuild := &Build{
		Id:          brignextBuild.ID,
		ProjectName: brignextBuild.ProjectName,
		Type:        brignextBuild.Type,
		Provider:    brignextBuild.Provider,
		ShortTitle:  brignextBuild.ShortTitle,
		LongTitle:   brignextBuild.LongTitle,
		CloneURL:    brignextBuild.CloneURL,
		Payload:     brignextBuild.Payload,
		Script:      brignextBuild.Script,
		Config:      brignextBuild.Config,
		LogLevel:    brignextBuild.LogLevel,
	}

	if brignextBuild.Revision != nil {
		wireBuild.Revision = &Revision{
			Commit: brignextBuild.Revision.Commit,
			Ref:    brignextBuild.Revision.Ref,
		}
	}

	if brignextBuild.Worker != nil {
		wireBuild.Worker = &Worker{
			Id:        brignextBuild.Worker.ID,
			BuildID:   brignextBuild.Worker.BuildID,
			ProjectID: brignextBuild.Worker.ProjectID,
			StartTime: brignextBuild.Worker.StartTime.UnixNano(),
			EndTime:   brignextBuild.Worker.EndTime.UnixNano(),
			ExitCode:  brignextBuild.Worker.ExitCode,
			Status:    string(brignextBuild.Worker.Status),
		}
	}

	return wireBuild
}
