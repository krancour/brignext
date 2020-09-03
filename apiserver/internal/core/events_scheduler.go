package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/queue"
	myk8s "github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

type EventsScheduler interface {
	PreCreate(
		ctx context.Context,
		project Project,
		event Event,
	) (Event, error)
	Create(
		ctx context.Context,
		project Project,
		event Event,
	) error
	Delete(context.Context, Event) error

	StartWorker(ctx context.Context, event Event) error

	CreateJob(
		ctx context.Context,
		event Event,
		jobName string,
	) error
	StartJob(
		ctx context.Context,
		event Event,
		jobName string,
	) error
}

type eventsScheduler struct {
	config             Config
	queueWriterFactory queue.WriterFactory
	kubeClient         *kubernetes.Clientset
}

func NewEventsScheduler(
	config Config,
	queueWriterFactory queue.WriterFactory,
	kubeClient *kubernetes.Clientset,
) EventsScheduler {
	return &eventsScheduler{
		config:             config,
		queueWriterFactory: queueWriterFactory,
		kubeClient:         kubeClient,
	}
}

func (e *eventsScheduler) PreCreate(
	ctx context.Context,
	proj Project,
	event Event,
) (Event, error) {
	// Fill in scheduler-specific details
	event.Kubernetes = proj.Kubernetes
	event.Worker.Spec.Kubernetes = proj.Spec.WorkerTemplate.Kubernetes
	return event, nil
}

func (e *eventsScheduler) Create(
	ctx context.Context,
	proj Project,
	event Event,
) error {
	// Get the project's secrets
	projectSecretsSecret, err := e.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"error finding secret \"project-secrets\" in namespace %q",
			event.Kubernetes.Namespace,
		)
	}
	secrets := map[string]string{}
	for key, value := range projectSecretsSecret.Data {
		secrets[key] = string(value)
	}

	type project struct {
		ID         string            `json:"id"`
		Kubernetes *KubernetesConfig `json:"kubernetes"`
		Secrets    map[string]string `json:"secrets"`
	}

	type worker struct {
		APIAddress           string            `json:"apiAddress"`
		APIToken             string            `json:"apiToken"`
		LogLevel             LogLevel          `json:"logLevel"`
		ConfigFilesDirectory string            `json:"configFilesDirectory"`
		DefaultConfigFiles   map[string]string `json:"defaultConfigFiles" bson:"defaultConfigFiles"` // nolint: lll
	}

	// Create a secret with event details
	eventJSON, err := json.MarshalIndent(
		struct {
			ID         string  `json:"id"`
			Project    project `json:"project"`
			Source     string  `json:"source"`
			Type       string  `json:"type"`
			ShortTitle string  `json:"shortTitle"`
			LongTitle  string  `json:"longTitle"`
			Payload    string  `json:"payload"`
			Worker     worker  `json:"worker"`
		}{
			ID: event.ID,
			Project: project{
				ID:         event.ProjectID,
				Kubernetes: event.Kubernetes,
				Secrets:    secrets,
			},
			Source:     event.Source,
			Type:       event.Type,
			ShortTitle: event.ShortTitle,
			LongTitle:  event.LongTitle,
			Payload:    event.Payload,
			Worker: worker{
				APIAddress:           e.config.APIAddress,
				APIToken:             event.Worker.Token,
				LogLevel:             event.Worker.Spec.LogLevel,
				ConfigFilesDirectory: event.Worker.Spec.ConfigFilesDirectory,
				DefaultConfigFiles:   event.Worker.Spec.DefaultConfigFiles,
			},
		},
		"",
		"  ",
	)
	if err != nil {
		return errors.Wrapf(err, "error marshaling event %q", event.ID)
	}

	data := map[string][]byte{}
	data["event.json"] = eventJSON
	data["gitSSHKey"] = projectSecretsSecret.Data["gitSSHKey"]
	data["gitSSHCert"] = projectSecretsSecret.Data["gitSSHCert"]

	if _, err = e.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("event-%s", event.ID),
				Labels: map[string]string{
					myk8s.ComponentLabel: "event",
					myk8s.ProjectLabel:   event.ProjectID,
					myk8s.EventLabel:     event.ID,
				},
			},
			Type: corev1.SecretType("brignext.io/event"),
			Data: data,
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating secret %q in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Schedule worker for asynchronous execution
	queueWriter, err := e.queueWriterFactory.NewQueueWriter(
		fmt.Sprintf("workers.%s", event.ProjectID),
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating queue writer for project %q workers",
			event.ProjectID,
		)
	}
	defer func() {
		closeCtx, cancelCloseCtx :=
			context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelCloseCtx()
		queueWriter.Close(closeCtx)
	}()

	if err := queueWriter.Write(ctx, event.ID); err != nil {
		return errors.Wrapf(
			err,
			"error submitting execution task for event %q worker",
			event.ID,
		)
	}

	return nil
}

func (e *eventsScheduler) Delete(
	ctx context.Context,
	event Event,
) error {
	matchesEvent, _ := labels.NewRequirement(
		myk8s.EventLabel,
		selection.Equals,
		[]string{event.ID},
	)
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*matchesEvent)

	// Delete all pods related to this event
	if err := e.deletePodsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q pods in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all persistent volume claims related to this event
	if err := e.deletePersistentVolumeClaimsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q persistent volume claims in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all config maps related to this event. BrigNext itself doesn't
	// create any, but we're not discounting the possibility that a worker or job
	// might create some. We are, of course, assuming that anything created by a
	// worker or job is labeled appropriately.
	if err := e.deleteConfigMapsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q config maps in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all secrets related to this event
	if err := e.deleteSecretsByLabelSelector(
		ctx,
		event.Kubernetes.Namespace,
		labelSelector,
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q secrets in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	return nil
}

func (e *eventsScheduler) StartWorker(
	ctx context.Context,
	event Event,
) error {
	if event.Worker.Spec.UseWorkspace {
		if err := e.createWorkspacePVC(ctx, event); err != nil {
			return errors.Wrapf(
				err,
				"error creating workspace for event %q worker",
				event.ID,
			)
		}
	}
	if err := e.createWorkerPod(ctx, event); err != nil {
		return errors.Wrapf(
			err,
			"error creating pod for event %q worker",
			event.ID,
		)
	}
	return nil
}

func (e *eventsScheduler) CreateJob(
	ctx context.Context,
	event Event,
	jobName string,
) error {
	// Schedule job for asynchronous execution
	queueWriter, err := e.queueWriterFactory.NewQueueWriter(
		fmt.Sprintf("jobs.%s", event.ProjectID),
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error creating queue writer for project %q jobs",
			event.ProjectID,
		)
	}
	defer func() {
		closeCtx, cancelCloseCtx :=
			context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelCloseCtx()
		queueWriter.Close(closeCtx)
	}()

	if err := queueWriter.Write(
		ctx,
		fmt.Sprintf("%s:%s", event.ID, jobName),
	); err != nil {
		return errors.Wrapf(
			err,
			"error submitting execution task for event %q job %q",
			event.ID,
			jobName,
		)
	}
	return nil
}

func (e *eventsScheduler) StartJob(
	ctx context.Context,
	event Event,
	jobName string,
) error {
	jobSpec := event.Worker.Jobs[jobName].Spec
	if err := e.createJobSecret(ctx, event, jobName, jobSpec); err != nil {
		return errors.Wrapf(
			err,
			"error creating secret for event %q job %q",
			event.ID,
			jobName,
		)
	}
	if err := e.createJobPod(ctx, event, jobName, jobSpec); err != nil {
		return errors.Wrapf(
			err,
			"error creating pod for event %q job %q",
			event.ID,
			jobName,
		)
	}
	return nil
}

func (e *eventsScheduler) deletePodsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return e.kubeClient.CoreV1().Pods(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (e *eventsScheduler) deletePersistentVolumeClaimsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return e.kubeClient.CoreV1().PersistentVolumeClaims(
		namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (e *eventsScheduler) deleteConfigMapsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return e.kubeClient.CoreV1().ConfigMaps(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (e *eventsScheduler) deleteSecretsByLabelSelector(
	ctx context.Context,
	namespace string,
	labelSelector labels.Selector,
) error {
	return e.kubeClient.CoreV1().Secrets(namespace).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	)
}

func (e *eventsScheduler) createWorkspacePVC(
	ctx context.Context,
	event Event,
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
			StorageClassName: &e.config.WorkspaceStorageClass,
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
		e.kubeClient.CoreV1().PersistentVolumeClaims(event.Kubernetes.Namespace)
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

func (e *eventsScheduler) createWorkerPod(
	ctx context.Context,
	event Event,
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
		event.Worker.Spec.Container = &ContainerSpec{}
	}
	image := event.Worker.Spec.Container.Image
	if image == "" {
		image = e.config.DefaultWorkerImage
	}
	imagePullPolicy := event.Worker.Spec.Container.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = e.config.DefaultWorkerImagePullPolicy
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

	podClient := e.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace)
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

func (e *eventsScheduler) createJobSecret(
	ctx context.Context,
	event Event,
	jobName string,
	jobSpec JobSpec,
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
		jobSecret.StringData[fmt.Sprintf("%e.%s", jobName, k)] = v
	}
	for sidecarName, sidecareSpec := range jobSpec.SidecarContainers {
		for k, v := range sidecareSpec.Environment {
			jobSecret.StringData[fmt.Sprintf("%e.%s", sidecarName, k)] = v
		}
	}

	secretsClient := e.kubeClient.CoreV1().Secrets(event.Kubernetes.Namespace)
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

func (e *eventsScheduler) createJobPod(
	ctx context.Context,
	event Event,
	jobName string,
	jobSpec JobSpec,
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
	}

	// This slice is big enough to hold the primary container AND all (if any)
	// sidecar containers.
	containers := make([]corev1.Container, len(jobSpec.SidecarContainers)+1)

	// The primary container will be the 0 container in this list.
	// nolint: lll
	containers[0] = corev1.Container{
		Name:            jobName, // Primary container takes the job's name
		Image:           jobSpec.PrimaryContainer.Image,
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
					Key: fmt.Sprintf("%e.%s", jobName, key),
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
						Key: fmt.Sprintf("%e.%s", sidecarName, key),
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

	podClient := e.kubeClient.CoreV1().Pods(event.Kubernetes.Namespace)
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

	return nil
}
