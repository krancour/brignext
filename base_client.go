package brignext

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

type apiRequest struct {
	method      string
	path        string
	queryParams map[string]string
	authHeaders map[string]string
	headers     map[string]string
	reqBodyObj  interface{}
	successCode int
	respObj     interface{}
}

type baseClient struct {
	apiAddress string
	apiToken   string
	httpClient *http.Client
}

func (b *baseClient) bearerTokenAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", b.apiToken),
	}
}

// TODO: This needs a better name
func (b *baseClient) doAPIRequest(apiReq apiRequest) error {
	resp, err := b.doAPIRequest2(apiReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if apiReq.respObj != nil {
		respBodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "error reading response body")
		}
		if err := json.Unmarshal(respBodyBytes, apiReq.respObj); err != nil {
			return errors.Wrap(err, "error unmarshaling response body")
		}
	}
	return nil
}

// TODO: This needs a better name
func (b *baseClient) doAPIRequest2(apiReq apiRequest) (*http.Response, error) {
	var reqBodyReader io.Reader
	if apiReq.reqBodyObj != nil {
		switch rb := apiReq.reqBodyObj.(type) {
		case []byte:
			reqBodyReader = bytes.NewBuffer(rb)
		default:
			reqBodyBytes, err := json.Marshal(apiReq.reqBodyObj)
			if err != nil {
				return nil, errors.Wrap(err, "error marshaling request body")
			}
			reqBodyReader = bytes.NewBuffer(reqBodyBytes)
		}
	}

	req, err := http.NewRequest(
		apiReq.method,
		fmt.Sprintf("%s/%s", b.apiAddress, apiReq.path),
		reqBodyReader,
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error creating request %s %s",
			apiReq.method,
			apiReq.path,
		)
	}
	if len(apiReq.queryParams) > 0 {
		q := req.URL.Query()
		for k, v := range apiReq.queryParams {
			q.Set(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}
	for k, v := range apiReq.authHeaders {
		req.Header.Add(k, v)
	}
	for k, v := range apiReq.headers {
		req.Header.Add(k, v)
	}

	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "error invoking API")
	}

	if (apiReq.successCode == 0 && resp.StatusCode != http.StatusOK) ||
		(apiReq.successCode != 0 && resp.StatusCode != apiReq.successCode) {
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

func (b *baseClient) receiveLogStream(
	ctx context.Context,
	reader io.ReadCloser,
	logEntryCh chan<- LogEntry,
	errCh chan<- error,
) {
	defer reader.Close()
	decoder := json.NewDecoder(reader)
	for {
		logEntry := LogEntry{}
		if err := decoder.Decode(&logEntry); err != nil {
			select {
			case errCh <- err:
			case <-ctx.Done():
			}
			return
		}
		select {
		case logEntryCh <- logEntry:
		case <-ctx.Done():
			return
		}
	}
}
