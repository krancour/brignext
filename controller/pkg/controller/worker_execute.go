package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/krancour/brignext"
	"github.com/krancour/brignext/pkg/messaging"
	"github.com/pkg/errors"
)

// workerExecute launches a pod corresponding to the specified worker then
// schedules a follow-up task to monitor that pod for completion.
func (c *controller) workerExecute(
	ctx context.Context,
	message messaging.Message,
) error {
	messageBodyStruct := struct {
		Event  string `json:"event"`
		Worker string `json:"worker"`
	}{}
	if err := json.Unmarshal(message.Body(), &messageBodyStruct); err != nil {
		// TODO: How can we make this error a little more descriptive?
		return errors.Errorf("error decoding message %q", message.ID())
	}

	eventID := messageBodyStruct.Event
	workerName := messageBodyStruct.Worker

	log.Printf(
		"INFO: received executeWorker task for worker %q of event %q",
		workerName,
		eventID,
	)

	// Find the event
	event, err := c.apiClient.GetEvent(ctx, eventID)
	if err != nil {
		if _, ok := err.(*brignext.ErrEventNotFound); ok {
			// The event wasn't found. The likely scenario is that it was deleted.
			// We're not going to treat this as an error. We're just going to move on.
			return nil
		}
		return errors.Wrapf(
			err,
			"error retrieving event %q for worker %q execution",
			eventID,
			workerName,
		)
	}

	worker, ok := event.Workers[workerName]
	if !ok {
		return errors.Errorf(
			"executeWorker task %q failed because event %q did not have a worker "+
				"named %q",
			message.ID(),
			eventID,
			workerName,
		)
	}

	// There's an unlikely, but non-zero possibility that this handler runs with
	// the worker status already, unexpectedly in a RUNNING state. This could only
	// happen if the handler has already run for this event at least once before
	// and succeeded in updating the worker's status in the database, but the
	// controller process exited unexpectedly before the worker completed.
	//
	// So...
	//
	// If the status is already RUNNING, don't do any updates to the database.
	// Just start waiting for the worker to complete.
	if worker.Status == brignext.WorkerStatusRunning {
		// TODO: Wait for the worker to complete
		select {}
	}

	// If the event status is anything other than PENDING, just log it and move
	// on.
	if worker.Status != brignext.WorkerStatusPending {
		log.Printf(
			"WARNING: worker %q of event %q status was unexpectedly %q when "+
				"initiating worker execution. Taking no action and moving on.",
			workerName,
			event.ID,
			worker.Status,
		)
		return nil
	}

	// Get the worker pod up and running if it isn't already
	if err := c.createWorkerPod(ctx, event, workerName); err != nil {
		return errors.Wrapf(
			err,
			"error ensuring worker pod running for worker %q of event %q",
			workerName,
			event.ID,
		)
	}

	// Update the worker status in the database
	if err := c.apiClient.UpdateEventWorkerStatus(
		ctx,
		eventID,
		workerName,
		brignext.WorkerStatusRunning,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status on worker %q of event %q",
			workerName,
			event.ID,
		)
	}

	// TODO: Wait for the worker to complete
	select {}
}

// TODO: Implement this
// krancour: Finishing the implementation here is the magic that will expose
// whatever isn't right with the new domain model.
func (c *controller) createWorkerPod(
	ctx context.Context,
	event brignext.Event,
	workerName string,
) error {
	podName := getWorkerPodName(event.ID, workerName)
	log.Printf(
		"this is where worker pod %q in namespace %q would have been created",
		podName,
		event.Kubernetes.Namespace,
	)

	// podsClient := c.kubeClient.CoreV1().Pods(event.Namespace)

	// // ---------------------------------------------------------------------------

	// env := workerEnv(project, build, config)

	// cmd := []string{"yarn", "-s", "start"}
	// if config.WorkerCommand != "" {
	// 	cmd = strings.Split(config.WorkerCommand, " ")
	// }
	// if cmdBytes, ok := project.Data["workerCommand"]; ok && len(cmdBytes) > 0 {
	// 	cmd = strings.Split(string(cmdBytes), " ")
	// }

	// image, pullPolicy := workerImageConfig(project, config)

	// volumeMounts := []v1.VolumeMount{}
	// buildVolumeMount := v1.VolumeMount{
	// 	Name:      "brigade-build",
	// 	MountPath: "/etc/brigade",
	// 	ReadOnly:  true,
	// }
	// projectVolumeMount := v1.VolumeMount{
	// 	Name:      "brigade-project",
	// 	MountPath: "/etc/brigade-project",
	// 	ReadOnly:  true,
	// }
	// sidecarVolumeMount := v1.VolumeMount{
	// 	Name:      "vcs-sidecar",
	// 	MountPath: "/vcs",
	// }
	// volumeMounts = append(volumeMounts, buildVolumeMount, projectVolumeMount)

	// volumes := []v1.Volume{}
	// buildVolume := v1.Volume{
	// 	Name: buildVolumeMount.Name,
	// 	VolumeSource: v1.VolumeSource{
	// 		Secret: &v1.SecretVolumeSource{SecretName: build.Name},
	// 	},
	// }
	// projectVolume := v1.Volume{
	// 	Name: projectVolumeMount.Name,
	// 	VolumeSource: v1.VolumeSource{
	// 		Secret: &v1.SecretVolumeSource{SecretName: project.Name},
	// 	},
	// }
	// sidecarVolume := v1.Volume{
	// 	Name: sidecarVolumeMount.Name,
	// 	VolumeSource: v1.VolumeSource{
	// 		EmptyDir: &v1.EmptyDirVolumeSource{},
	// 	},
	// }
	// volumes = append(volumes, buildVolume, projectVolume)

	// initContainers := []v1.Container{}
	// // Only add the sidecar resources if sidecar pod image is supplied.
	// if image := project.Data["vcsSidecar"]; len(image) > 0 {
	// 	volumeMounts = append(volumeMounts, sidecarVolumeMount)
	// 	volumes = append(volumes, sidecarVolume)
	// 	initContainers = append(initContainers,
	// 		v1.Container{
	// 			Name:            "vcs-sidecar",
	// 			Image:           string(image),
	// 			ImagePullPolicy: v1.PullPolicy(pullPolicy),
	// 			VolumeMounts:    []v1.VolumeMount{sidecarVolumeMount},
	// 			Env:             env,
	// 			Resources:       vcsSidecarResources(project),
	// 		})
	// }

	// spec := v1.PodSpec{
	// 	ServiceAccountName: config.WorkerServiceAccount,
	// 	NodeSelector: map[string]string{
	// 		"beta.kubernetes.io/os": "linux",
	// 	},
	// 	Containers: []v1.Container{{
	// 		Name:            "brigade-runner",
	// 		Image:           image,
	// 		ImagePullPolicy: v1.PullPolicy(pullPolicy),
	// 		Command:         cmd,
	// 		VolumeMounts:    volumeMounts,
	// 		Env:             env,
	// 		Resources:       workerResources(config),
	// 	}},
	// 	InitContainers: initContainers,
	// 	Volumes:        volumes,
	// 	RestartPolicy:  v1.RestartPolicyNever,
	// }

	// if scriptName := project.Data["defaultScriptName"]; len(scriptName) > 0 {
	// 	attachConfigMap(&spec, string(scriptName), "/etc/brigade-default-script")
	// }

	// if configName := project.Data["defaultConfigName"]; len(configName) > 0 {
	// 	attachConfigMap(&spec, string(configName), "/etc/brigade-default-config")
	// }

	// if ips := project.Data["imagePullSecrets"]; len(ips) > 0 {
	// 	pullSecs := strings.Split(string(ips), ",")
	// 	refs := []v1.LocalObjectReference{}
	// 	for _, pullSec := range pullSecs {
	// 		ref := v1.LocalObjectReference{Name: strings.TrimSpace(pullSec)}
	// 		refs = append(refs, ref)
	// 	}
	// 	spec.ImagePullSecrets = refs
	// }

	// return v1.Pod{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:   build.Name,
	// 		Labels: build.Labels,
	// 	},
	// 	Spec: spec,
	// }

	// // ---------------------------------------------------------------------------

	return nil
}

// ---------------------------------------------------------------------------

// func workerEnv(event brignext.Event, workerName string, worker brignext.Worker) []v1.EnvVar {
// 	envs := []v1.EnvVar{
// 		// Project details:
// 		{Name: "BRIGNEXT_PROJECT_ID", Value: event.ProjectID},
// 		{Name: "BRIGNEXT_PROJECT_NAMESPACE", Value: event.Namespace},
// 		// -------------------------------------------------------------------------
// 		// Event details:
// 		{Name: "BRIGNEXT_EVENT_ID", Value: event.ID},
// 		{Name: "BRIGNEXT_EVENT_PROVIDER", Value: event.Provider},
// 		{Name: "BRIGNEXT_EVENT_TYPE", Value: event.Type},
// 		// -------------------------------------------------------------------------
// 		// Worker details:
// 		// The worker needs to know this because it will probably derive job names
// 		// from its own name:
// 		{Name: "BRIGNEXT_WORKER_NAME", Value: workerName},
// 		// The worker probably doesn't need to know it's own service account:
// 		// {Name: "BRIGNEXT_WORKER_SERVICE_ACCOUNT", Value: worker.Kubernetes.ServiceAccount},
// 		// The worker probably doesn't need to know the storage class:
// 		// {Name: "BRIGNEXT_WORKSPACE_STORAGE_CLASS", Value: worker.Kubernetes.WorkspaceStorageClass},
// 		// -------------------------------------------------------------------------
// 		// Source details:
// 		{Name: "BRIGNEXT_GIT_CLONE_URL", Value: worker.Git.CloneURL}, // TODO: Check for nil; allow to be overridden
// 		{Name: "BRIGNEXT_GIT_COMMIT", Value: worker.Git.Commit},      // TODO: Check for nil
// 		{Name: "BRIGNEXT_GIT_REF", Value: worker.Git.Ref},            // TODO: Check for nil
// 		// -------------------------------------------------------------------------
// 		// Job details:
// 		{Name: "BRIGNEXT_JOBS_SERVICE_ACCOUNT", Value: worker.Jobs.Kubernetes.ServiceAccount},                             // TODO: Check for nil
// 		{Name: "BRIGNEXT_JOBS_ALLOW_SECRET_KEY_REF", Value: strconv.FormatBool(worker.Jobs.Kubernetes.AllowSecretKeyRef)}, // TODO: Check for nil
// 		{Name: "BRIGNEXT_JOBS_CACHE_STORAGE_CLASS", Value: worker.Jobs.Kubernetes.CacheStorageClass},
// 		// -------------------------------------------------------------------------
// 		// Misc.:
// 		{Name: "BRIGNEXT_LOG_LEVEL", Value: string(worker.LogLevel)},
// 		// -------------------------------------------------------------------------
// 		// Things I haven't figured out yet:

// 		{Name: "BRIGADE_SCRIPT", Value: brigadejsPath},
// 		{Name: "BRIGADE_CONFIG", Value: brigadeConfigPath},
// 		{Name: "BRIGNEXT_REPO_KEY", ValueFrom: secretRef("sshKey", project)},
// 		{Name: "BRIGNEXT_REPO_SSH_CERT", ValueFrom: secretRef("sshCert", project)},
// 		{Name: "BRIGNEXT_REPO_AUTH_TOKEN", ValueFrom: secretRef("github.token", project)},
// 		// -------------------------------------------------------------------------
// 		// Things that probably aren't valuable:
// 		// {Name: "BRIGNEXT_WORKSPACE", Value: "/vcs"},
// 	}

// 	return envs
// }

// ---------------------------------------------------------------------------

func getWorkerPodName(eventID, workerName string) string {
	return fmt.Sprintf("%s-%s", eventID, workerName)
}
