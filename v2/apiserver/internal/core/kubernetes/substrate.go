package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/crypto"
	"github.com/brigadecore/brigade/v2/apiserver/internal/lib/queue"
	myk8s "github.com/brigadecore/brigade/v2/internal/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
)

type substrate struct {
	config             core.Config
	queueWriterFactory queue.WriterFactory
	kubeClient         *kubernetes.Clientset
}

func NewSubstrate(
	config core.Config,
	queueWriterFactory queue.WriterFactory,
	kubeClient *kubernetes.Clientset,
) core.Substrate {
	return &substrate{
		config:             config,
		queueWriterFactory: queueWriterFactory,
		kubeClient:         kubeClient,
	}
}

func (s *substrate) PreCreateProject(
	ctx context.Context,
	project core.Project,
) (core.Project, error) {
	// Generate and assign a unique Kubernetes namespace name for the Project,
	// but don't create it yet
	project.Kubernetes = &core.KubernetesConfig{
		Namespace: strings.ToLower(
			fmt.Sprintf("brigade-%s-%s", project.ID, crypto.NewToken(10)),
		),
	}
	return project, nil
}

func (s *substrate) CreateProject(
	ctx context.Context,
	project core.Project,
) error {
	// Create the Project's Kubernetes namespace
	if _, err := s.kubeClient.CoreV1().Namespaces().Create(
		ctx,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: project.Kubernetes.Namespace,
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating namespace %q for project %q",
			project.Kubernetes.Namespace,
			project.ID,
		)
	}

	// Create an RBAC role for use by all the project's workers
	if _, err := s.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		ctx,
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
			Rules: []rbacv1.PolicyRule{},
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating role \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a service account for use by all of the Project's Workers
	if _, err := s.kubeClient.CoreV1().ServiceAccounts(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating service account \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role binding to associate the workers service account with
	// the workers RBAC role
	if _, err := s.kubeClient.RbacV1().RoleBindings(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "workers",
					Namespace: project.Kubernetes.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "Role",
				Name: "workers",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating role binding \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role for use by all of the Project's Jobs
	if _, err := s.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		ctx,
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
			Rules: []rbacv1.PolicyRule{},
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating role \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a service account for use by all of the Project's Jobs
	if _, err := s.kubeClient.CoreV1().ServiceAccounts(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating service account \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role binding to associate the jobs service account with the
	// jobs RBAC role
	if _, err := s.kubeClient.RbacV1().RoleBindings(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&rbacv1.RoleBinding{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "jobs",
					Namespace: project.Kubernetes.Namespace,
				},
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "Role",
				Name: "jobs",
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating role binding \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a Kubernetes secret to store the Project's Secrets. Note that the
	// Kubernetes-based implementation of the SecretStore interface will assume
	// this Kubernetes secret exists.
	if _, err := s.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "project-secrets",
				Labels: map[string]string{
					myk8s.LabelComponent: "project-secrets",
					myk8s.LabelProject:   project.ID,
				},
			},
			Type: myk8s.SecretTypeProjectSecrets,
		},
		metav1.CreateOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error creating secret \"project-secrets\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	return nil
}

func (s *substrate) PreUpdateProject(
	ctx context.Context,
	oldProject core.Project,
	newProject core.Project,
) (core.Project, error) {
	// This Kubernetes-specific configuration isn't specified by the end-user, so
	// when an end-user is updating a Project, this information is missing. In
	// order for it to not get completely clobbered, we copy it over from the
	// Project's pre-update state.
	newProject.Kubernetes = oldProject.Kubernetes
	return newProject, nil
}

func (s *substrate) UpdateProject(
	context.Context,
	core.Project,
	core.Project,
) error {
	// There's nothing to do here.
	return nil
}

func (s *substrate) DeleteProject(
	ctx context.Context,
	project core.Project,
) error {
	// Just delete the Project's entire Kubernetes namespace and it should take
	// all other Project resources along with it.
	if err := s.kubeClient.CoreV1().Namespaces().Delete(
		ctx,
		project.Kubernetes.Namespace,
		metav1.DeleteOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting namespace %q",
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func (s *substrate) PreCreateEvent(
	ctx context.Context,
	project core.Project,
	event core.Event,
) (core.Event, error) {
	// Amend the Event with Kubernetes-specific details copied over from the
	// Project
	event.Kubernetes = project.Kubernetes
	event.Worker.Spec.Kubernetes = project.Spec.WorkerTemplate.Kubernetes
	return event, nil
}

func (s *substrate) PreScheduleWorker(
	ctx context.Context,
	event core.Event,
) error {
	// Create a Kubernetes secret containing relevant Event and Project details.
	// This is created PRIOR to scheduling so that these details will reflect an
	// accurate snapshot of Project configuration at the time the Event was
	// created.

	projectSecretsSecret, err := s.kubeClient.CoreV1().Secrets(
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
		ID         string                 `json:"id"`
		Kubernetes *core.KubernetesConfig `json:"kubernetes"`
		Secrets    map[string]string      `json:"secrets"`
	}

	type worker struct {
		APIAddress           string            `json:"apiAddress"`
		APIToken             string            `json:"apiToken"`
		LogLevel             core.LogLevel     `json:"logLevel"`
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
				APIAddress:           s.config.APIAddress,
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

	if _, err = s.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: fmt.Sprintf("event-%s", event.ID),
				Labels: map[string]string{
					myk8s.LabelComponent: "event",
					myk8s.LabelProject:   event.ProjectID,
					myk8s.LabelEvent:     event.ID,
				},
			},
			Type: myk8s.SecretTypeEvent,
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

	return nil
}

func (s *substrate) ScheduleWorker(
	ctx context.Context,
	event core.Event,
) error {
	queueWriter, err := s.queueWriterFactory.NewQueueWriter(
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

func (s *substrate) StartWorker(ctx context.Context, event core.Event) error {
	if event.Worker.Spec.UseWorkspace {
		if err := s.createWorkspacePVC(ctx, event); err != nil {
			return errors.Wrapf(
				err,
				"error creating workspace for event %q worker",
				event.ID,
			)
		}
	}
	if err := s.createWorkerPod(ctx, event); err != nil {
		return errors.Wrapf(
			err,
			"error creating pod for event %q worker",
			event.ID,
		)
	}
	return nil
}

func (s *substrate) PreScheduleJob(
	ctx context.Context,
	event core.Event,
	jobName string,
) error {
	// Nothing to do here. We'll create all resources needed for the Job at start
	// time.
	return nil
}

func (s *substrate) ScheduleJob(
	ctx context.Context,
	event core.Event,
	jobName string,
) error {
	// Schedule job for asynchronous execution
	queueWriter, err := s.queueWriterFactory.NewQueueWriter(
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

func (s *substrate) StartJob(
	ctx context.Context,
	event core.Event,
	jobName string,
) error {
	jobSpec := event.Worker.Jobs[jobName].Spec
	if err := s.createJobSecret(ctx, event, jobName, jobSpec); err != nil {
		return errors.Wrapf(
			err,
			"error creating secret for event %q job %q",
			event.ID,
			jobName,
		)
	}
	if err := s.createJobPod(ctx, event, jobName, jobSpec); err != nil {
		return errors.Wrapf(
			err,
			"error creating pod for event %q job %q",
			event.ID,
			jobName,
		)
	}
	return nil
}

func (s *substrate) DeleteWorkerAndJobs(
	ctx context.Context,
	event core.Event,
) error {
	matchesEvent, _ := labels.NewRequirement(
		myk8s.LabelEvent,
		selection.Equals,
		[]string{event.ID},
	)
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*matchesEvent)

	// Delete all pods related to this Event
	if err := s.kubeClient.CoreV1().Pods(
		event.Kubernetes.Namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q pods in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all persistent volume claims related to this Event
	if err := s.kubeClient.CoreV1().PersistentVolumeClaims(
		event.Kubernetes.Namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
	); err != nil {
		return errors.Wrapf(
			err,
			"error deleting event %q persistent volume claims in namespace %q",
			event.ID,
			event.Kubernetes.Namespace,
		)
	}

	// Delete all secrets related to this Event
	if err := s.kubeClient.CoreV1().Secrets(
		event.Kubernetes.Namespace,
	).DeleteCollection(
		ctx,
		metav1.DeleteOptions{},
		metav1.ListOptions{
			LabelSelector: labelSelector.String(),
		},
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

func (s *substrate) createWorkspacePVC(
	ctx context.Context,
	event core.Event,
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
				myk8s.LabelComponent: "workspace",
				myk8s.LabelProject:   event.ProjectID,
				myk8s.LabelEvent:     event.ID,
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

func (s *substrate) createWorkerPod(
	ctx context.Context,
	event core.Event,
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

	if event.Worker.Spec.Container == nil {
		event.Worker.Spec.Container = &core.ContainerSpec{}
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
				Name: "vcs",
				// TODO: For now, we're using the Brigade 1.x git init image, but for
				// the sake of consistency and to lower the bar for creating additional
				// VCS integrations, we should develop a new/improved image that gets
				// input from a chunk of JSON, just like the actual worker image does.
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
				myk8s.LabelComponent: "worker",
				myk8s.LabelProject:   event.ProjectID,
				myk8s.LabelEvent:     event.ID,
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

func (s *substrate) createJobSecret(
	ctx context.Context,
	event core.Event,
	jobName string,
	jobSpec core.JobSpec,
) error {

	jobSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("job-%s-%s", event.ID, strings.ToLower(jobName)),
			Namespace: event.Kubernetes.Namespace,
			Labels: map[string]string{
				myk8s.LabelComponent: "job",
				myk8s.LabelProject:   event.ProjectID,
				myk8s.LabelEvent:     event.ID,
				myk8s.LabelJob:       jobName,
			},
		},
		Type:       myk8s.SecretTypeJobSecrets,
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

func (s *substrate) createJobPod(
	ctx context.Context,
	event core.Event,
	jobName string,
	jobSpec core.JobSpec,
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
				myk8s.LabelComponent: "job",
				myk8s.LabelProject:   event.ProjectID,
				myk8s.LabelEvent:     event.ID,
				myk8s.LabelJob:       jobName,
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

	return nil
}
