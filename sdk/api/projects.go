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

// ProjectListOptions represents useful filter criteria when selecting multiple
// Projects for API group operations like list.
type ProjectListOptions struct {
	// Continue aids in pagination of long lists. It permits clients to echo an
	// opaque value obtained from a previous API call back to the API in a
	// subsequent call in order to indicate what resource was the last on the
	// previous page.
	Continue string
	// Limit aids in pagination of long lists. It permits clients to specify page
	// size when making API calls. The API server provides a default when a value
	// is not specified and may reject or override invalid values (non-positive)
	// numbers or very large page sizes.
	Limit int64
}

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

// SecretListOptions represents useful filter criteria when selecting multiple
// Secrets for API group operations like list.
type SecretListOptions struct {
	// Continue aids in pagination of long lists. It permits clients to echo an
	// opaque value obtained from a previous API call back to the API in a
	// subsequent call in order to indicate what resource was the last on the
	// previous page.
	Continue string
	// Limit aids in pagination of long lists. It permits clients to specify page
	// size when making API calls. The API server provides a default when a value
	// is not specified and may reject or override invalid values (non-positive)
	// numbers or very large page sizes.
	Limit int64
}

// SecretList is an ordered and pageable list of Secrets.
type SecretList struct {
	// ListMeta contains list metadata.
	meta.ListMeta `json:"metadata"`
	// Items is a slice of Secrets.
	Items []sdk.Secret `json:"items,omitempty"`
}

// MarshalJSON amends SecretList instances with type metadata so that clients do
// not need to be concerned with the tedium of doing so.
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
	List(context.Context, ProjectListOptions) (ProjectList, error)
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

	// ListSecrets returns a SecretList who Items (Secrets) contain Keys only and
	// not Values (all Value fields are empty). i.e. Once a secret is set, end
	// clients are unable to retrieve values.
	ListSecrets(
		ctx context.Context,
		projectID string,
		opts SecretListOptions,
	) (SecretList, error)
	// SetSecret set the value of a new Secret or updates the value of an existing
	// Secret.
	SetSecret(ctx context.Context, projectID string, secret sdk.Secret) error
	// UnsetSecret clears the value of an existing Secret.
	UnsetSecret(ctx context.Context, projectID string, key string) error
}

type projectsClient struct {
	*baseClient
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
	opts ProjectListOptions,
) (ProjectList, error) {
	projects := ProjectList{}
	return projects, p.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			queryParams: p.appendListQueryParams(nil, opts.Continue, opts.Limit),
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

func (p *projectsClient) ListSecrets(
	ctx context.Context,
	projectID string,
	opts SecretListOptions,
) (SecretList, error) {
	secrets := SecretList{}
	return secrets, p.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			authHeaders: p.bearerTokenAuthHeaders(),
			queryParams: p.appendListQueryParams(nil, opts.Continue, opts.Limit),
			successCode: http.StatusOK,
			respObj:     &secrets,
		},
	)
}

func (p *projectsClient) SetSecret(
	ctx context.Context,
	projectID string,
	secret sdk.Secret,
) error {
	return p.executeRequest(
		outboundRequest{
			method: http.MethodPut,
			path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				secret.Key,
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
	key string,
) error {
	return p.executeRequest(
		outboundRequest{
			method: http.MethodDelete,
			path: fmt.Sprintf(
				"v2/projects/%s/secrets/%s",
				projectID,
				key,
			),
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
		},
	)
}
