package controller

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/deis/async"
	"github.com/krancour/brignext"
	"github.com/pkg/errors"
)

// workerExecute launches a pod corresponding to the specified worker then
// schedules a follow-up task to monitor that pod for completion.
func (c *controller) workerExecute(
	ctx context.Context,
	task async.Task,
) ([]async.Task, error) {
	eventID, ok := task.GetArgs()["eventID"]
	if !ok {
		return nil, errors.Errorf(
			"executeWorker task %q did not include an event ID argument",
			task.GetID(),
		)
	}

	workerName, ok := task.GetArgs()["worker"]
	if !ok {
		return nil, errors.Errorf(
			"executeWorker task %q did not include a worker name argument",
			task.GetID(),
		)
	}

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
			return nil, nil
		}
		return nil, errors.Wrapf(
			err,
			"error retrieving event %q for worker %q execution",
			eventID,
			workerName,
		)
	}

	worker, ok := event.Workers[workerName]
	if !ok {
		return nil, errors.Errorf(
			"executeWorker task %q failed because event %q did not have a worker "+
				"named %q",
			task.GetID(),
			eventID,
			workerName,
		)
	}

	// There's an unlikely, but non-zero possibility that this handler runs with
	// the worker status already, unexpectedly in a RUNNING state. This could only
	// happen if the handler has already run for this event at least once before
	// and succeeded in updating the worker's status in the database, but the
	// controller process exited unexpectedly before the task completed-- and
	// hence before the follow-up task to monitor the worker was added to the
	// async engine's work queue.
	//
	// So...
	//
	// If the status is already RUNNING, don't do any updates to the database.
	// Just return the follow-up tasks to monitor the worker.
	if worker.Status == brignext.WorkerStatusRunning {
		return []async.Task{
			// A task that will monitor the worker
			async.NewTask(
				"monitorWorker",
				map[string]string{
					"eventID": event.ID,
					"worker":  workerName,
				},
			),
		}, nil
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
		return nil, nil
	}

	// Get the worker pod up and running if it isn't already
	if err := c.createWorkerPod(ctx, event, workerName); err != nil {
		return nil, errors.Wrapf(
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
		return nil, errors.Wrapf(
			err,
			"error updating status on worker %q of event %q",
			workerName,
			event.ID,
		)
	}

	// Schedule a task that will monitor the worker
	return []async.Task{
		// A task that will monitor the worker
		async.NewDelayedTask(
			"monitorWorker",
			map[string]string{
				"eventID": event.ID,
				"worker":  workerName,
			},
			5*time.Second,
		),
	}, nil
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
		event.Namespace,
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

// func workerEnv(project, build *v1.Secret, config *Config) []v1.EnvVar {
// 	allowSecretKeyRef := false
// 	// older projects won't have allowSecretKeyRef set so just check for it
// 	if string(project.Data["kubernetes.allowSecretKeyRef"]) != "" {
// 		var err error
// 		allowSecretKeyRef, err = strconv.ParseBool(string(project.Data["kubernetes.allowSecretKeyRef"]))
// 		if err != nil {
// 			// if we errored parsing the bool something is wrong so just log it and ignore what the project set
// 			log.Printf("error parsing allowSecretKeyRef in project %s: %s", project.Annotations["projectName"], err)
// 		}
// 	}

// 	psv := kube.SecretValues(project.Data)
// 	bsv := kube.SecretValues(build.Data)

// 	serviceAccount := config.ProjectServiceAccount
// 	if string(project.Data["serviceAccount"]) != "" {
// 		serviceAccount = string(project.Data["serviceAccount"])
// 		// Update the service account regex if previously set to the default
// 		if config.ProjectServiceAccountRegex == DefaultJobServiceAccountName {
// 			config.ProjectServiceAccountRegex = serviceAccount
// 		}
// 	}

// 	// Try to get cloneURL from the build first. This allows gateways to override
// 	// the project-level cloneURL if the commit that should be built, for
// 	// instance, exists only within a fork. If this isn't set at the build-level,
// 	// fall back to the project-level default.
// 	cloneURL := bsv.String("clone_url")
// 	if cloneURL == "" {
// 		cloneURL = string(project.Data["cloneURL"])
// 	}

// 	envs := []v1.EnvVar{
// 		{Name: "CI", Value: "true"},
// 		{Name: "BRIGADE_BUILD_ID", Value: build.Labels["build"]},
// 		{Name: "BRIGADE_BUILD_NAME", Value: bsv.String("build_name")},
// 		{Name: "BRIGADE_COMMIT_ID", Value: bsv.String("commit_id")},
// 		{Name: "BRIGADE_COMMIT_REF", Value: bsv.String("commit_ref")},
// 		{Name: "BRIGADE_EVENT_PROVIDER", Value: bsv.String("event_provider")},
// 		{Name: "BRIGADE_EVENT_TYPE", Value: bsv.String("event_type")},
// 		{Name: "BRIGADE_PROJECT_ID", Value: bsv.String("project_id")},
// 		{Name: "BRIGADE_LOG_LEVEL", Value: bsv.String("log_level")},
// 		{Name: "BRIGADE_REMOTE_URL", Value: cloneURL},
// 		{Name: "BRIGADE_WORKSPACE", Value: "/vcs"},
// 		{Name: "BRIGADE_PROJECT_NAMESPACE", Value: build.Namespace},
// 		{Name: "BRIGADE_SERVICE_ACCOUNT", Value: serviceAccount},
// 		{Name: "BRIGADE_SECRET_KEY_REF", Value: strconv.FormatBool(allowSecretKeyRef)},
// 		{
// 			Name:      "BRIGADE_REPO_KEY",
// 			ValueFrom: secretRef("sshKey", project),
// 		},
// 		{
// 			Name:      "BRIGADE_REPO_SSH_CERT",
// 			ValueFrom: secretRef("sshCert", project),
// 		},
// 		{
// 			Name:      "BRIGADE_REPO_AUTH_TOKEN",
// 			ValueFrom: secretRef("github.token", project),
// 		},
// 		{Name: "BRIGADE_DEFAULT_BUILD_STORAGE_CLASS", Value: config.DefaultBuildStorageClass},
// 		{Name: "BRIGADE_DEFAULT_CACHE_STORAGE_CLASS", Value: config.DefaultCacheStorageClass},
// 	}

// 	if config.ProjectServiceAccountRegex != "" {
// 		envs = append(envs, v1.EnvVar{Name: "BRIGADE_SERVICE_ACCOUNT_REGEX", Value: config.ProjectServiceAccountRegex})
// 	}

// 	brigadejsPath := psv.String("brigadejsPath")
// 	if brigadejsPath != "" {
// 		if filepath.IsAbs(brigadejsPath) {
// 			log.Printf("Warning: 'brigadejsPath' is set on Project Secret but will be ignored because provided path '%s' is an absolute path", brigadejsPath)
// 		} else {
// 			envs = append(envs, v1.EnvVar{Name: "BRIGADE_SCRIPT", Value: filepath.Join("/vcs", brigadejsPath)})
// 		}
// 	}

// 	brigadeConfigPath := psv.String("brigadeConfigPath")
// 	if brigadeConfigPath != "" {
// 		if filepath.IsAbs(brigadeConfigPath) {
// 			log.Printf("Warning: 'brigadeConfigPath' is set on Project Secret but will be ignored because provided path '%s' is an absolute path", brigadeConfigPath)
// 		} else {
// 			envs = append(envs, v1.EnvVar{Name: "BRIGADE_CONFIG", Value: filepath.Join("/vcs", brigadeConfigPath)})
// 		}
// 	}

// 	return envs
// }

// ---------------------------------------------------------------------------

func getWorkerPodName(eventID, workerName string) string {
	return fmt.Sprintf("%s-%s", eventID, workerName)
}
