package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/common/messaging"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *controller) handleProjectWorkerMessage(
	ctx context.Context,
	msg messaging.Message,
) error {
	workerCtx := workerContext{}
	if err := json.Unmarshal(msg.Body(), &workerCtx); err != nil {
		return errors.Wrap(
			err,
			"error unmarshaling message body into worker context",
		)
	}

	// Use the API to find the worker and check if anything even needs to be
	// done
	event, err := c.apiClient.Events().Get(ctx, workerCtx.EventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q", workerCtx.EventID)
	}
	// If the worker's phase isn't PENDING or RUNNING, there's nothing for us to
	// do. It's already in a terminal state.
	if event.Status.WorkerStatus.Phase != brignext.WorkerPhasePending &&
		event.Status.WorkerStatus.Phase != brignext.WorkerPhaseRunning {
		return nil
	}

	// If the phase is pending, we'll wait for available capacity and then get
	// the worker pod started
	if event.Status.WorkerStatus.Phase == brignext.WorkerPhasePending {
		// Wait for capacity, then start the pod
		select {
		case <-c.availabilityCh:
			if err :=
				c.createWorkspacePVC(ctx, event); err != nil {
				return errors.Wrapf(
					err,
					"error creating workspace for event %q worker",
					workerCtx.EventID,
				)
			}
			if err := c.createWorkerPod(
				ctx,
				event,
			); err != nil {
				return errors.Wrapf(
					err,
					"error starting pod for event %q worker",
					workerCtx.EventID,
				)
			}
		case <-ctx.Done():
			return nil
		}
	}

	// Wait for the worker to reach a terminal state. We do this by periodically
	// polling the API because that's much simpler than trying to coordinate with
	// the routine that's continuously monitoring worker pods. We can revisit this
	// later if need be.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			event, err := c.apiClient.Events().Get(ctx, workerCtx.EventID)
			if err != nil {
				return errors.Wrapf(
					err,
					"error retrieving event %q while polling for completion",
					workerCtx.EventID,
				)
			}
			// If worker phase isn't RUNNING, we're done
			if event.Status.WorkerStatus.Phase != brignext.WorkerPhaseRunning {
				return nil
			}
		// TODO: We should also have a case for worker timeout
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *controller) createWorkspacePVC(
	ctx context.Context,
	event brignext.Event,
) error {
	storageQuantityStr := event.Worker.WorkspaceSize
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
			StorageClassName: &c.controllerConfig.WorkspaceStorageClass,
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
		c.kubeClient.CoreV1().PersistentVolumeClaims(event.Kubernetes.Namespace)
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

func (c *controller) createWorkerPod(
	ctx context.Context,
	event brignext.Event,
) error {
	imagePullSecrets := []corev1.LocalObjectReference{}
	for _, imagePullSecret := range event.Worker.Kubernetes.ImagePullSecrets {
		imagePullSecrets = append(
			imagePullSecrets,
			corev1.LocalObjectReference{
				Name: imagePullSecret,
			},
		)
	}

	image := event.Worker.Container.Image
	if image == "" {
		image = c.controllerConfig.DefaultWorkerImage
	}
	imagePullPolicy := event.Worker.Container.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = c.controllerConfig.DefaultWorkerImagePullPolicy
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
		{
			Name: "workspace",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: fmt.Sprintf("workspace-%s", event.ID),
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "event",
			MountPath: "/var/event",
			ReadOnly:  true,
		},
		{
			Name:      "workspace",
			MountPath: "/var/workspace",
			ReadOnly:  true,
		},
	}

	initContainers := []corev1.Container{}
	if event.Worker.Git.CloneURL != "" {
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
						Value: event.Worker.Git.CloneURL,
					},
					{
						Name:  "BRIGADE_COMMIT_ID",
						Value: event.Worker.Git.Commit,
					},
					{
						Name:  "BRIGADE_COMMIT_REF",
						Value: event.Worker.Git.Ref,
					},
					{
						Name: "BRIGADE_REPO_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: fmt.Sprintf("worker-%s", event.ID),
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
									Name: fmt.Sprintf("worker-%s", event.ID),
								},
								Key: "gitSSHCert",
							},
						},
					},
					{
						Name:  "BRIGADE_SUBMODULES",
						Value: strconv.FormatBool(event.Worker.Git.InitSubmodules),
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
	for key, val := range event.Worker.Container.Environment {
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
					Command:         strings.Split(event.Worker.Container.Command, ""),
					Env:             env,
					VolumeMounts:    volumeMounts,
				},
			},
			Volumes: volumes,
		},
	}

	podClient := c.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace)
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
