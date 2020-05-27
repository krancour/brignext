package scheduler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/crypto" // nolint: lll
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type ProjectsScheduler interface {
	Create(
		ctx context.Context,
		project brignext.Project,
	) (brignext.Project, error)
	Delete(
		ctx context.Context,
		project brignext.Project,
	) error
	ListSecrets(
		ctx context.Context,
		project brignext.Project,
	) (brignext.SecretList, error)
	SetSecret(
		ctx context.Context,
		project brignext.Project,
		secret brignext.Secret,
	) error
	UnsetSecret(ctx context.Context, project brignext.Project, key string) error
}

type projectsScheduler struct {
	kubeClient *kubernetes.Clientset
}

func NewProjectsScheduler(kubeClient *kubernetes.Clientset) ProjectsScheduler {
	return &projectsScheduler{
		kubeClient: kubeClient,
	}
}

func (p *projectsScheduler) Create(
	ctx context.Context,
	project brignext.Project,
) (brignext.Project, error) {
	// Create a unique namespace name for the project
	project.Kubernetes = &brignext.KubernetesConfig{
		Namespace: strings.ToLower(
			fmt.Sprintf("brignext-%s-%s", project.ID, crypto.NewToken(10)),
		),
	}

	// Create a the project's namespace
	if _, err := p.kubeClient.CoreV1().Namespaces().Create(
		ctx,
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: project.Kubernetes.Namespace,
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating namespace %q for project %q",
			project.Kubernetes.Namespace,
			project.ID,
		)
	}

	// Create an RBAC role for use by all the project's workers
	if _, err := p.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		ctx,
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workers",
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{""},
					Resources: []string{"configmaps", "secrets", "pods", "pods/log"},
					Verbs:     []string{"create", "get", "list", "watch"},
				},
			},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a service account for use by all of the project's workers
	if _, err := p.kubeClient.CoreV1().ServiceAccounts(
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
		return project, errors.Wrapf(
			err,
			"error creating service account \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role binding to associate the workers service account with
	// the workers RBAC role
	if _, err := p.kubeClient.RbacV1().RoleBindings(
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
		return project, errors.Wrapf(
			err,
			"error creating role binding \"workers\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role for use by all of the project's jobs
	if _, err := p.kubeClient.RbacV1().Roles(project.Kubernetes.Namespace).Create(
		ctx,
		&rbacv1.Role{
			ObjectMeta: metav1.ObjectMeta{
				Name: "jobs",
			},
			Rules: []rbacv1.PolicyRule{},
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating role \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a service account for use by all of the project's workers
	if _, err := p.kubeClient.CoreV1().ServiceAccounts(
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
		return project, errors.Wrapf(
			err,
			"error creating service account \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create an RBAC role binding to associate the jobs service account with the
	// jobs RBAC role
	if _, err := p.kubeClient.RbacV1().RoleBindings(
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
		return project, errors.Wrapf(
			err,
			"error creating role binding \"jobs\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	// Create a secret to hold project secrets
	if _, err := p.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Create(
		ctx,
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "project-secrets",
				Labels: map[string]string{
					componentLabel: "project-secrets",
					projectLabel:   project.ID,
				},
			},
			Type: corev1.SecretType("brignext.io/project-secrets"),
		},
		metav1.CreateOptions{},
	); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating secret \"project-secrets\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}

	return project, nil
}

func (p *projectsScheduler) Delete(
	ctx context.Context,
	project brignext.Project,
) error {
	if err := p.kubeClient.CoreV1().Namespaces().Delete(
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

func (p *projectsScheduler) ListSecrets(
	ctx context.Context,
	project brignext.Project,
) (brignext.SecretList, error) {
	secretList := brignext.NewSecretList()

	k8sSecret, err := p.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return secretList, errors.Wrapf(
			err,
			"error retrieving secret \"project-secrets\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}
	secretList.Items = make([]brignext.Secret, len(k8sSecret.Data))
	var i int
	for key := range k8sSecret.Data {
		secretList.Items[i] = brignext.NewSecret(key, "*** REDACTED ***")
		i++
	}
	return secretList, nil
}

func (p *projectsScheduler) SetSecret(
	ctx context.Context,
	project brignext.Project,
	secret brignext.Secret,
) error {
	patch := struct {
		Data map[string]string `json:"data"`
	}{
		Data: map[string]string{
			secret.Key: base64.StdEncoding.EncodeToString([]byte(secret.Value)),
		},
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrapf(
			err,
			"error marshaling patch for project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	if _, err := p.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Patch(
		ctx,
		"project-secrets",
		types.StrategicMergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error patching project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	return nil
}

func (p *projectsScheduler) UnsetSecret(
	ctx context.Context,
	project brignext.Project,
	key string,
) error {
	// Note: If we blindly try to patch the k8s secret to remove the specified
	// key, we'll get an error if that key isn't in the map, so we retrieve the
	// k8s secret and have a peek first. If that key is undefined, we bail early
	// and return no error.
	k8sSecret, err := p.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	if _, ok := k8sSecret.Data[key]; !ok {
		return nil
	}
	patch := []struct {
		Op   string `json:"op"`
		Path string `json:"path"`
	}{
		{
			Op:   "remove",
			Path: fmt.Sprintf("/data/%s", key),
		},
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return errors.Wrapf(
			err,
			"error marshaling patch for project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	if _, err := p.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Patch(
		ctx,
		"project-secrets",
		types.JSONPatchType,
		patchBytes,
		metav1.PatchOptions{},
	); err != nil {
		return errors.Wrapf(
			err,
			"error patching project %q secret in namespace %q",
			project.ID,
			project.Kubernetes.Namespace,
		)
	}
	return nil
}
