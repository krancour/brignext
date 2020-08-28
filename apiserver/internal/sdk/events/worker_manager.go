package events

import (
	"context"
	"fmt"
	"strconv"

	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TODO: Rename this file or find other homes for these functions

func (s *scheduler) createWorkspacePVC(
	ctx context.Context,
	event brignext.Event,
) error {
	storageQuantityStr := event.Worker.Spec.WorkspaceSize
	if storageQuantityStr == "" {
		storageQuantityStr = "1G"
	}
	storageQuantity, err := resource.ParseQuantity(storageQuantityStr)
	if err != nil {
		return errors.Wrapf(
			err,
			"error parsing storage quantity %q for event %q worker",
			storageQuantityStr,
			event.ID,
		)
	}

	workspacePVC := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("workspace-%s", event.ID),
			Namespace: event.Kubernetes.Namespace,
			Labels: map[string]string{
				"brignext.io/component": "workspace",
				"brignext.io/project":   event.ProjectID,
				"brignext.io/event":     event.ID,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &s.config.WorkspaceStorageClass,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteMany,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storageQuantity,
				},
			},
		},
	}

	pvcClient :=
		s.kubeClient.CoreV1().PersistentVolumeClaims(event.Kubernetes.Namespace)
	if _, err := pvcClient.Create(
		ctx,
		&workspacePVC,
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating workspace PVC for event %q worker",
			event.ID,
		)
	}

	return nil
}

func (s *scheduler) createWorkerPod(
	ctx context.Context,
	event brignext.Event,
) error {
	imagePullSecrets := []corev1.LocalObjectReference{}
	if event.Worker.Spec.Kubernetes != nil {
		for _, imagePullSecret := range event.Worker.Spec.Kubernetes.ImagePullSecrets { // nolint: lll
			imagePullSecrets = append(
				imagePullSecrets,
				corev1.LocalObjectReference{
					Name: imagePullSecret,
				},
			)
		}
	}

	// TODO: Decide on the right place to do this stuff. Probably it should be
	// when (near future state), this scheduler uses the API to start the worker.
	if event.Worker.Spec.Container == nil {
		event.Worker.Spec.Container = &brignext.ContainerSpec{}
	}
	image := event.Worker.Spec.Container.Image
	if image == "" {
		image = s.config.DefaultWorkerImage
	}
	imagePullPolicy := event.Worker.Spec.Container.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = s.config.DefaultWorkerImagePullPolicy
	}

	volumes := []corev1.Volume{
		{
			Name: "event",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: fmt.Sprintf("event-%s", event.ID),
				},
			},
		},
	}
	if event.Worker.Spec.UseWorkspace {
		volumes = append(
			volumes,
			corev1.Volume{
				Name: "workspace",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: fmt.Sprintf("workspace-%s", event.ID),
					},
				},
			},
		)
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "event",
			MountPath: "/var/event",
			ReadOnly:  true,
		},
	}
	if event.Worker.Spec.UseWorkspace {
		volumeMounts = append(
			volumeMounts,
			corev1.VolumeMount{
				Name:      "workspace",
				MountPath: "/var/workspace",
				ReadOnly:  true,
			},
		)
	}

	initContainers := []corev1.Container{}
	if event.Worker.Spec.Git != nil && event.Worker.Spec.Git.CloneURL != "" {
		volumes = append(
			volumes,
			corev1.Volume{
				Name: "vcs",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		)

		vcsVolumeMount := corev1.VolumeMount{
			Name:      "vcs",
			MountPath: "/var/vcs",
		}

		volumeMounts = append(volumeMounts, vcsVolumeMount)

		initContainers = append(
			initContainers,
			corev1.Container{
				Name:            "vcs",
				Image:           "brigadecore/git-sidecar:v1.4.0",
				ImagePullPolicy: corev1.PullAlways,
				VolumeMounts: []corev1.VolumeMount{
					vcsVolumeMount,
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
		)
	}

	env := []corev1.EnvVar{}
	for key, val := range event.Worker.Spec.Container.Environment {
		env = append(
			env,
			corev1.EnvVar{
				Name:  key,
				Value: val,
			},
		)
	}

	workerPod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("worker-%s", event.ID),
			Namespace: event.Kubernetes.Namespace,
			Labels: map[string]string{
				"brignext.io/component": "worker",
				"brignext.io/project":   event.ProjectID,
				"brignext.io/event":     event.ID,
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: "workers",
			ImagePullSecrets:   imagePullSecrets,
			RestartPolicy:      corev1.RestartPolicyNever,
			InitContainers:     initContainers,
			Containers: []corev1.Container{
				{
					Name:            "worker",
					Image:           image,
					ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
					Command:         event.Worker.Spec.Container.Command,
					Args:            event.Worker.Spec.Container.Arguments,
					Env:             env,
					VolumeMounts:    volumeMounts,
				},
			},
			Volumes: volumes,
		},
	}

	podClient := s.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace)
	if _, err := podClient.Create(
		ctx,
		&workerPod,
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating pod for event %q worker",
			event.ID,
		)
	}

	return nil
}
