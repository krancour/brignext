package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	errs "github.com/krancour/brignext/v2/internal/pkg/errors"
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

func (b *BaseClient) ExecuteRequest(apiReq Request) error {
	resp, err := b.SubmitRequest(apiReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if apiReq.RespObj != nil {
		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "error reading response body")
		}
		if err := json.Unmarshal(respBodyBytes, apiReq.RespObj); err != nil {
			return errors.Wrap(err, "error unmarshaling response body")
		}
	}
	return nil
}

func (b *BaseClient) SubmitRequest(
	apiReq Request,
) (*http.Response, error) {
	var reqBodyReader io.Reader
	if apiReq.ReqBodyObj != nil {
		switch rb := apiReq.ReqBodyObj.(type) {
		case []byte:
			reqBodyReader = bytes.NewBuffer(rb)
		default:
			reqBodyBytes, err := json.Marshal(apiReq.ReqBodyObj)
			if err != nil {
				return nil, errors.Wrap(err, "error marshaling request body")
			}
			reqBodyReader = bytes.NewBuffer(reqBodyBytes)
		}
	}

	req, err := http.NewRequest(
		apiReq.Method,
		fmt.Sprintf("%s/%s", b.APIAddress, apiReq.Path),
		reqBodyReader,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error creating request %s %s",
			apiReq.Method,
			apiReq.Path,
		)
	}
	if len(apiReq.QueryParams) > 0 {
		q := req.URL.Query()
		for k, v := range apiReq.QueryParams {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	for k, v := range apiReq.AuthHeaders {
		req.Header.Add(k, v)
	}
	for k, v := range apiReq.Headers {
		req.Header.Add(k, v)
	}

	resp, err := b.HTTPClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}

	if (apiReq.SuccessCode == 0 && resp.StatusCode != http.StatusOK) ||
		(apiReq.SuccessCode != 0 && resp.StatusCode != apiReq.SuccessCode) {
		// HTTP Response code hints at what sort of error might be in the body
		// of the response
		var apiErr error
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			apiErr = &errs.ErrAuthentication{}
		case http.StatusForbidden:
			apiErr = &errs.ErrAuthorization{}
		case http.StatusBadRequest:
			apiErr = &errs.ErrBadRequest{}
		case http.StatusNotFound:
			apiErr = &errs.ErrNotFound{}
		case http.StatusConflict:
			apiErr = &errs.ErrConflict{}
		case http.StatusInternalServerError:
			apiErr = &errs.ErrInternalServer{}
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
