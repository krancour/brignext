package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/internal/apimachinery"
)

type ProjectsClient interface {
	// TODO: This should return the project because the system will have provided
	// values for some fields that are beyond a client's control, but are not
	// necessarily beyond a client's interest.
	Create(context.Context, brignext.Project) error
	CreateFromBytes(context.Context, []byte) error
	List(context.Context) (brignext.ProjectReferenceList, error)
	Get(context.Context, string) (brignext.Project, error)
	// TODO: This should return the project because the system will have provided
	// values for some fields that are beyond a client's control, but are not
	// necessarily beyond a client's interest.
	Update(context.Context, brignext.Project) error
	UpdateFromBytes(context.Context, string, []byte) error
	Delete(context.Context, string) error

	ListSecrets(
		ctx context.Context,
		projectID string,
	) (brignext.SecretReferenceList, error)
	SetSecret(ctx context.Context, projectID string, secret brignext.Secret) error
	UnsetSecret(ctx context.Context, projectID string, key string) error
}

type projectsClient struct {
	*apimachinery.BaseClient
}

func NewProjectsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ProjectsClient {
	return &projectsClient{
		BaseClient: &apimachinery.BaseClient{
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

func (p *projectsClient) Create(
	_ context.Context,
	project brignext.Project,
) error {
	return p.ExecuteRequest(
		apimachinery.OutboundRequest{
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
		apimachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusCreated,
		},
	)
}

func (p *projectsClient) List(
	context.Context,
) (brignext.ProjectReferenceList, error) {
	projectList := brignext.ProjectReferenceList{}
	return projectList, p.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &projectList,
		},
	)
}

func (p *projectsClient) Get(
	_ context.Context,
	id string,
) (brignext.Project, error) {
	project := brignext.Project{}
	return project, p.ExecuteRequest(
		apimachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
			RespObj:     &project,
		},
	)
}

func (p *projectsClient) Update(
	_ context.Context,
	project brignext.Project,
) error {
	return p.ExecuteRequest(
		apimachinery.OutboundRequest{
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
		apimachinery.OutboundRequest{
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
		apimachinery.OutboundRequest{
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
) (brignext.SecretReferenceList, error) {
	secretList := brignext.SecretReferenceList{}
	return secretList, p.ExecuteRequest(
		apimachinery.OutboundRequest{
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
	secret brignext.Secret,
) error {
	return p.ExecuteRequest(
		apimachinery.OutboundRequest{
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
		apimachinery.OutboundRequest{
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
