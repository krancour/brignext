package projects

import (
	fmt "fmt"

	"github.com/brigadecore/brigade/pkg/brigade"
	"github.com/krancour/brignext/pkg/brignext"
)

func WireProjectToBrigadeProject(wireProject *Project) *brigade.Project {
	brigadeProject := &brigade.Project{
		ID:                   brigade.ProjectID(wireProject.Name),
		Name:                 wireProject.Name,
		Repo:                 brigade.Repo{},
		DefaultScript:        wireProject.DefaultScript,
		DefaultScriptName:    wireProject.DefaultScriptName,
		SharedSecret:         wireProject.SharedSecret,
		Secrets:              brigade.SecretsMap{},
		InitGitSubmodules:    wireProject.InitGitSubmodules,
		AllowPrivilegedJobs:  wireProject.AllowPrivilegedJobs,
		AllowHostMounts:      wireProject.AllowHostMounts,
		ImagePullSecrets:     wireProject.ImagePullSecrets,
		WorkerCommand:        wireProject.WorkerCommand,
		BrigadejsPath:        wireProject.BrigadejsPath,
		BrigadeConfigPath:    wireProject.BrigadeConfigPath,
		GenericGatewaySecret: wireProject.GenericGatewaySecret,
	}

	if wireProject.Repo != nil {
		brigadeProject.Repo = brigade.Repo{
			Name:     wireProject.Repo.Name,
			CloneURL: wireProject.Repo.CloneURL,
			SSHKey:   wireProject.Repo.SshKey,
			SSHCert:  wireProject.Repo.SshCert,
		}
	}

	if wireProject.Kubernetes != nil {
		brigadeProject.Kubernetes = brigade.Kubernetes{
			Namespace:         wireProject.Kubernetes.Namespace,
			VCSSidecar:        wireProject.Kubernetes.VcsSidecar,
			BuildStorageSize:  wireProject.Kubernetes.BuildStorageSize,
			BuildStorageClass: wireProject.Kubernetes.BuildStorageClass,
			CacheStorageClass: wireProject.Kubernetes.CacheStorageClass,
			AllowSecretKeyRef: wireProject.Kubernetes.AllowSecretKeyRef,
			ServiceAccount:    wireProject.Kubernetes.ServiceAccount,
		}
	}

	if wireProject.Github != nil {
		brigadeProject.Github = brigade.Github{
			Token:     wireProject.Github.Token,
			BaseURL:   wireProject.Github.BaseURL,
			UploadURL: wireProject.Github.UploadURL,
		}
	}

	for k, v := range wireProject.Secrets {
		brigadeProject.Secrets[k] = v
	}

	if wireProject.Worker != nil {
		brigadeProject.Worker = brigade.WorkerConfig{
			Registry:   wireProject.Worker.Registry,
			Name:       wireProject.Worker.Name,
			Tag:        wireProject.Worker.Tag,
			PullPolicy: wireProject.Worker.PullPolicy,
		}
	}

	return brigadeProject
}

func BrigadeProjectToWireProject(brigadeProject *brigade.Project) *Project {
	wireProject := &Project{
		Name: brigadeProject.Name,
		Repo: &Repo{
			Name:     brigadeProject.Repo.Name,
			CloneURL: brigadeProject.Repo.CloneURL,
			SshKey:   brigadeProject.Repo.SSHKey,
			SshCert:  brigadeProject.Repo.SSHCert,
		},
		DefaultScript:     brigadeProject.DefaultScript,
		DefaultScriptName: brigadeProject.DefaultScriptName,
		Kubernetes: &Kubernetes{
			Namespace:         brigadeProject.Kubernetes.Namespace,
			VcsSidecar:        brigadeProject.Kubernetes.VCSSidecar,
			BuildStorageSize:  brigadeProject.Kubernetes.BuildStorageSize,
			BuildStorageClass: brigadeProject.Kubernetes.BuildStorageClass,
			CacheStorageClass: brigadeProject.Kubernetes.CacheStorageClass,
			AllowSecretKeyRef: brigadeProject.Kubernetes.AllowSecretKeyRef,
			ServiceAccount:    brigadeProject.Kubernetes.ServiceAccount,
		},
		SharedSecret: brigadeProject.SharedSecret,
		Github: &Github{
			Token:     brigadeProject.Github.Token,
			BaseURL:   brigadeProject.Github.BaseURL,
			UploadURL: brigadeProject.Github.UploadURL,
		},
		Secrets: map[string]string{},
		Worker: &WorkerConfig{
			Name:       brigadeProject.Worker.Name,
			Tag:        brigadeProject.Worker.Tag,
			PullPolicy: brigadeProject.Worker.PullPolicy,
		},
		InitGitSubmodules:    brigadeProject.InitGitSubmodules,
		AllowPrivilegedJobs:  brigadeProject.AllowPrivilegedJobs,
		AllowHostMounts:      brigadeProject.AllowHostMounts,
		ImagePullSecrets:     brigadeProject.ImagePullSecrets,
		WorkerCommand:        brigadeProject.WorkerCommand,
		BrigadejsPath:        brigadeProject.BrigadejsPath,
		BrigadeConfigPath:    brigadeProject.BrigadeConfigPath,
		GenericGatewaySecret: brigadeProject.GenericGatewaySecret,
	}

	for k, v := range brigadeProject.Secrets {
		wireProject.Secrets[k] = fmt.Sprintf("%v", v)
	}

	return wireProject
}

func WireProjectToBrignextProject(wireProject *Project) *brignext.Project {
	brignextProject := &brignext.Project{
		Name:                 wireProject.Name,
		Repo:                 brignext.Repo{},
		DefaultScript:        wireProject.DefaultScript,
		DefaultScriptName:    wireProject.DefaultScriptName,
		SharedSecret:         wireProject.SharedSecret,
		Secrets:              brignext.SecretsMap{},
		InitGitSubmodules:    wireProject.InitGitSubmodules,
		AllowPrivilegedJobs:  wireProject.AllowPrivilegedJobs,
		AllowHostMounts:      wireProject.AllowHostMounts,
		ImagePullSecrets:     wireProject.ImagePullSecrets,
		WorkerCommand:        wireProject.WorkerCommand,
		BrigadejsPath:        wireProject.BrigadejsPath,
		BrigadeConfigPath:    wireProject.BrigadeConfigPath,
		GenericGatewaySecret: wireProject.GenericGatewaySecret,
	}

	if wireProject.Repo != nil {
		brignextProject.Repo = brignext.Repo{
			Name:     wireProject.Repo.Name,
			CloneURL: wireProject.Repo.CloneURL,
			SSHKey:   wireProject.Repo.SshKey,
			SSHCert:  wireProject.Repo.SshCert,
		}
	}

	if wireProject.Kubernetes != nil {
		brignextProject.Kubernetes = brignext.Kubernetes{
			Namespace:         wireProject.Kubernetes.Namespace,
			VCSSidecar:        wireProject.Kubernetes.VcsSidecar,
			BuildStorageSize:  wireProject.Kubernetes.BuildStorageSize,
			BuildStorageClass: wireProject.Kubernetes.BuildStorageClass,
			CacheStorageClass: wireProject.Kubernetes.CacheStorageClass,
			AllowSecretKeyRef: wireProject.Kubernetes.AllowSecretKeyRef,
			ServiceAccount:    wireProject.Kubernetes.ServiceAccount,
		}
	}

	if wireProject.Github != nil {
		brignextProject.Github = brignext.Github{
			Token:     wireProject.Github.Token,
			BaseURL:   wireProject.Github.BaseURL,
			UploadURL: wireProject.Github.UploadURL,
		}
	}

	for k, v := range wireProject.Secrets {
		brignextProject.Secrets[k] = v
	}

	if wireProject.Worker != nil {
		brignextProject.Worker = brignext.WorkerConfig{
			Registry:   wireProject.Worker.Registry,
			Name:       wireProject.Worker.Name,
			Tag:        wireProject.Worker.Tag,
			PullPolicy: wireProject.Worker.PullPolicy,
		}
	}

	return brignextProject
}

func BrignextProjectToWireProject(brignextProject *brignext.Project) *Project {
	wireProject := &Project{
		Name: brignextProject.Name,
		Repo: &Repo{
			Name:     brignextProject.Repo.Name,
			CloneURL: brignextProject.Repo.CloneURL,
			SshKey:   brignextProject.Repo.SSHKey,
			SshCert:  brignextProject.Repo.SSHCert,
		},
		DefaultScript:     brignextProject.DefaultScript,
		DefaultScriptName: brignextProject.DefaultScriptName,
		Kubernetes: &Kubernetes{
			Namespace:         brignextProject.Kubernetes.Namespace,
			VcsSidecar:        brignextProject.Kubernetes.VCSSidecar,
			BuildStorageSize:  brignextProject.Kubernetes.BuildStorageSize,
			BuildStorageClass: brignextProject.Kubernetes.BuildStorageClass,
			CacheStorageClass: brignextProject.Kubernetes.CacheStorageClass,
			AllowSecretKeyRef: brignextProject.Kubernetes.AllowSecretKeyRef,
			ServiceAccount:    brignextProject.Kubernetes.ServiceAccount,
		},
		SharedSecret: brignextProject.SharedSecret,
		Github: &Github{
			Token:     brignextProject.Github.Token,
			BaseURL:   brignextProject.Github.BaseURL,
			UploadURL: brignextProject.Github.UploadURL,
		},
		Secrets: map[string]string{},
		Worker: &WorkerConfig{
			Registry:   brignextProject.Worker.Registry,
			Name:       brignextProject.Worker.Name,
			Tag:        brignextProject.Worker.Tag,
			PullPolicy: brignextProject.Worker.PullPolicy,
		},
		InitGitSubmodules:    brignextProject.InitGitSubmodules,
		AllowPrivilegedJobs:  brignextProject.AllowPrivilegedJobs,
		AllowHostMounts:      brignextProject.AllowHostMounts,
		ImagePullSecrets:     brignextProject.ImagePullSecrets,
		WorkerCommand:        brignextProject.WorkerCommand,
		BrigadejsPath:        brignextProject.BrigadejsPath,
		BrigadeConfigPath:    brignextProject.BrigadeConfigPath,
		GenericGatewaySecret: brignextProject.GenericGatewaySecret,
	}

	for k, v := range brignextProject.Secrets {
		wireProject.Secrets[k] = fmt.Sprintf("%v", v)
	}

	return wireProject
}
