package restmachinery

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/brigadecore/brigade/sdk/v2/meta"
	"github.com/stretchr/testify/require"
)

func TestBaseClientBasicAuthHeaders(t *testing.T) {
	const testUsername = "bruce@wayneenterprises.com"
	const testPassword = "ironmansucks"
	client := BaseClient{}
	headers := client.BasicAuthHeaders(testUsername, testPassword)
	header, ok := headers["Authorization"]
	require.True(t, ok)
	require.Contains(t, header, "Basic")
	require.NotContains(t, header, testUsername)
	require.NotContains(t, header, testPassword)
}

func TestBaseClientBearerTokenAuthHeaders(t *testing.T) {
	client := BaseClient{
		APIToken: "11235813213455",
	}
	headers := client.BearerTokenAuthHeaders()
	header, ok := headers["Authorization"]
	require.True(t, ok)
	require.Contains(t, header, "Bearer")
	require.Contains(t, header, client.APIToken)
}

func TestBaseClientAppendListQueryParams(t *testing.T) {
	queryParams := map[string]string{}
	listOpts := meta.ListOptions{
		Continue: "whereileftoff",
		Limit:    10,
	}
	client := BaseClient{}
	queryParams = client.AppendListQueryParams(queryParams, listOpts)
	cntinue, ok := queryParams["continue"]
	require.True(t, ok)
	require.Equal(t, listOpts.Continue, cntinue)
	limitStr, ok := queryParams["limit"]
	require.True(t, ok)
	limit, err := strconv.Atoi(limitStr)
	require.NoError(t, err)
	require.Equal(t, listOpts.Limit, int64(limit))
}

func TestBaseClientExecuteRequest(t *testing.T) {
	type respObjType struct {
		Foo string `json:"foo"`
	}
	testRespObj := respObjType{
		Foo: "bar",
	}
	server := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				bodyBytes, err := json.Marshal(testRespObj)
				require.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				fmt.Fprintln(w, string(bodyBytes))
			},
		),
	)
	client := BaseClient{
		APIAddress: server.URL,
		APIToken:   "11235813213455",
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
	respObj := respObjType{}
	req := OutboundRequest{
		RespObj: &respObj,
	}
	err := client.ExecuteRequest(context.Background(), req)
	require.NoError(t, err)
	require.Equal(t, testRespObj, respObj)
}

func TestBaseClientSubmitRequest(t *testing.T) {
	testCases := []struct {
		name       string
		req        OutboundRequest
		respStatus int
		respBody   []byte
		assertions func(t *testing.T, resp *http.Response, err error)
	}{
		{
			name: "with body object",
			req: OutboundRequest{
				ReqBodyObj: struct {
					Foo string `json:"foo"`
				}{
					Foo: "bar",
				},
			},
			respStatus: http.StatusOK,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
			},
		},
		{
			name: "with body bytes",
			req: OutboundRequest{
				ReqBodyObj: []byte("{}"),
			},
			respStatus: http.StatusOK,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
			},
		},
		{
			name: "with auth header",
			req: OutboundRequest{
				AuthHeaders: map[string]string{
					"Authorization": "Basic dG9ueUBzdGFya2luZHVzdHJpZXMuY29tOmlhbWlyb25tYW4=", // nolint: lll
				},
			},
			respStatus: http.StatusOK,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
			},
		},
		{
			name: "with additional headers",
			req: OutboundRequest{
				Headers: map[string]string{
					"marco": "polo",
				},
			},
			respStatus: http.StatusOK,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
			},
		},
		{
			name: "with query params",
			req: OutboundRequest{
				QueryParams: map[string]string{
					"marco": "polo",
				},
			},
			respStatus: http.StatusOK,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.NoError(t, err)
				require.NotNil(t, resp)
			},
		},
		{
			name:       "with authn error",
			req:        OutboundRequest{},
			respStatus: http.StatusUnauthorized,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.IsType(t, &meta.ErrAuthentication{}, err)
			},
		},
		{
			name:       "with authz error",
			req:        OutboundRequest{},
			respStatus: http.StatusForbidden,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.IsType(t, &meta.ErrAuthorization{}, err)
			},
		},
		{
			name:       "with bad request",
			req:        OutboundRequest{},
			respStatus: http.StatusBadRequest,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.IsType(t, &meta.ErrBadRequest{}, err)
			},
		},
		{
			name:       "with not found",
			req:        OutboundRequest{},
			respStatus: http.StatusNotFound,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.IsType(t, &meta.ErrNotFound{}, err)
			},
		},
		{
			name:       "with conflict",
			req:        OutboundRequest{},
			respStatus: http.StatusConflict,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.IsType(t, &meta.ErrConflict{}, err)
			},
		},
		{
			name:       "with internal server error",
			req:        OutboundRequest{},
			respStatus: http.StatusInternalServerError,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.IsType(t, &meta.ErrInternalServer{}, err)
			},
		},
		{
			name:       "with other error",
			req:        OutboundRequest{},
			respStatus: http.StatusBadGateway,
			assertions: func(t *testing.T, resp *http.Response, err error) {
				require.Equal(
					t,
					fmt.Sprintf("received %d from API server", http.StatusBadGateway),
					err.Error(),
				)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(
				http.HandlerFunc(
					func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(testCase.respStatus)
						fmt.Fprintln(w, "{}")
					},
				),
			)
			client := BaseClient{
				APIAddress: server.URL,
				APIToken:   "11235813213455",
				HTTPClient: &http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: true,
						},
					},
				},
			}
			resp, err := client.SubmitRequest(context.Background(), testCase.req)
			testCase.assertions(t, resp, err)
		})
	}
}

// 	testProject := Project{
// 		ObjectMeta: meta.ObjectMeta{
// 			ID: "bluebook",
// 		},
// 	}
// 	server := httptest.NewServer(
// 		http.HandlerFunc(
// 			func(w http.ResponseWriter, r *http.Request) {
// 				defer r.Body.Close()
// 				require.Equal(t, http.MethodPost, r.Method)
// 				require.Equal(t, "/v2/projects", r.URL.Path)
// 				bodyBytes, err := ioutil.ReadAll(r.Body)
// 				require.NoError(t, err)
// 				project := Project{}
// 				err = json.Unmarshal(bodyBytes, &project)
// 				require.NoError(t, err)
// 				require.Equal(t, testProject, project)
// 				w.WriteHeader(http.StatusCreated)
// 				fmt.Fprintln(w, string(bodyBytes))
// 			},
// 		),
// 	)
// 	defer server.Close()
// 	client := NewProjectsClient(
// 		server.URL,
// 		testAPIToken,
// 		testClientAllowInsecure,
// 	)
// 	project, err := client.Create(
// 		context.Background(),
// 		testProject,
// 	)
// 	require.NoError(t, err)
// 	require.Equal(t, testProject, project)
// }
