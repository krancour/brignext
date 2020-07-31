package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk"
)

type ProjectsClient interface {
	// TODO: This should return the project because the system will have provided
	// values for some fields that are beyond a client's control, but are not
	// necessarily beyond a client's interest.
	Create(context.Context, sdk.Project) error
	CreateFromBytes(context.Context, []byte) error
	List(context.Context) (ProjectReferenceList, error)
	Get(context.Context, string) (sdk.Project, error)
	// TODO: This should return the project because the system will have provided
	// values for some fields that are beyond a client's control, but are not
	// necessarily beyond a client's interest.
	Update(context.Context, sdk.Project) error
	UpdateFromBytes(context.Context, string, []byte) error
	Delete(context.Context, string) error

	ListSecrets(
		ctx context.Context,
		projectID string,
	) (SecretReferenceList, error)
	SetSecret(ctx context.Context, projectID string, secret sdk.Secret) error
	UnsetSecret(ctx context.Context, projectID string, key string) error
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

func (p *projectsClient) Create(
	_ context.Context,
	project sdk.Project,
) error {
	return p.ExecuteRequest(
		OutboundRequest{
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
		OutboundRequest{
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
) (ProjectReferenceList, error) {
	projectList := ProjectReferenceList{}
	return projectList, p.ExecuteRequest(
		OutboundRequest{
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
) (sdk.Project, error) {
	project := sdk.Project{}
	return project, p.ExecuteRequest(
		OutboundRequest{
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
	project sdk.Project,
) error {
	return p.ExecuteRequest(
		OutboundRequest{
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
		OutboundRequest{
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
		OutboundRequest{
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
) (SecretReferenceList, error) {
	secretList := SecretReferenceList{}
	return secretList, p.ExecuteRequest(
		OutboundRequest{
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
	secret sdk.Secret,
) error {
	return p.ExecuteRequest(
		OutboundRequest{
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
		OutboundRequest{
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
