package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/internal/pkg/api"
)

type ProjectsClient interface {
	Create(context.Context, Project) error
	CreateFromBytes(context.Context, []byte) error
	List(context.Context) (ProjectList, error)
	Get(context.Context, string) (Project, error)
	Update(context.Context, Project) error
	UpdateFromBytes(context.Context, string, []byte) error
	Delete(context.Context, string) error

	ListSecrets(ctx context.Context, projectID string) (SecretList, error)
	SetSecret(ctx context.Context, projectID string, secret Secret) error
	UnsetSecret(ctx context.Context, projectID string, key string) error
}

type projectsClient struct {
	*api.BaseClient
}

func NewProjectsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ProjectsClient {
	return &projectsClient{
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

func (p *projectsClient) Create(_ context.Context, project Project) error {
	return p.ExecuteRequest(
		api.Request{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  project,
			SuccessCode: http.StatusCreated,
		},
	)
}

func (p *projectsClient) CreateFromBytes(
	_ context.Context,
	projectBytes []byte,
) error {
	return p.ExecuteRequest(
		api.Request{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusCreated,
		},
	)
}

func (p *projectsClient) List(context.Context) (ProjectList, error) {
	projectList := ProjectList{}
	return projectList, p.ExecuteRequest(
		api.Request{
			Method:      http.MethodGet,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &projectList,
		},
	)
}

func (p *projectsClient) Get(_ context.Context, id string) (Project, error) {
	project := Project{}
	return project, p.ExecuteRequest(
		api.Request{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &project,
		},
	)
}

func (p *projectsClient) Update(_ context.Context, project Project) error {
	return p.ExecuteRequest(
		api.Request{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", project.ID),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  project,
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) UpdateFromBytes(
	_ context.Context,
	projectID string,
	projectBytes []byte,
) error {
	return p.ExecuteRequest(
		api.Request{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", projectID),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) Delete(_ context.Context, id string) error {
	return p.ExecuteRequest(
		api.Request{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) ListSecrets(
	ctx context.Context,
	projectID string,
) (SecretList, error) {
	secretList := SecretList{}
	return secretList, p.ExecuteRequest(
		api.Request{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &secretList,
		},
	)
}

func (p *projectsClient) SetSecret(
	ctx context.Context,
	projectID string,
	secret Secret,
) error {
	return p.ExecuteRequest(
		api.Request{
			Method: http.MethodPut,
			Path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secret.Key,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  secret,
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) UnsetSecret(
	ctx context.Context,
	projectID string,
	key string,
) error {
	return p.ExecuteRequest(
		api.Request{
			Method: http.MethodDelete,
			Path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				key,
			),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}
