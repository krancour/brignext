package api

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/krancour/brignext/v2"
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

func (s *server) readAndValidateAPIRequestBody(
	w http.ResponseWriter,
	r *http.Request,
	bodySchemaLoader gojsonschema.JSONLoader,
	bodyObj interface{},
) bool {
	defer r.Body.Close()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.writeResponse(
			w,
			http.StatusBadRequest,
			brignext.NewErrBadRequest("could not read request body"),
		)
		return false
	}
	if bodySchemaLoader != nil {
		validationResult, err := gojsonschema.Validate(
			bodySchemaLoader,
			gojsonschema.NewBytesLoader(bodyBytes),
		)
		if err != nil {
			s.writeResponse(
				w,
				http.StatusBadRequest,
				brignext.NewErrBadRequest("could not validate request body"),
			)
			return false
		}
		if !validationResult.Valid() {
			verrStrs := make([]string, len(validationResult.Errors()))
			for i, verr := range validationResult.Errors() {
				verrStrs[i] = verr.String()
			}
			s.writeResponse(
				w,
				http.StatusBadRequest,
				brignext.NewErrBadRequest(
					"request failed JSON validation",
					verrStrs...,
				),
			)
			return false
		}
	}
	if bodyObj != nil {
		if err = json.Unmarshal(bodyBytes, bodyObj); err != nil {
			s.writeResponse(
				w,
				http.StatusBadRequest,
				brignext.NewErrBadRequest("could not unmarshal request body"),
			)
			return false
		}
	}
	return true
}

func (s *server) serveAPIRequest(apiReq apiRequest) {
	if apiReq.reqBodySchemaLoader != nil || apiReq.reqBodyObj != nil {
		if !s.readAndValidateAPIRequestBody(
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
			s.writeResponse(apiReq.w, http.StatusUnauthorized, e)
		case *brignext.ErrAuthorization:
			s.writeResponse(apiReq.w, http.StatusForbidden, e)
		case *brignext.ErrBadRequest:
			s.writeResponse(apiReq.w, http.StatusBadRequest, e)
		case *brignext.ErrNotFound:
			s.writeResponse(apiReq.w, http.StatusNotFound, e)
		case *brignext.ErrConflict:
			s.writeResponse(apiReq.w, http.StatusConflict, e)
		default:
			log.Println(err)
			s.writeResponse(
				apiReq.w,
				http.StatusInternalServerError,
				brignext.NewErrInternalServer(),
			)
		}
		return
	}
	s.writeResponse(apiReq.w, apiReq.successCode, respBodyObj)
}
