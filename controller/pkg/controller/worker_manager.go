package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	// If the worker's status isn't PENDING or RUNNING, there's nothing for us to
	// do. It's already in a terminal state.
	if worker.Status != brignext.WorkerStatusPending &&
		worker.Status != brignext.WorkerStatusRunning {
		return nil
	}

	// If the status is pending, we'll wait for available capacity and then get
	// the worker pod started
	if worker.Status == brignext.WorkerStatusPending {
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
			if err := c.apiClient.UpdateEventWorkerStatus(
				ctx,
				workerContext.EventID,
				workerContext.WorkerName,
				brignext.WorkerStatusRunning,
			); err != nil {
				return errors.Wrapf(
					err,
					"error updating status for event %q worker %q",
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
			// If worker status isn't RUNNING, we're done
			if worker.Status != brignext.WorkerStatusRunning {
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

	// TODO: Does the empty string default to the cluster's default storage
	// class? We should probably have a configurable default at the controller
	// level as well.
	storageClassName := worker.Kubernetes.WorkspaceStorageClass

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
			StorageClassName: &storageClassName,
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

// TODO: Finish implementing this
func (c *controller) createWorkerPod(
	event brignext.Event,
	workerName string,
) error {
	worker := event.Workers[workerName]

	image := worker.Container.Image
	if image == "" {
		image = "krancour/brignext-worker:latest" // TODO: Change this
	}
	imagePullPolicy := worker.Container.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = "Always" // TODO: Change this
	}

	secretsClient := c.kubeClient.CoreV1().Secrets(event.Kubernetes.Namespace)
	eventSecrets, err := secretsClient.Get(event.ID, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"error finding event secrets for event %q worker %q",
			event.ID,
			workerName,
		)
	}

	envVars := make([]corev1.EnvVar, len(eventSecrets.Data))
	var i = 0
	for key := range eventSecrets.Data {
		envVars[i] = corev1.EnvVar{
			Name: key,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: event.ID,
					},
					Key: key,
				},
			},
		}
		i++
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
			// TODO: If there's any git configuration, we need an init container
			RestartPolicy: corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				corev1.Container{
					Name:            strings.ToLower(workerName),
					Image:           image,
					ImagePullPolicy: corev1.PullPolicy(imagePullPolicy),
					// TODO: This probably isn't a good enough way, to split up command
					// tokens, but it's what Brigade 1.x does. Good enough for PoC.
					Command: strings.Split(worker.Container.Command, ""),
					Env:     envVars, // These are the secrets
					VolumeMounts: []corev1.VolumeMount{
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
					},
				},
			},
			Volumes: []corev1.Volume{
				corev1.Volume{
					Name: "event",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: event.ID,
							},
						},
					},
				},
				corev1.Volume{
					Name: "worker",
					VolumeSource: corev1.VolumeSource{
						ConfigMap: &corev1.ConfigMapVolumeSource{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: qualifiedWorkerKey(event.ID, workerName),
							},
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
			},
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
