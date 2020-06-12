package projects

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/pkg/api"
)

type Client interface {
	Create(context.Context, brignext.Project) error
	CreateFromBytes(context.Context, []byte) error
	List(context.Context) (brignext.ProjectList, error)
	Get(context.Context, string) (brignext.Project, error)
	Update(context.Context, brignext.Project) error
	UpdateFromBytes(context.Context, string, []byte) error
	Delete(context.Context, string) error

	ListSecrets(ctx context.Context, projectID string) (brignext.SecretList, error)
	SetSecret(ctx context.Context, projectID string, secret brignext.Secret) error
	UnsetSecret(ctx context.Context, projectID string, key string) error
}

type client struct {
	*api.BaseClient
}

func NewClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) Client {
	return &client{
		BaseClient: &api.BaseClient{
			APIAddress: apiAddress,
			APIToken:   apiToken,
			HTTPClient: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: allowInsecure,
					},
				},
			},
		},
	}
}

func (c *client) Create(_ context.Context, project brignext.Project) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  project,
			SuccessCode: http.StatusCreated,
		},
	)
}

func (c *client) CreateFromBytes(
	_ context.Context,
	projectBytes []byte,
) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusCreated,
		},
	)
}

func (c *client) List(context.Context) (brignext.ProjectList, error) {
	projectList := brignext.ProjectList{}
	return projectList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/projects",
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &projectList,
		},
	)
}

func (c *client) Get(_ context.Context, id string) (brignext.Project, error) {
	project := brignext.Project{}
	return project, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &project,
		},
	)
}

func (c *client) Update(_ context.Context, project brignext.Project) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", project.ID),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  project,
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) UpdateFromBytes(
	_ context.Context,
	projectID string,
	projectBytes []byte,
) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", projectID),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) Delete(_ context.Context, id string) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) ListSecrets(
	ctx context.Context,
	projectID string,
) (brignext.SecretList, error) {
	secretList := brignext.SecretList{}
	return secretList, c.ExecuteRequest(
		api.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &secretList,
		},
	)
}

func (c *client) SetSecret(
	ctx context.Context,
	projectID string,
	secret brignext.Secret,
) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secret.Key,
			),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			ReqBodyObj:  secret,
			SuccessCode: http.StatusOK,
		},
	)
}

func (c *client) UnsetSecret(
	ctx context.Context,
	projectID string,
	key string,
) error {
	return c.ExecuteRequest(
		api.OutboundRequest{
			Method: http.MethodDelete,
			Path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				key,
			),
			AuthHeaders: c.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
