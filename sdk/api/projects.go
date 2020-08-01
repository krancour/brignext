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

// ProjectReference is an abridged representation of a Project useful to
// API operations that construct and return potentially large collections of
// projects. Utilizing such an abridged representation both limits response size
// and accounts for the reality that not all clients with authorization to list
// projects are authorized to view the details of every Project.
type ProjectReference struct {
	// ObjectReferenceMeta encapsulates an abridged representation of Project
	// metadata.
	meta.ObjectReferenceMeta `json:"metadata"`
	// Description is a natural language description of the Project.
	Description string `json:"description,omitempty"`
}

// MarshalJSON amends ProjectReference instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (p ProjectReference) MarshalJSON() ([]byte, error) {
	type Alias ProjectReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectReference",
			},
			Alias: (Alias)(p),
		},
	)
}

// ProjectReferenceList is an ordered list of ProjectReferences.
type ProjectReferenceList struct {
	// TODO: When pagination is implemented, list metadata will need to be added
	// Items is a slice of ProjectReferences.
	Items []ProjectReference `json:"items,omitempty"`
}

// MarshalJSON amends ProjectReferenceList instances with type metadata so that
// clients do not need to be concerned with the tedium of doing so.
func (p ProjectReferenceList) MarshalJSON() ([]byte, error) {
	type Alias ProjectReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "ProjectReferenceList",
			},
			Alias: (Alias)(p),
		},
	)
}

type SecretReference struct {
	Key string `json:"key,omitempty"`
}

func (s SecretReference) MarshalJSON() ([]byte, error) {
	type Alias SecretReference
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SecretReference",
			},
			Alias: (Alias)(s),
		},
	)
}

type SecretReferenceList struct {
	Items []SecretReference `json:"items,omitempty"`
}

func (s SecretReferenceList) MarshalJSON() ([]byte, error) {
	type Alias SecretReferenceList
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "SecretReferenceList",
			},
			Alias: (Alias)(s),
		},
	)
}

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
	return p.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  project,
			successCode: http.StatusCreated,
		},
	)
}

func (p *projectsClient) CreateFromBytes(
	_ context.Context,
	projectBytes []byte,
) error {
	return p.executeRequest(
		outboundRequest{
			method:      http.MethodPost,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  projectBytes,
			successCode: http.StatusCreated,
		},
	)
}

func (p *projectsClient) List(
	context.Context,
) (ProjectReferenceList, error) {
	projectList := ProjectReferenceList{}
	return projectList, p.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        "v2/projects",
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &projectList,
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
) error {
	return p.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/projects/%s", project.ID),
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  project,
			successCode: http.StatusOK,
		},
	)
}

func (p *projectsClient) UpdateFromBytes(
	_ context.Context,
	projectID string,
	projectBytes []byte,
) error {
	return p.executeRequest(
		outboundRequest{
			method:      http.MethodPut,
			path:        fmt.Sprintf("v2/projects/%s", projectID),
			authHeaders: p.bearerTokenAuthHeaders(),
			reqBodyObj:  projectBytes,
			successCode: http.StatusOK,
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
) (SecretReferenceList, error) {
	secretList := SecretReferenceList{}
	return secretList, p.executeRequest(
		outboundRequest{
			method:      http.MethodGet,
			path:        fmt.Sprintf("v2/projects/%s/secrets", projectID),
			authHeaders: p.bearerTokenAuthHeaders(),
			successCode: http.StatusOK,
			respObj:     &secretList,
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
