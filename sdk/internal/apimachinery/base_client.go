package apimachinery

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/krancour/brignext/v2/sdk"
	"github.com/pkg/errors"
)

type BaseClient struct {
	APIAddress string
	APIToken   string
	HTTPClient *http.Client
}

func (b *BaseClient) BasicAuthHeaders(
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

func (b *BaseClient) BearerTokenAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", b.APIToken),
	}
}

func (b *BaseClient) ExecuteRequest(req OutboundRequest) error {
	resp, err := b.SubmitRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if req.RespObj != nil {
		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "error reading response body")
		}
		if err := json.Unmarshal(respBodyBytes, req.RespObj); err != nil {
			return errors.Wrap(err, "error unmarshaling response body")
		}
	}
	return nil
}

func (b *BaseClient) SubmitRequest(
	req OutboundRequest,
) (*http.Response, error) {
	var reqBodyReader io.Reader
	if req.ReqBodyObj != nil {
		switch rb := req.ReqBodyObj.(type) {
		case []byte:
			reqBodyReader = bytes.NewBuffer(rb)
		default:
			reqBodyBytes, err := json.Marshal(req.ReqBodyObj)
			if err != nil {
				return nil, errors.Wrap(err, "error marshaling request body")
			}
			reqBodyReader = bytes.NewBuffer(reqBodyBytes)
		}
	}

	r, err := http.NewRequest(
		req.Method,
		fmt.Sprintf("%s/%s", b.APIAddress, req.Path),
		reqBodyReader,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error creating request %s %s",
			req.Method,
			req.Path,
		)
	}
	if len(req.QueryParams) > 0 {
		q := r.URL.Query()
		for k, v := range req.QueryParams {
			q.Set(k, v)
		}
		r.URL.RawQuery = q.Encode()
	}
	for k, v := range req.AuthHeaders {
		r.Header.Add(k, v)
	}
	for k, v := range req.Headers {
		r.Header.Add(k, v)
	}

	resp, err := b.HTTPClient.Do(r)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}

	if (req.SuccessCode == 0 && resp.StatusCode != http.StatusOK) ||
		(req.SuccessCode != 0 && resp.StatusCode != req.SuccessCode) {
		// HTTP Response code hints at what sort of error might be in the body
		// of the response
		var apiErr error
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			apiErr = &sdk.ErrAuthentication{}
		case http.StatusForbidden:
			apiErr = &sdk.ErrAuthorization{}
		case http.StatusBadRequest:
			apiErr = &sdk.ErrBadRequest{}
		case http.StatusNotFound:
			apiErr = &sdk.ErrNotFound{}
		case http.StatusConflict:
			apiErr = &sdk.ErrConflict{}
		case http.StatusInternalServerError:
			apiErr = &sdk.ErrInternalServer{}
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
