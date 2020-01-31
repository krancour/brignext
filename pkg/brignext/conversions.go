package brignext

import (
	"github.com/brigadecore/brigade/pkg/brigade"
)

func BrigNextProjectToBrigadeProject(
	brignextProject *Project,
) *brigade.Project {
	brigadeProject := &brigade.Project{
		ID:   brigade.ProjectID(brignextProject.Name),
		Name: brignextProject.Name,
		Repo: brigade.Repo{
			Name:     brignextProject.Repo.Name,
			CloneURL: brignextProject.Repo.CloneURL,
			SSHKey:   brignextProject.Repo.SSHKey,
			SSHCert:  brignextProject.Repo.SSHCert,
		},
		DefaultScript:        brignextProject.DefaultScript,
		DefaultScriptName:    brignextProject.DefaultScriptName,
		SharedSecret:         brignextProject.SharedSecret,
		Secrets:              brigade.SecretsMap{},
		InitGitSubmodules:    brignextProject.InitGitSubmodules,
		AllowPrivilegedJobs:  brignextProject.AllowPrivilegedJobs,
		AllowHostMounts:      brignextProject.AllowHostMounts,
		ImagePullSecrets:     brignextProject.ImagePullSecrets,
		WorkerCommand:        brignextProject.WorkerCommand,
		BrigadejsPath:        brignextProject.BrigadejsPath,
		BrigadeConfigPath:    brignextProject.BrigadeConfigPath,
		GenericGatewaySecret: brignextProject.GenericGatewaySecret,
		Kubernetes: brigade.Kubernetes{
			Namespace:         brignextProject.Kubernetes.Namespace,
			VCSSidecar:        brignextProject.Kubernetes.VCSSidecar,
			BuildStorageSize:  brignextProject.Kubernetes.BuildStorageSize,
			BuildStorageClass: brignextProject.Kubernetes.BuildStorageClass,
			CacheStorageClass: brignextProject.Kubernetes.CacheStorageClass,
			AllowSecretKeyRef: brignextProject.Kubernetes.AllowSecretKeyRef,
			ServiceAccount:    brignextProject.Kubernetes.ServiceAccount,
		},
		Github: brigade.Github{
			Token:     brignextProject.Github.Token,
			BaseURL:   brignextProject.Github.BaseURL,
			UploadURL: brignextProject.Github.UploadURL,
		},
		Worker: brigade.WorkerConfig{
			Registry:   brignextProject.Worker.Registry,
			Name:       brignextProject.Worker.Name,
			Tag:        brignextProject.Worker.Tag,
			PullPolicy: brignextProject.Worker.PullPolicy,
		},
	}

	for k, v := range brignextProject.Secrets {
		brigadeProject.Secrets[k] = v
	}

	return brigadeProject
}

func BrigadeProjectToBrigNextProject(
	brigadeProject *brigade.Project,
) *Project {
	brignextProject := &Project{
		Name: brigadeProject.Name,
		Repo: Repo{
			Name:     brigadeProject.Repo.Name,
			CloneURL: brigadeProject.Repo.CloneURL,
			SSHKey:   brigadeProject.Repo.SSHKey,
			SSHCert:  brigadeProject.Repo.SSHCert,
		},
		DefaultScript:        brigadeProject.DefaultScript,
		DefaultScriptName:    brigadeProject.DefaultScriptName,
		SharedSecret:         brigadeProject.SharedSecret,
		Secrets:              SecretsMap{},
		InitGitSubmodules:    brigadeProject.InitGitSubmodules,
		AllowPrivilegedJobs:  brigadeProject.AllowPrivilegedJobs,
		AllowHostMounts:      brigadeProject.AllowHostMounts,
		ImagePullSecrets:     brigadeProject.ImagePullSecrets,
		WorkerCommand:        brigadeProject.WorkerCommand,
		BrigadejsPath:        brigadeProject.BrigadejsPath,
		BrigadeConfigPath:    brigadeProject.BrigadeConfigPath,
		GenericGatewaySecret: brigadeProject.GenericGatewaySecret,
		Kubernetes: Kubernetes{
			Namespace:         brigadeProject.Kubernetes.Namespace,
			VCSSidecar:        brigadeProject.Kubernetes.VCSSidecar,
			BuildStorageSize:  brigadeProject.Kubernetes.BuildStorageSize,
			BuildStorageClass: brigadeProject.Kubernetes.BuildStorageClass,
			CacheStorageClass: brigadeProject.Kubernetes.CacheStorageClass,
			AllowSecretKeyRef: brigadeProject.Kubernetes.AllowSecretKeyRef,
			ServiceAccount:    brigadeProject.Kubernetes.ServiceAccount,
		},
		Github: Github{
			Token:     brigadeProject.Github.Token,
			BaseURL:   brigadeProject.Github.BaseURL,
			UploadURL: brigadeProject.Github.UploadURL,
		},
		Worker: WorkerConfig{
			Registry:   brigadeProject.Worker.Registry,
			Name:       brigadeProject.Worker.Name,
			Tag:        brigadeProject.Worker.Tag,
			PullPolicy: brigadeProject.Worker.PullPolicy,
		},
	}

	for k, v := range brigadeProject.Secrets {
		brignextProject.Secrets[k] = v
	}

	return brignextProject
}

func BrigNextBuildToBrigadeBuild(brignextBuild *Build) *brigade.Build {
	brigadeBuild := &brigade.Build{
		ID:         brignextBuild.ID,
		ProjectID:  brigade.ProjectID(brignextBuild.ProjectName),
		Type:       brignextBuild.Type,
		Provider:   brignextBuild.Provider,
		ShortTitle: brignextBuild.ShortTitle,
		LongTitle:  brignextBuild.LongTitle,
		CloneURL:   brignextBuild.CloneURL,
		Payload:    brignextBuild.Payload,
		Script:     brignextBuild.Script,
		Config:     brignextBuild.Config,
		LogLevel:   brignextBuild.LogLevel,
	}

	if brignextBuild.Revision != nil {
		brigadeBuild.Revision = &brigade.Revision{
			Commit: brignextBuild.Revision.Commit,
			Ref:    brignextBuild.Revision.Ref,
		}
	}

	if brignextBuild.Worker != nil {
		brigadeBuild.Worker = &brigade.Worker{
			ID:        brignextBuild.Worker.ID,
			BuildID:   brignextBuild.Worker.BuildID,
			ProjectID: brignextBuild.Worker.ProjectID,
			StartTime: brignextBuild.Worker.StartTime,
			EndTime:   brignextBuild.Worker.EndTime,
			ExitCode:  brignextBuild.Worker.ExitCode,
			Status:    brigade.JobStatus(brignextBuild.Worker.Status),
		}
	}

	return brigadeBuild
}

func BrigadeBuildToBrigNextBuild(
	brigadeBuild *brigade.Build,
	projectName string,
) *Build {
	brignextBuild := &Build{
		ID:          brigadeBuild.ID,
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
		brignextBuild.Revision = &Revision{
			Commit: brigadeBuild.Revision.Commit,
			Ref:    brigadeBuild.Revision.Ref,
		}
	}

	if brigadeBuild.Worker != nil {
		brignextBuild.Worker = &Worker{
			ID:        brigadeBuild.Worker.ID,
			BuildID:   brigadeBuild.Worker.BuildID,
			ProjectID: brigadeBuild.Worker.ProjectID,
			StartTime: brigadeBuild.Worker.StartTime,
			EndTime:   brigadeBuild.Worker.EndTime,
			ExitCode:  brigadeBuild.Worker.ExitCode,
			Status:    JobStatus(brigadeBuild.Worker.Status),
		}
	}

	return brignextBuild
}
