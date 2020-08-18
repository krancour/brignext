package events

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Rename this file or find other homes for these functions

func (s *scheduler) createJobSecret(
	ctx context.Context,
	event brignext.Event,
	jobName string,
	jobSpec brignext.JobSpec,
) error {

	jobSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("job-%s-%s", event.ID, strings.ToLower(jobName)),
			Namespace: event.Kubernetes.Namespace,
			Labels: map[string]string{
				"brignext.io/component": "job",
				"brignext.io/project":   event.ProjectID,
				"brignext.io/event":     event.ID,
				"brignext.io/job":       jobName,
			},
		},
		Type:       "brignext.io/job",
		StringData: map[string]string{},
	}

	for k, v := range jobSpec.PrimaryContainer.Environment {
		jobSecret.StringData[fmt.Sprintf("%s.%s", jobName, k)] = v
	}
	for sidecarName, sidecareSpec := range jobSpec.SidecarContainers {
		for k, v := range sidecareSpec.Environment {
			jobSecret.StringData[fmt.Sprintf("%s.%s", sidecarName, k)] = v
		}
	}

	secretsClient := s.kubeClient.CoreV1().Secrets(event.Kubernetes.Namespace)
	if _, err := secretsClient.Create(
		ctx,
		&jobSecret,
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating secret for event %q job %q",
			event.ID,
			jobName,
		)
	}

	return nil
}

func (s *scheduler) createJobPod(
	ctx context.Context,
	event brignext.Event,
	jobName string,
	jobSpec brignext.JobSpec,
) error {
	// Determine if ANY of the job's containers:
	//   1. Use shared workspace
	//   2. Use source code from git
	//   3. Mount the host's Docker socket
	var useWorkspace = jobSpec.PrimaryContainer.UseWorkspace
	var useSource = jobSpec.PrimaryContainer.UseSource
	var useDockerSocket = jobSpec.PrimaryContainer.UseHostDockerSocket
	for _, sidecarContainer := range jobSpec.SidecarContainers {
		if sidecarContainer.UseWorkspace {
			useWorkspace = true
		}
		if sidecarContainer.UseSource {
			useSource = true
		}
		if sidecarContainer.UseHostDockerSocket {
			useDockerSocket = true
		}
	}

	imagePullSecrets := []corev1.LocalObjectReference{}
	if event.Worker.Spec.Kubernetes != nil {
		imagePullSecrets = make(
			[]corev1.LocalObjectReference,
			len(event.Worker.Spec.Kubernetes.ImagePullSecrets),
		)
		for i, imagePullSecret := range event.Worker.Spec.Kubernetes.ImagePullSecrets { // nolint: lll
			imagePullSecrets[i] = corev1.LocalObjectReference{
				Name: imagePullSecret,
			}
		}
	}

	volumes := []corev1.Volume{}
	if useWorkspace {
		volumes = []corev1.Volume{
			{
				Name: "workspace",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("workspace-%s", event.ID),
					},
				},
			},
		}
	}
	if useSource &&
		event.Worker.Spec.Git != nil &&
		event.Worker.Spec.Git.CloneURL != "" {
		volumes = append(
			volumes,
			corev1.Volume{
				Name: "vcs",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		)
	}
	if useDockerSocket {
		volumes = append(
			volumes,
			corev1.Volume{
				Name: "docker-socket",
				VolumeSource: corev1.VolumeSource{
					HostPath: &corev1.HostPathVolumeSource{
						Path: "/var/run/docker.sock",
					},
				},
			},
		)
	}

	initContainers := []corev1.Container{}
	if useSource &&
		event.Worker.Spec.Git != nil &&
		event.Worker.Spec.Git.CloneURL != "" {
		initContainers = []corev1.Container{
			{
				Name:            "vcs",
				Image:           "brigadecore/git-sidecar:v1.4.0",
				ImagePullPolicy: corev1.PullAlways,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "vcs",
						MountPath: "/var/vcs",
					},
				},
				Env: []corev1.EnvVar{
					{
						Name:  "BRIGADE_REMOTE_URL",
						Value: event.Worker.Spec.Git.CloneURL,
					},
					{
						Name:  "BRIGADE_COMMIT_ID",
						Value: event.Worker.Spec.Git.Commit,
					},
					{
						Name:  "BRIGADE_COMMIT_REF",
						Value: event.Worker.Spec.Git.Ref,
					},
					{
						Name: "BRIGADE_REPO_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: fmt.Sprintf("event-%s", event.ID),
								},
								Key: "gitSSHKey",
							},
						},
					},
					{
						Name: "BRIGADE_REPO_SSH_CERT",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: fmt.Sprintf("event-%s", event.ID),
								},
								Key: "gitSSHCert",
							},
						},
					},
					{
						Name:  "BRIGADE_SUBMODULES",
						Value: strconv.FormatBool(event.Worker.Spec.Git.InitSubmodules),
					},
					{
						Name:  "BRIGADE_WORKSPACE",
						Value: "/var/vcs",
					},
				},
			},
		}

		// This slice is big enough to hold the primary container AND all (if any)
		// sidecar containers.
		containers := make([]corev1.Container, len(jobSpec.SidecarContainers)+1)

		// The primary container will be the 0 container in this list.
		// nolint: lll
		containers[0] = corev1.Container{
			Name:            jobName, // Primary container takes the job's name
			ImagePullPolicy: corev1.PullPolicy(jobSpec.PrimaryContainer.ImagePullPolicy),
			Command:         jobSpec.PrimaryContainer.Command,
			Args:            jobSpec.PrimaryContainer.Arguments,
			Env:             make([]corev1.EnvVar, len(jobSpec.PrimaryContainer.Environment)),
			VolumeMounts:    []corev1.VolumeMount{},
		}
		i := 0
		for key := range jobSpec.PrimaryContainer.Environment {
			containers[0].Env[i] = corev1.EnvVar{
				Name: key,
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: fmt.Sprintf(
								"job-%s-%s",
								event.ID,
								strings.ToLower(jobName),
							),
						},
						Key: fmt.Sprintf("%s.%s", jobName, key),
					},
				},
			}
			i++
		}
		if jobSpec.PrimaryContainer.UseWorkspace {
			containers[0].VolumeMounts = []corev1.VolumeMount{
				{
					Name:      "workspace",
					MountPath: jobSpec.PrimaryContainer.WorkspaceMountPath,
				},
			}
		}
		if jobSpec.PrimaryContainer.UseSource {
			containers[0].VolumeMounts = append(
				containers[0].VolumeMounts,
				corev1.VolumeMount{
					Name:      "vcs",
					MountPath: jobSpec.PrimaryContainer.SourceMountPath,
				},
			)
		}
		if jobSpec.PrimaryContainer.UseHostDockerSocket {
			containers[0].VolumeMounts = append(
				containers[0].VolumeMounts,
				corev1.VolumeMount{
					Name:      "docker-socket",
					MountPath: "/var/run/docker.sock",
				},
			)
		}
		if jobSpec.PrimaryContainer.Privileged {
			tru := true
			containers[0].SecurityContext = &corev1.SecurityContext{
				Privileged: &tru,
			}
		}

		// Now add all the sidecars...
		i = 1
		for sidecarName, sidecarSpec := range jobSpec.SidecarContainers {
			containers[i] = corev1.Container{
				Name:            sidecarName,
				ImagePullPolicy: corev1.PullPolicy(sidecarSpec.ImagePullPolicy),
				Command:         sidecarSpec.Command,
				Args:            sidecarSpec.Arguments,
				Env:             make([]corev1.EnvVar, len(sidecarSpec.Environment)),
			}
			j := 0
			for key := range sidecarSpec.Environment {
				containers[i].Env[j] = corev1.EnvVar{
					Name: key,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: fmt.Sprintf(
									"job-%s-%s",
									event.ID,
									strings.ToLower(jobName),
								),
							},
							Key: fmt.Sprintf("%s.%s", sidecarName, key),
						},
					},
				}
				j++
			}
			if sidecarSpec.UseWorkspace {
				containers[i].VolumeMounts = []corev1.VolumeMount{
					{
						Name:      "workspace",
						MountPath: sidecarSpec.WorkspaceMountPath,
					},
				}
			}
			if sidecarSpec.UseSource {
				containers[i].VolumeMounts = append(
					containers[i].VolumeMounts,
					corev1.VolumeMount{
						Name:      "vcs",
						MountPath: sidecarSpec.SourceMountPath,
					},
				)
			}
			if sidecarSpec.UseHostDockerSocket {
				containers[i].VolumeMounts = append(
					containers[i].VolumeMounts,
					corev1.VolumeMount{
						Name:      "docker-socket",
						MountPath: "/var/run/docker.sock",
					},
				)
			}
			if sidecarSpec.Privileged {
				tru := true
				containers[i].SecurityContext = &corev1.SecurityContext{
					Privileged: &tru,
				}
			}
			i++
		}

		jobPod := corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("job-%s-%s", event.ID, strings.ToLower(jobName)),
				Namespace: event.Kubernetes.Namespace,
				Labels: map[string]string{
					"brignext.io/component": "job",
					"brignext.io/project":   event.ProjectID,
					"brignext.io/event":     event.ID,
					"brignext.io/job":       jobName,
				},
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: "jobs",
				ImagePullSecrets:   imagePullSecrets,
				RestartPolicy:      corev1.RestartPolicyNever,
				InitContainers:     initContainers,
				Containers:         containers,
				Volumes:            volumes,
			},
		}

		podClient := s.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace)
		if _, err := podClient.Create(
			ctx,
			&jobPod,
			metav1.CreateOptions{},
		); err != nil {
			return errors.Wrapf(
				err,
				"error creating pod for event %q job %q",
				event.ID,
				jobName,
			)
		}

	}

	return nil
}
