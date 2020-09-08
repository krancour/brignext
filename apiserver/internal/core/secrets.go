package core

import (
	"context"
	"encoding/json"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

type Secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (s Secret) MarshalJSON() ([]byte, error) {
	type Alias Secret
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Secret",
			},
			Alias: (Alias)(s),
		},
	)
}

// SecretList is an ordered and pageable list of Secrets.
type SecretList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Secrets.
	Items []Secret `json:"items,omitempty"`
}

func (s SecretList) Len() int {
	return len(s.Items)
}

func (s SecretList) Swap(i, j int) {
	s.Items[i], s.Items[j] = s.Items[j], s.Items[i]
}

func (s SecretList) Less(i, j int) bool {
	return s.Items[i].Key < s.Items[j].Key
}

func (s SecretList) MarshalJSON() ([]byte, error) {
	type Alias SecretList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SecretList",
			},
			Alias: (Alias)(s),
		},
	)
}

type SecretsService interface {
	// List returns a SecretList who Items (Secrets) contain Keys only and
	// not Values (all Value fields are empty). i.e. Once a secret is set, end
	// clients are unable to retrieve values.
	List(
		ctx context.Context,
		projectID string,
		opts meta.ListOptions,
	) (SecretList, error)
	// Set set the value of a new Secret or updates the value of an existing
	// Secret.
	Set(
		ctx context.Context,
		projectID string,
		secret Secret,
	) error
	// Unset clears the value of an existing Secret.
	Unset(ctx context.Context, projectID string, key string) error
}

type secretsService struct {
	authorize     authx.AuthorizeFn
	projectsStore ProjectsStore
	secretsStore  SecretsStore
}

func NewSecretsService(
	projectsStore ProjectsStore,
	secretsStore SecretsStore,
) SecretsService {
	return &secretsService{
		authorize:     authx.Authorize,
		projectsStore: projectsStore,
		secretsStore:  secretsStore,
	}
}

func (s *secretsService) List(
	ctx context.Context,
	projectID string,
	opts meta.ListOptions,
) (SecretList, error) {
	if err := s.authorize(ctx, authx.RoleReader()); err != nil {
		return SecretList{}, err
	}

	secrets := SecretList{}
	project, err := s.projectsStore.Get(ctx, projectID)
	if err != nil {
		return secrets, errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if opts.Limit == 0 {
		opts.Limit = 20
	}
	if secrets, err =
		s.secretsStore.List(ctx, project, opts); err != nil {
		return secrets, errors.Wrapf(
			err,
			"error getting worker secrets for project %q from store",
			projectID,
		)
	}
	return secrets, nil
}

func (s *secretsService) Set(
	ctx context.Context,
	projectID string,
	secret Secret,
) error {
	if err := s.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	project, err := s.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if err := s.secretsStore.Set(ctx, project, secret); err != nil {
		return errors.Wrapf(
			err,
			"error setting secret for project %q worker in store",
			projectID,
		)
	}
	return nil
}

func (s *secretsService) Unset(
	ctx context.Context,
	projectID string,
	key string,
) error {
	if err := s.authorize(ctx, authx.RoleProjectAdmin(projectID)); err != nil {
		return err
	}

	project, err := s.projectsStore.Get(ctx, projectID)
	if err != nil {
		return errors.Wrapf(
			err,
			"error retrieving project %q from store",
			projectID,
		)
	}
	if err :=
		s.secretsStore.Unset(ctx, project, key); err != nil {
		return errors.Wrapf(
			err,
			"error unsetting secrets for project %q worker in store",
			projectID,
		)
	}
	return nil
}

type SecretsStore interface {
	List(ctx context.Context,
		project Project,
		opts meta.ListOptions,
	) (SecretList, error)
	Set(ctx context.Context, project Project, secret Secret) error
	Unset(ctx context.Context, project Project, key string) error
}
