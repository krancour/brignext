package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
)

// Project is Brigade's fundamental management construct. Through a
// ProjectSpec, it pairs EventSubscriptions with a template WorkerSpec.
type Project struct {
	// ObjectMeta contains Project metadata.
	meta.ObjectMeta `json:"metadata" bson:",inline"`
	// Description is a natural language description of the Project.
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	// Spec is an instance of a ProjectSpec that pairs EventSubscriptions with
	// a WorkerTemplate.
	Spec ProjectSpec `json:"spec" bson:"spec"`
	// Kubernetes contains Kubernetes-specific details of the Project's
	// environment.
	Kubernetes *KubernetesConfig `json:"kubernetes,omitempty" bson:"kubernetes,omitempty"` // nolint: lll
}

// MarshalJSON amends Project instances with type metadata.
func (p Project) MarshalJSON() ([]byte, error) {
	type Alias Project
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Project",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectSpec is the technical component of a Project. It pairs
// EventSubscriptions with a prototypical WorkerSpec that is used as a template
// for creating new Workers.
type ProjectSpec struct {
	// EventSubscription defines a set of trigger conditions under which a new
	// Worker should be created.
	EventSubscriptions []EventSubscription `json:"eventSubscriptions,omitempty" bson:"eventSubscriptions,omitempty"` // nolint: lll
	// WorkerTemplate is a prototypical WorkerSpec.
	WorkerTemplate WorkerSpec `json:"workerTemplate" bson:"workerTemplate"`
}

// EventSubscription defines a set of Events of interest. ProjectSpecs utilize
// these in defining the Events that should trigger the execution of a new
// Worker. An Event matches a subscription if it meets ALL of the specified
// criteria.
type EventSubscription struct {
	// Source specifies the origin of an Event (e.g. a gateway).
	Source string `json:"source,omitempty" bson:"source,omitempty"`
	// Types enumerates specific Events of interest from the specified source.
	// This is useful in narrowing a subscription when a source also emits many
	// events that are NOT of interest.
	Types []string `json:"types,omitempty" bson:"types,omitempty"`
	// Labels enumerates specific key/value pairs with which Events of interest
	// must be labeled. An event must have ALL of these labels to match this
	// subscription.
	Labels Labels `json:"labels,omitempty" bson:"labels,omitempty"`
}

// UnmarshalBSON implements custom BSON unmarshaling for the EventSubscription
// type. This does little more than guarantees that the Labels field isn't nil
// so that custom unmarshaling of the EventLabels (which is more involved) can
// succeed.
func (e *EventSubscription) UnmarshalBSON(bytes []byte) error {
	if e.Labels == nil {
		e.Labels = Labels{}
	}
	type EventSubscriptionAlias EventSubscription
	return bson.Unmarshal(
		bytes,
		&struct {
			*EventSubscriptionAlias `bson:",inline"`
		}{
			EventSubscriptionAlias: (*EventSubscriptionAlias)(e),
		},
	)
}

// KubernetesConfig represents Kubernetes-specific configuration. This is used
// primarily at the Project level, but is also denormalized onto Events so that
// Event handling doesn't required a Project lookup to obtain
// Kubernetes-specific configuration.
type KubernetesConfig struct {
	// Namespace is the dedicated Kubernetes namespace for the Project. This is
	// NOT specified by clients when creating a new Project. The namespace is
	// created by / assigned by the system. This detail is a necessity to prevent
	// clients from naming existing namespaces in an attempt to hijack them.
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty"`
}

// ProjectsSelector represents useful filter criteria when selecting multiple
// Projects for API group operations like list. It currently has no fields, but
// exists for future expansion.
type ProjectsSelector struct{}

// ProjectList is an ordered and pageable list of Projects.
type ProjectList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Projects.
	Items []Project `json:"items"`
}

// MarshalJSON amends ProjectList instances with type metadata.
func (p ProjectList) MarshalJSON() ([]byte, error) {
	type Alias ProjectList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectList",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectsService is the specialized interface for managing Projects. It's
// decoupled from underlying technology choices (e.g. data store, message bus,
// etc.) to keep business logic reusable and consistent while the underlying
// tech stack remains free to change.
type ProjectsService interface {
	// Create creates a new Project.
	Create(context.Context, Project) (Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(
		context.Context,
		ProjectsSelector,
		meta.ListOptions,
	) (ProjectList, error)
	// Get retrieves a single Project specified by its identifier.
	Get(context.Context, string) (Project, error)
	// Update updates an existing Project.
	Update(context.Context, Project) (Project, error)
	// Delete deletes a single Project specified by its identifier.
	Delete(context.Context, string) error
}

type projectsService struct {
	authorize            authx.AuthorizeFn
	projectsStore        ProjectsStore
	usersStore           authx.UsersStore
	serviceAccountsStore authx.ServiceAccountsStore
	rolesStore           authx.RolesStore
	substrate            Substrate
}

// NewProjectsService returns a specialized interface for managing Projects.
func NewProjectsService(
	projectsStore ProjectsStore,
	usersStore authx.UsersStore,
	serviceAccountsStore authx.ServiceAccountsStore,
	rolesStore authx.RolesStore,
	substrate Substrate,
) ProjectsService {
	return &projectsService{
		authorize:            authx.Authorize,
		projectsStore:        projectsStore,
		usersStore:           usersStore,
		serviceAccountsStore: serviceAccountsStore,
		rolesStore:           rolesStore,
		substrate:            substrate,
	}
}

func (p *projectsService) Create(
	ctx context.Context,
	project Project,
) (Project, error) {
	if err := p.authorize(ctx, authx.RoleProjectCreator()); err != nil {
		return project, err
	}

	now := time.Now()
	project.Created = &now

	// Add substrate-specific details before we persist.
	var err error
	if project, err = p.substrate.PreCreateProject(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error pre-creating project %q on the substrate",
			project.ID,
		)
	}

	if err = p.projectsStore.Create(ctx, project); err != nil {
		return project,
			errors.Wrapf(err, "error storing new project %q", project.ID)
	}
	if err = p.substrate.CreateProject(ctx, project); err != nil {
		return project, errors.Wrapf(
			err,
			"error creating project %q on the substrate",
			project.ID,
		)
	}

	// Assign roles to the principal who created the project...

	principal := authx.PincipalFromContext(ctx)

	var principalType authx.PrincipalType
	var principalID string
	if user, ok := principal.(*authx.User); ok {
		principalType = authx.PrincipalTypeUser
		principalID = user.ID
	} else if serviceAccount, ok := principal.(*authx.ServiceAccount); ok {
		principalType = authx.PrincipalTypeServiceAccount
		serviceAccount.ID = user.ID
	} else {
		return project, nil
	}

	if err := p.rolesStore.Grant(
		ctx,
		principalType,
		principalID,
		[]authx.Role{
			authx.RoleProjectAdmin(project.ID),
			authx.RoleProjectDeveloper(project.ID),
			authx.RoleProjectUser(project.ID),
		}...,
	); err != nil {
		return Project{}, errors.Wrapf(
			err,
			"error storing project %q roles for %s %q",
			project.ID,
			principalType,
			principalID,
		)
	}

	return project, nil
}

func (p *projectsService) List(
	ctx context.Context,
	selector ProjectsSelector,
	opts meta.ListOptions,
) (ProjectList, error) {
	if err := p.authorize(ctx, authx.RoleReader()); err != nil {
		return ProjectList{}, err
	}

	if opts.Limit == 0 {
		opts.Limit = 20
	}
	projects, err := p.projectsStore.List(ctx, selector, opts)
	if err != nil {
		return projects, errors.Wrap(err, "error retrieving projects from store")
	}
	return projects, nil
}

func (p *projectsService) Get(
	ctx context.Context,
	id string,
) (Project, error) {
	if err := p.authorize(ctx, authx.RoleReader()); err != nil {
		return Project{}, err
	}

	project, err := p.projectsStore.Get(ctx, id)
	if err != nil {
		return project, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			id,
		)
	}
	return project, nil
}

func (p *projectsService) Update(
	ctx context.Context,
	updatedProject Project,
) (Project, error) {
	if err := p.authorize(
		ctx,
		authx.RoleProjectDeveloper(updatedProject.ID),
	); err != nil {
		return Project{}, err
	}

	var err error
	oldProject, err := p.projectsStore.Get(ctx, updatedProject.ID)
	if err != nil {
		return updatedProject, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			updatedProject.ID,
		)
	}

	// Update substrate-specific details before we persist.
	if updatedProject, err =
		p.substrate.PreUpdateProject(ctx, oldProject, updatedProject); err != nil {
		return updatedProject, errors.Wrapf(
			err,
			"error pre-updating project %q on the substrate",
			updatedProject.ID,
		)
	}

	if err = p.projectsStore.Update(ctx, updatedProject); err != nil {
		return updatedProject, errors.Wrapf(
			err,
			"error updating project %q in store",
			updatedProject.ID,
		)
	}
	if err =
		p.substrate.UpdateProject(ctx, oldProject, updatedProject); err != nil {
		return updatedProject, errors.Wrapf(
			err,
			"error updating project %q on the substrate",
			updatedProject.ID,
		)
	}
	return updatedProject, nil
}

func (p *projectsService) Delete(ctx context.Context, id string) error {
	if err := p.authorize(ctx, authx.RoleProjectAdmin(id)); err != nil {
		return err
	}

	project, err := p.projectsStore.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "error retrieving project %q from store", id)
	}

	if err := p.projectsStore.Delete(ctx, id); err != nil {
		return errors.Wrapf(err, "error removing project %q from store", id)
	}
	if err := p.substrate.DeleteProject(ctx, project); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project %q from substrate",
			id,
		)
	}
	return nil
}

type ProjectsStore interface {
	Create(context.Context, Project) error
	List(
		context.Context,
		ProjectsSelector,
		meta.ListOptions,
	) (ProjectList, error)
	ListSubscribers(
		ctx context.Context,
		event Event,
	) (ProjectList, error)
	Get(context.Context, string) (Project, error)
	Update(context.Context, Project) error
	Delete(context.Context, string) error
}
