package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/krancour/brignext/v2/apiserver/internal/lib/crypto"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	myk8s "github.com/krancour/brignext/v2/internal/kubernetes"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type ProjectsScheduler interface {
	PreCreate(
		ctx context.Context,
		project Project,
	) (Project, error)
	Create(
		ctx context.Context,
		project Project,
	) error
	PreUpdate(
		ctx context.Context,
		project Project,
	) (Project, error)
	Update(
		ctx context.Context,
		project Project,
	) error
	Delete(
		ctx context.Context,
		project Project,
	) error
	ListSecrets(
		ctx context.Context,
		project Project,
		opts meta.ListOptions,
	) (SecretList, error)
	SetSecret(
		ctx context.Context,
		project Project,
		secret Secret,
	) error
	UnsetSecret(ctx context.Context, project Project, key string) error
}

type projectsScheduler struct {
	kubeClient *kubernetes.Clientset
}

func NewScheduler(kubeClient *kubernetes.Clientset) ProjectsScheduler {
	return &projectsScheduler{
		kubeClient: kubeClient,
	}
}

func (p *projectsScheduler) PreCreate(
	ctx context.Context,
	project Project,
) (Project, error) {
	// Create a unique namespace name for the project
	project.Kubernetes = &KubernetesConfig{
		Namespace: strings.ToLower(
			fmt.Sprintf("brignext-%s-%s", project.ID, crypto.NewToken(10)),
		),
	}
	project = p.projectWithDefaults(project)
	return project, nil
}

func (p *projectsScheduler) Create(
	ctx context.Context,
	project Project,
) error {
	// Create the project's namespace
	if _, err := p.kubeClient.CoreV1().Namespaces().Create(
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
		return errors.Wrapf(
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
		return errors.Wrapf(
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
		return errors.Wrapf(
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
		return errors.Wrapf(
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
		return errors.Wrapf(
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
		return errors.Wrapf(
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
					myk8s.ComponentLabel: "project-secrets",
					myk8s.ProjectLabel:   project.ID,
				},
			},
			Type: corev1.SecretType("brignext.io/project-secrets"),
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

func (p *projectsScheduler) PreUpdate(
	ctx context.Context,
	project Project,
) (Project, error) {
	return p.projectWithDefaults(project), nil
}

func (p *projectsScheduler) Update(context.Context, Project) error {
	return nil
}

func (p *projectsScheduler) Delete(
	ctx context.Context,
	project Project,
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
	project Project,
	opts meta.ListOptions,
) (SecretList, error) {
	secrets := SecretList{}

	k8sSecret, err := p.kubeClient.CoreV1().Secrets(
		project.Kubernetes.Namespace,
	).Get(ctx, "project-secrets", metav1.GetOptions{})
	if err != nil {
		return secrets, errors.Wrapf(
			err,
			"error retrieving secret \"project-secrets\" in namespace %q",
			project.Kubernetes.Namespace,
		)
	}
	secrets.Items = make([]Secret, len(k8sSecret.Data))
	var i int
	for key := range k8sSecret.Data {
		secrets.Items[i] = Secret{Key: key}
		i++
	}

	sort.Sort(secrets)

	// Paginate...

	// Technically, it's really kind of pointless to do this. The main reason we
	// paginate any sort of response is to avoid causing OOMS by reading gigantic
	// collections (like millions of Events) into memory, but here, all of these
	// secrets are ALREADY in memory, so we're not ready avoiding any real problem
	// here. But we're going to do it anyway just for the sake of making the
	// ListSecrets operation behave consistently with all other list operations.
	if opts.Continue != "" {
		for i := 0; i < len(secrets.Items); i++ {
			if secrets.Items[i].Key == opts.Continue {
				secrets.Items = secrets.Items[i+1:]
				break
			}
		}
	}
	if int64(len(secrets.Items)) > opts.Limit {
		secrets.RemainingItemCount = int64(len(secrets.Items)) - opts.Limit
		secrets.Items = secrets.Items[:opts.Limit]
		secrets.Continue = secrets.Items[opts.Limit-1].Key
	}

	return secrets, nil
}

func (p *projectsScheduler) SetSecret(
	ctx context.Context,
	project Project,
	secret Secret,
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
	project Project,
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

// TODO: We might not want to set this stuff now, but rather defer it to later.
// nolint: lll
func (p *projectsScheduler) projectWithDefaults(
	project Project,
) Project {

	// if project.Spec.WorkerTemplate.JobPolicies == nil {
	// 	project.Spec.WorkerTemplate.JobPolicies = &JobPolicies{}
	// }
	// if project.Spec.WorkerTemplate.JobPolicies.Kubernetes == nil {
	// 	project.Spec.WorkerTemplate.JobPolicies.Kubernetes =
	// 		&KubernetesJobPolicies{}
	// }
	// if project.Spec.WorkerTemplate.JobPolicies.Kubernetes.ImagePullSecrets == nil {
	// 	project.Spec.WorkerTemplate.JobPolicies.Kubernetes.ImagePullSecrets =
	// 		[]string{}
	// }

	return project
}
