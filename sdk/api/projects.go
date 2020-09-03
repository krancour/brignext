package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/krancour/brignext/v2/sdk"
	"github.com/krancour/brignext/v2/sdk/meta"
)

// ProjectsSelector represents useful filter criteria when selecting multiple
// Projects for API group operations like list. It currently has no fields, but
// exists for future expansion.
type ProjectsSelector struct{}

// ProjectList is an ordered and pageable list of ProjectS.
type ProjectList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Projects.
	Items []sdk.Project `json:"items,omitempty"`
}

// MarshalJSON amends ProjectList instances with type metadata so that clients
// do not need to be concerned with the tedium of doing so.
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

// ProjectsClient is the specialized client for managing Projects with the
// BrigNext API.
type ProjectsClient interface {
	// Create creates a new Project.
	Create(context.Context, sdk.Project) (sdk.Project, error)
	// CreateFromBytes creates a new Project using raw (unprocessed by the client)
	// bytes, presumably originating from a file. This is the preferred way to
	// create Projects defined by an end user since server-side validation will
	// then be applied directly to the Project definition as the user has written
	// it (i.e. WITHOUT any normalization or corrections the client may have made
	// when unmarshaling the original data or when marshaling the outbound
	// request).
	CreateFromBytes(context.Context, []byte) (sdk.Project, error)
	// List returns a ProjectList, with its Items (Projects) ordered
	// alphabetically by Project ID.
	List(context.Context, ProjectsSelector, meta.ListOptions) (ProjectList, error)
	// Get retrieves a single Project specified by its identifier.
	Get(context.Context, string) (sdk.Project, error)
	// Update updates an existing Project.
	Update(context.Context, sdk.Project) (sdk.Project, error)
	// UpdateFromBytes updates an existing Project using raw (unprocessed by the
	// client) bytes, presumably originating from a file. This is the preferred
	// way to update Projects defined by an end user since server-side validation
	// will then be applied directly to the Project definition as the user has
	// written it (i.e. WITHOUT any normalization or corrections the client may
	// have made when unmarshaling the original data or when marshaling the
	// outbound request).
	UpdateFromBytes(context.Context, string, []byte) (sdk.Project, error)
	// Delete deletes a single Project specified by its identifier.
	Delete(context.Context, string) error

	// Roles returns a specialized client for Project Role management.
	Roles() ProjectRolesClient

	// Secrets returns a specialized client for Secret management.
	Secrets() SecretsClient
}

type projectsClient struct {
	*baseClient
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
		rolesClient:   NewProjectRolesClient(apiAddress, apiToken, allowInsecure),
		secretsClient: NewSecretsClient(apiAddress, apiToken, allowInsecure),
	}
}

func (p *projectsClient) Create(
	_ context.Context,
	project sdk.Project,
) (sdk.Project, error) {
	createdProject := sdk.Project{}
	return createdProject, p.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  project,
			successCode: http.StatusCreated,
			respObj:     &createdProject,
		},
	)
}

func (p *projectsClient) CreateFromBytes(
	_ context.Context,
	projectBytes []byte,
) (sdk.Project, error) {
	createdProject := sdk.Project{}
	return createdProject, p.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  projectBytes,
			successCode: http.StatusCreated,
			respObj:     &createdProject,
		},
	)
}

func (p *projectsClient) List(
	_ context.Context,
	_ ProjectsSelector,
	opts meta.ListOptions,
) (ProjectList, error) {
	projects := ProjectList{}
	return projects, p.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			queryParams: p.appendListQueryParams(nil, opts),
			successCode: http.StatusOK,
			respObj:     &projects,
		},
	)
}

func (p *projectsClient) Get(
	_ context.Context,
	id string,
) (sdk.Project, error) {
	project := sdk.Project{}
	return project, p.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/projects/%s", id),
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &project,
		},
	)
}

func (p *projectsClient) Update(
	_ context.Context,
	project sdk.Project,
) (sdk.Project, error) {
	updatedProject := sdk.Project{}
	return updatedProject, p.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/projects/%s", project.ID),
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  project,
			successCode: http.StatusOK,
			respObj:     &updatedProject,
		},
	)
}

func (p *projectsClient) UpdateFromBytes(
	_ context.Context,
	projectID string,
	projectBytes []byte,
) (sdk.Project, error) {
	updatedProject := sdk.Project{}
	return updatedProject, p.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/projects/%s", projectID),
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  projectBytes,
			successCode: http.StatusOK,
			respObj:     &updatedProject,
		},
	)
}

func (p *projectsClient) Delete(_ context.Context, id string) error {
	return p.executeRequest(
		outboundRequest{
			method:      http.MethodDelete,
			path:        fmt.Sprintf("v2/projects/%s", id),
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) Roles() ProjectRolesClient {
	return p.rolesClient
}

func (p *projectsClient) Secrets() SecretsClient {
	return p.secretsClient
}
