package brignext

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
)

type ProjectsClient interface {
	Create(context.Context, Project) error
	List(context.Context) (ProjectList, error)
	Get(context.Context, string) (Project, error)
	Update(context.Context, Project) error
	Delete(context.Context, string) error

	ListSecrets(ctx context.Context, projectID string) (SecretList, error)
	SetSecret(ctx context.Context, projectID string, secret Secret) error
	UnsetSecret(ctx context.Context, projectID string, secretID string) error
}

type projectsClient struct {
	*baseClient
}

func NewProjectsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ProjectsClient {
	return &projectsClient{
		baseClient: &baseClient{
			apiAddress: apiAddress,
			apiToken:   apiToken,
			httpClient: &http.Client{
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
	return p.doAPIRequest(
		apiRequest{
			method:      http.MethodPost,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  project,
			successCode: http.StatusCreated,
		},
	)
}

func (p *projectsClient) List(context.Context) (ProjectList, error) {
	projectList := ProjectList{}
	err := p.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &projectList,
		},
	)
	return projectList, err
}

func (p *projectsClient) Get(_ context.Context, id string) (Project, error) {
	project := Project{}
	err := p.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/projects/%s", id),
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &project,
		},
	)
	return project, err
}

func (p *projectsClient) Update(_ context.Context, project Project) error {
	return p.doAPIRequest(
		apiRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/projects/%s", project.ID),
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  project,
			successCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) Delete(_ context.Context, id string) error {
	return p.doAPIRequest(
		apiRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/projects/%s", id),
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) ListSecrets(
	ctx context.Context,
	projectID string,
) (SecretList, error) {
	secretList := SecretList{}
	err := p.doAPIRequest(
		apiRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &secretList,
		},
	)
	return secretList, err
}

func (p *projectsClient) SetSecret(
	ctx context.Context,
	projectID string,
	secret Secret,
) error {
	return p.doAPIRequest(
		apiRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secret.ID,
			),
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  secret,
			successCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) UnsetSecret(
	ctx context.Context,
	projectID string,
	secretID string,
) error {
	return p.doAPIRequest(
		apiRequest{
			method: http.MethodDelete,
			path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secretID,
			),
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}
