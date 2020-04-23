package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/krancour/brignext"
	"github.com/krancour/brignext/pkg/messaging"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *controller) handleProjectWorkerMessage(
	ctx context.Context,
	msg messaging.Message,
) error {
	workerContext := workerContext{}
	if err := json.Unmarshal(msg.Body(), &workerContext); err != nil {
		return errors.Wrap(
			err,
			"error unmarshaling message body into worker context",
		)
	}

	// Use the API to find the worker and check if anything even needs to be
	// done
	event, err := c.apiClient.GetEvent(ctx, workerContext.EventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q", workerContext.EventID)
	}
	worker, ok := event.Workers[workerContext.WorkerName]
	if !ok {
		// If the event doesn't have a worker by the indicated name, we're done
		log.Printf(
			"WARNING: cannot handle unkown worker %q for event %q",
			workerContext.WorkerName,
			workerContext.EventID,
		)
		return nil
	}
	// If the worker's phase isn't PENDING or RUNNING, there's nothing for us to
	// do. It's already in a terminal state.
	if worker.Status.Phase != brignext.WorkerPhasePending &&
		worker.Status.Phase != brignext.WorkerPhaseRunning {
		return nil
	}

	// If the phase is pending, we'll wait for available capacity and then get
	// the worker pod started
	if worker.Status.Phase == brignext.WorkerPhasePending {
		// Wait for capacity, then start the pod
		select {
		case <-c.availabilityCh:
			if err :=
				c.createWorkspacePVC(event, workerContext.WorkerName); err != nil {
				return errors.Wrapf(
					err,
					"error creating workspace for event %q worker %q",
					workerContext.EventID,
					workerContext.WorkerName,
				)
			}
			if err := c.createWorkerPod(event, workerContext.WorkerName); err != nil {
				return errors.Wrapf(
					err,
					"error starting pod for event %q worker %q",
					workerContext.EventID,
					workerContext.WorkerName,
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
			event, err := c.apiClient.GetEvent(ctx, workerContext.EventID)
			if err != nil {
				return errors.Wrapf(
					err,
					"error retrieving event %q while polling for completion",
					workerContext.EventID,
				)
			}
			worker := event.Workers[workerContext.WorkerName]
			// If worker phase isn't RUNNING, we're done
			if worker.Status.Phase != brignext.WorkerPhaseRunning {
				return nil
			}
		// TODO: We should also have a case for worker timeout
		case <-ctx.Done():
			return nil
		}
	}
}

func (c *controller) createWorkspacePVC(
	event brignext.Event,
	workerName string,
) error {
	worker := event.Workers[workerName]

	storageQuantityStr := worker.WorkspaceSize
	if storageQuantityStr == "" {
		storageQuantityStr = "1G"
	}
	storageQuantity, err := resource.ParseQuantity(storageQuantityStr)
	if err != nil {
		return errors.Wrapf(
			err,
			"error parsing storage quantity %q for event %q worker %q",
			storageQuantityStr,
			event.ID,
			workerName,
		)
	}

	workspacePVC := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      qualifiedWorkerKey(event.ID, workerName),
			Namespace: event.Kubernetes.Namespace,
			Labels: map[string]string{
				"brignext.io/component": "workspace",
				"brignext.io/project":   event.ProjectID,
				"brignext.io/event":     event.ID,
				"brignext.io/worker":    workerName,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: &c.controllerConfig.WorkspaceStorageClass,
			AccessModes:      []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": storageQuantity,
				},
			},
		},
	}

	pvcClient := c.kubeClient.CoreV1().PersistentVolumeClaims(event.Kubernetes.Namespace)
	if _, err := pvcClient.Create(&workspacePVC); err != nil {
		return errors.Wrapf(
			err,
			"error creating workspace PVC for event %q worker %q",
			event.ID,
			workerName,
		)
	}

	return nil
}

func (c *controller) createWorkerPod(
	event brignext.Event,
	workerName string,
) error {
	worker := event.Workers[workerName]

	imagePullSecrets := []corev1.LocalObjectReference{}
	for _, imagePullSecret := range worker.Kubernetes.ImagePullSecrets {
		imagePullSecrets = append(
			imagePullSecrets,
			corev1.LocalObjectReference{
				Name: imagePullSecret,
			},
		)
	}

	image := worker.Container.Image
	if image == "" {
		image = "krancour/brignext-worker:latest" // TODO: Change this
	}
	imagePullPolicy := worker.Container.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = "Always" // TODO: Change this
	}

	volumes := []corev1.Volume{
		corev1.Volume{
			Name: "event",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: event.ID,
				},
			},
		},
		corev1.Volume{
			Name: "worker",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: qualifiedWorkerKey(event.ID, workerName),
				},
			},
		},
		corev1.Volume{
			Name: "workspace",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: qualifiedWorkerKey(event.ID, workerName),
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "event",
			MountPath: "/var/event",
			ReadOnly:  true,
		},
		corev1.VolumeMount{
			Name:      "worker",
			MountPath: "/var/worker",
			ReadOnly:  true,
		},
		corev1.VolumeMount{
			Name:      "workspace",
			MountPath: "/var/workspace",
			ReadOnly:  true,
		},
	}

	initContainers := []corev1.Container{}
	if worker.Git.CloneURL != "" {
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
				Image:           "brigadecore/git-sidecar:latest",
				ImagePullPolicy: corev1.PullAlways,
				VolumeMounts: []corev1.VolumeMount{
					vcsVolumeMount,
				},
				Env: []corev1.EnvVar{
					corev1.EnvVar{
						Name:  "BRIGADE_REMOTE_URL",
						Value: worker.Git.CloneURL,
					},
					corev1.EnvVar{
						Name:  "BRIGADE_COMMIT_ID",
						Value: worker.Git.Commit,
					},
					corev1.EnvVar{
						Name:  "BRIGADE_COMMIT_REF",
						Value: worker.Git.Ref,
					},
					corev1.EnvVar{
						Name: "BRIGADE_REPO_KEY",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: qualifiedWorkerKey(event.ID, workerName),
								},
								Key: "gitSSHKey",
							},
						},
					},
					corev1.EnvVar{
						Name: "BRIGADE_REPO_SSH_CERT",
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: qualifiedWorkerKey(event.ID, workerName),
								},
								Key: "gitSSHCert",
							},
						},
					},
					corev1.EnvVar{
						Name:  "BRIGADE_SUBMODULES",
						Value: strconv.FormatBool(worker.Git.InitSubmodules),
					},
					corev1.EnvVar{
						Name:  "BRIGADE_WORKSPACE",
						Value: "/var/vcs",
					},
				},
			},
		)
	}

	env := []corev1.EnvVar{}
	for key, val := range worker.Container.Environment {
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
			Name:      qualifiedWorkerKey(event.ID, workerName),
			Namespace: event.Kubernetes.Namespace,
			Labels: map[string]string{
				"brignext.io/component": "worker",
				"brignext.io/project":   event.ProjectID,
				"brignext.io/event":     event.ID,
				"brignext.io/worker":    workerName,
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: "workers",
			ImagePullSecrets:   imagePullSecrets,
			RestartPolicy:      corev1.RestartPolicyNever,
			InitContainers:     initContainers,
			Containers: []corev1.Container{
				corev1.Container{
					Name:            strings.ToLower(workerName),
					Image:           image,
					ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
					Command:         strings.Split(worker.Container.Command, ""),
					Env:             env,
					VolumeMounts:    volumeMounts,
				},
			},
			Volumes: volumes,
		},
	}

	podClient := c.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace)
	if _, err := podClient.Create(&workerPod); err != nil {
		return errors.Wrapf(
			err,
			"error creating worker pod for event %q worker %q",
			event.ID,
			workerName,
		)
	}

	return nil
}

func qualifiedWorkerKey(eventID, workerName string) string {
	return fmt.Sprintf("%s-%s", eventID, strings.ToLower(workerName))
}
