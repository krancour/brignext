package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/api/auth"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type apiRequest struct {
	w                   http.ResponseWriter
	r                   *http.Request
	reqBodySchemaLoader gojsonschema.JSONLoader
	reqBodyObj          interface{}
	endpointLogic       func() (interface{}, error)
	successCode         int
}

type baseEndpoints struct {
	tokenAuthFilter auth.Filter
}

func (b *baseEndpoints) readAndValidateAPIRequestBody(
	w http.ResponseWriter,
	r *http.Request,
	bodySchemaLoader gojsonschema.JSONLoader,
	bodyObj interface{},
) bool {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		// Log it in case something is actually wrong...
		log.Println(errors.Wrap(err, "error reading request body"))
		// But we're going to assume this is because the request body is missing, so
		// we'll treat it as a bad request.
		b.writeAPIResponse(
			w,
			http.StatusBadRequest,
			brignext.NewErrBadRequest("Could not read request body."),
		)
		return false
	}
	if bodySchemaLoader != nil {
		var validationResult *gojsonschema.Result
		validationResult, err = gojsonschema.Validate(
			bodySchemaLoader,
			gojsonschema.NewBytesLoader(bodyBytes),
		)
		if err != nil {
			// Log it in case something is actually wrong...
			log.Println(errors.Wrap(err, "error validating request body"))
			// But as long as the schema itself was valid, the most likely scenario
			// here is that the request body wasn't valid JSON, so we'll treat this as
			// a bad request.
			b.writeAPIResponse(
				w,
				http.StatusBadRequest,
				brignext.NewErrBadRequest("Could not validate request body."),
			)
			return false
		}
		if !validationResult.Valid() {
			// We don't bother to log this because this is DEFINITELY a bad request.
			verrStrs := make([]string, len(validationResult.Errors()))
			for i, verr := range validationResult.Errors() {
				verrStrs[i] = verr.String()
			}
			b.writeAPIResponse(
				w,
				http.StatusBadRequest,
				brignext.NewErrBadRequest(
					"Request body failed JSON validation",
					verrStrs...,
				),
			)
			return false
		}
	}
	if bodyObj != nil {
		if err = json.Unmarshal(bodyBytes, bodyObj); err != nil {
			log.Println(errors.Wrap(err, "error marshaling request body"))
			// We were already able to validate the request body, which means it was
			// valid JSON. If something went wrong with marshaling, it's NOT because
			// of a bad request-- it's a real, internal problem.
			b.writeAPIResponse(
				w,
				http.StatusInternalServerError,
				brignext.NewErrInternalServer(),
			)
			return false
		}
	}
	return true
}

func (b *baseEndpoints) serveAPIRequest(apiReq apiRequest) {
	if apiReq.reqBodySchemaLoader != nil || apiReq.reqBodyObj != nil {
		if !b.readAndValidateAPIRequestBody(
			apiReq.w,
			apiReq.r,
			apiReq.reqBodySchemaLoader,
			apiReq.reqBodyObj,
		) {
			return
		}
	}
	respBodyObj, err := apiReq.endpointLogic()
	if err != nil {
		switch e := errors.Cause(err).(type) {
		case *brignext.ErrAuthentication:
			b.writeAPIResponse(apiReq.w, http.StatusUnauthorized, e)
		case *brignext.ErrAuthorization:
			b.writeAPIResponse(apiReq.w, http.StatusForbidden, e)
		case *brignext.ErrBadRequest:
			b.writeAPIResponse(apiReq.w, http.StatusBadRequest, e)
		case *brignext.ErrNotFound:
			b.writeAPIResponse(apiReq.w, http.StatusNotFound, e)
		case *brignext.ErrConflict:
			b.writeAPIResponse(apiReq.w, http.StatusConflict, e)
		case *brignext.ErrNotSupported:
			b.writeAPIResponse(apiReq.w, http.StatusNotImplemented, e)
		case *brignext.ErrInternalServer:
			b.writeAPIResponse(apiReq.w, http.StatusInternalServerError, e)
		default:
			log.Println(err)
			b.writeAPIResponse(
				apiReq.w,
				http.StatusInternalServerError,
				brignext.NewErrInternalServer(),
			)
		}
		return
	}
	b.writeAPIResponse(apiReq.w, apiReq.successCode, respBodyObj)
}

func (b *baseEndpoints) writeAPIResponse(
	w http.ResponseWriter,
	statusCode int,
	response interface{},
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	responseBody, ok := response.([]byte)
	if !ok {
		var err error
		if responseBody, err = json.Marshal(response); err != nil {
			log.Println(errors.Wrap(err, "error marshaling response body"))
		}
	}
	if _, err := w.Write(responseBody); err != nil {
		log.Println(errors.Wrap(err, "error writing response body"))
	}
}

type humanRequest struct {
	w             http.ResponseWriter
	endpointLogic func() (interface{}, error)
	successCode   int
}

func (b *baseEndpoints) serveHumanRequest(humanReq humanRequest) {
	respBodyObj, err := humanReq.endpointLogic()
	if err != nil {
		switch e := errors.Cause(err).(type) {
		case *brignext.ErrAuthentication:
			http.Error(humanReq.w, e.Error(), http.StatusUnauthorized)
		case *brignext.ErrAuthorization:
			http.Error(humanReq.w, e.Error(), http.StatusForbidden)
		case *brignext.ErrBadRequest:
			http.Error(humanReq.w, e.Error(), http.StatusBadRequest)
		case *brignext.ErrNotFound:
			http.Error(humanReq.w, e.Error(), http.StatusNotFound)
		case *brignext.ErrConflict:
			http.Error(humanReq.w, e.Error(), http.StatusConflict)
		case *brignext.ErrNotSupported:
			http.Error(humanReq.w, e.Error(), http.StatusNotImplemented)
		case *brignext.ErrInternalServer:
			http.Error(humanReq.w, e.Error(), http.StatusInternalServerError)
		default:
			log.Println(e)
			http.Error(humanReq.w, e.Error(), http.StatusInternalServerError)
		}
		return
	}
	humanReq.w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	humanReq.w.WriteHeader(humanReq.successCode)
	var responseBody []byte
	switch r := respBodyObj.(type) {
	case []byte:
		responseBody = r
	case string:
		responseBody = []byte(r)
	case fmt.Stringer:
		responseBody = []byte(r.String())
	}
	if _, err := humanReq.w.Write(responseBody); err != nil {
		log.Println(errors.Wrap(err, "error writing response body"))
	}
}
