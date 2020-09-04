package api

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk/core"
	"github.com/krancour/brignext/v2/sdk/internal/restmachinery"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// ProjectsClient is the specialized client for managing Projects with the
// BrigNext API.
type ProjectsClient interface {
	// Create creates a new Project.
	Create(context.Context, core.Project) (core.Project, error)
	// CreateFromBytes creates a new Project using raw (unprocessed by the client)
	// bytes, presumably originating from a file. This is the preferred way to
	// create Projects defined by an end user since server-side validation will
	// then be applied directly to the Project definition as the user has written
	// it (i.e. WITHOUT any normalization or corrections the client may have made
	// when unmarshaling the original data or when marshaling the outbound
	// request).
	CreateFromBytes(context.Context, []byte) (core.Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(context.Context, ProjectsSelector, meta.ListOptions) (ProjectList, error)
	// Get retrieves a single Project specified by its identifier.
	Get(context.Context, string) (core.Project, error)
	// Update updates an existing Project.
	Update(context.Context, core.Project) (core.Project, error)
	// UpdateFromBytes updates an existing Project using raw (unprocessed by the
	// client) bytes, presumably originating from a file. This is the preferred
	// way to update Projects defined by an end user since server-side validation
	// will then be applied directly to the Project definition as the user has
	// written it (i.e. WITHOUT any normalization or corrections the client may
	// have made when unmarshaling the original data or when marshaling the
	// outbound request).
	UpdateFromBytes(context.Context, string, []byte) (core.Project, error)
	// Delete deletes a single Project specified by its identifier.
	Delete(context.Context, string) error

	// Roles returns a specialized client for Project Role management.
	Roles() ProjectRolesClient

	// Secrets returns a specialized client for Secret management.
	Secrets() SecretsClient
}

type projectsClient struct {
	*restmachinery.BaseClient
	// rolesClient is a specialized client for Project Role managament.
	rolesClient ProjectRolesClient
	// secretsClient is a specialized client for Secret managament.
	secretsClient SecretsClient
}

// NewProjectsClient returns a specialized client for managing Projects.
func NewProjectsClient(
	apiAddress string,
	apiToken string,
	allowInsecure bool,
) ProjectsClient {
	return &projectsClient{
		BaseClient: &restmachinery.BaseClient{
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
		rolesClient:   NewProjectRolesClient(apiAddress, apiToken, allowInsecure),
		secretsClient: NewSecretsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (p *projectsClient) Create(
	_ context.Context,
	project core.Project,
) (core.Project, error) {
	createdProject := core.Project{}
	return createdProject, p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  project,
			SuccessCode: http.StatusCreated,
			RespObj:     &createdProject,
		},
	)
}

func (p *projectsClient) CreateFromBytes(
	_ context.Context,
	projectBytes []byte,
) (core.Project, error) {
	createdProject := core.Project{}
	return createdProject, p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPost,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusCreated,
			RespObj:     &createdProject,
		},
	)
}

func (p *projectsClient) List(
	_ context.Context,
	_ ProjectsSelector,
	opts meta.ListOptions,
) (ProjectList, error) {
	projects := ProjectList{}
	return projects, p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodGet,
			Path:        "v2/projects",
			AuthHeaders: p.BearerTokenAuthHeaders(),
			QueryParams: p.AppendListQueryParams(nil, opts),
			SuccessCode: http.StatusOK,
			RespObj:     &projects,
		},
	)
}

func (p *projectsClient) Get(
	_ context.Context,
	id string,
) (core.Project, error) {
	project := core.Project{}
	return project, p.ExecuteRequest(
		restmachinery.OutboundRequest{
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
	project core.Project,
) (core.Project, error) {
	updatedProject := core.Project{}
	return updatedProject, p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", project.ID),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  project,
			SuccessCode: http.StatusOK,
			RespObj:     &updatedProject,
		},
	)
}

func (p *projectsClient) UpdateFromBytes(
	_ context.Context,
	projectID string,
	projectBytes []byte,
) (core.Project, error) {
	updatedProject := core.Project{}
	return updatedProject, p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodPut,
			Path:        fmt.Sprintf("v2/projects/%s", projectID),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			ReqBodyObj:  projectBytes,
			SuccessCode: http.StatusOK,
			RespObj:     &updatedProject,
		},
	)
}

func (p *projectsClient) Delete(_ context.Context, id string) error {
	return p.ExecuteRequest(
		restmachinery.OutboundRequest{
			Method:      http.MethodDelete,
			Path:        fmt.Sprintf("v2/projects/%s", id),
			AuthHeaders: p.BearerTokenAuthHeaders(),
			SuccessCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) Roles() ProjectRolesClient {
	return p.rolesClient
}

func (p *projectsClient) Secrets() SecretsClient {
	return p.secretsClient
}
