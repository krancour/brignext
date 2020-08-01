package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/krancour/brignext/v2/sdk/meta"
	"github.com/pkg/errors"
)

// outboundRequest models of an outbound API call.
type outboundRequest struct {
	// method specifies the HTTP method to be used.
	method string
	// path specifies a path (relative to the root of the API) to be used.
	path string
	// queryParams optionally specifies any URL query parameters to be used.
	queryParams map[string]string
	// authHeaders optionally specifies any authentication headers to be used.
	authHeaders map[string]string
	// headers optionally specifies any miscellaneous HTTP headers to be used.
	headers map[string]string
	// reqBodyObj optionally provides an object that can be marshaled to create
	// the body of the HTTP request.
	reqBodyObj interface{}
	// successCode specifies what HTTP response code should indicate a successful
	// API call.
	successCode int
	// respObj optionally provides an object into which the HTTP response body can
	// be unmarshaled.
	respObj interface{}
}

// Token represents an opaque bearer token used to authenticate to the BrigNext
// API.
type Token struct {
	Value string `json:"value,omitempty"`
}

// MarshalJSON amends Token instances with type metadata so that clients do not
// need to be concerned with the tedium of doing so.
func (t Token) MarshalJSON() ([]byte, error) {
	type Alias Token
	return json.Marshal(
		struct {
			meta.TypeMeta `json:",inline"`
			Alias         `json:",inline"`
		}{
			TypeMeta: meta.TypeMeta{
				APIVersion: meta.APIVersion,
				Kind:       "Token",
			},
			Alias: (Alias)(t),
		},
	)
}

// baseClient provides "API machinery" used by all the specialized API clients.
// Its various functions remove the tedium from common API-related operations
// like managing authentication headers, encoding request bodies, interpretting
// response codes, decoding responses bodies, and more.
type baseClient struct {
	apiAddress string
	apiToken   string
	httpClient *http.Client
}

// basicAuthHeaders, given a username and password, returns a map[string]string
// populated with a Basic Auth header.
func (b *baseClient) basicAuthHeaders(
	username string,
	password string,
) map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf(
			"Basic %s",
			base64.StdEncoding.EncodeToString(
				[]byte(fmt.Sprintf("%s:%s", username, password)),
			),
		),
	}
}

// bearerTokenAuthHeaders returns a map[string]string populated with an
// authentication header that makes use of the client's bearer token.
func (b *baseClient) bearerTokenAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", b.apiToken),
	}
}

// executeRequest accepts one argument-- an outboundRequest-- that models all
// aspects of a single API call in a succinct fashion. Based on this
// information, this function prepares and executes an HTTP request, interprets
// the HTTP response code and decodes the response body into a user-supplied
// type.
func (b *baseClient) executeRequest(req outboundRequest) error {
	resp, err := b.submitRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if req.respObj != nil {
		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "error reading response body")
		}
		if err := json.Unmarshal(respBodyBytes, req.respObj); err != nil {
			return errors.Wrap(err, "error unmarshaling response body")
		}
	}
	return nil
}

// submitRequest accepts one argument-- an outboundRequest-- that models all
// aspects of a single API call in a succinct fashion. Based on this
// information, this function prepares and executes an HTTP request and returns
// the HTTP response. This is a lower-level function than executeRequest().
// It is used by executeRequest(), but is also suitable for uses in cases where
// specialized response handling is required.
func (b *baseClient) submitRequest(
	req outboundRequest,
) (*http.Response, error) {
	var reqBodyReader io.Reader
	if req.reqBodyObj != nil {
		switch rb := req.reqBodyObj.(type) {
		case []byte:
			reqBodyReader = bytes.NewBuffer(rb)
		default:
			reqBodyBytes, err := json.Marshal(req.reqBodyObj)
			if err != nil {
				return nil, errors.Wrap(err, "error marshaling request body")
			}
			reqBodyReader = bytes.NewBuffer(reqBodyBytes)
		}
	}

	r, err := http.NewRequest(
		req.method,
		fmt.Sprintf("%s/%s", b.apiAddress, req.path),
		reqBodyReader,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error creating request %s %s",
			req.method,
			req.path,
		)
	}
	if len(req.queryParams) > 0 {
		q := r.URL.Query()
		for k, v := range req.queryParams {
			q.Set(k, v)
		}
		r.URL.RawQuery = q.Encode()
	}
	for k, v := range req.authHeaders {
		r.Header.Add(k, v)
	}
	for k, v := range req.headers {
		r.Header.Add(k, v)
	}

	resp, err := b.httpClient.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}

	if (req.successCode == 0 && resp.StatusCode != http.StatusOK) ||
		(req.successCode != 0 && resp.StatusCode != req.successCode) {
		// HTTP Response code hints at what sort of error might be in the body
		// of the response
		var apiErr error
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			apiErr = &ErrAuthentication{}
		case http.StatusForbidden:
			apiErr = &ErrAuthorization{}
		case http.StatusBadRequest:
			apiErr = &ErrBadRequest{}
		case http.StatusNotFound:
			apiErr = &ErrNotFound{}
		case http.StatusConflict:
			apiErr = &ErrConflict{}
		case http.StatusInternalServerError:
			apiErr = &ErrInternalServer{}
		default:
			return nil, errors.Errorf("received %d from API server", resp.StatusCode)
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "error reading error response body")
		}
		if err = json.Unmarshal(bodyBytes, apiErr); err != nil {
			return nil, errors.Wrap(err, "error unmarshaling error response body")
		}
		return nil, apiErr
	}
	return resp, nil
}
