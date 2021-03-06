package restmachinery

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

type Endpoints interface {
	Register(router *mux.Router)
}

type BaseEndpoints struct {
	TokenAuthFilter Filter
}

func (b *BaseEndpoints) readAndValidateRequestBody(
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
		b.WriteAPIResponse(
			w,
			http.StatusBadRequest,
			&meta.ErrBadRequest{
				Reason: "Could not read request body.",
			},
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
			b.WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&meta.ErrBadRequest{
					Reason: "Could not validate request body.",
				},
			)
			return false
		}
		if !validationResult.Valid() {
			// We don't bother to log this because this is DEFINITELY a bad request.
			verrStrs := make([]string, len(validationResult.Errors()))
			for i, verr := range validationResult.Errors() {
				verrStrs[i] = verr.String()
			}
			b.WriteAPIResponse(
				w,
				http.StatusBadRequest,
				&meta.ErrBadRequest{
					Reason:  "Request body failed JSON validation",
					Details: verrStrs,
				},
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
			b.WriteAPIResponse(
				w,
				http.StatusInternalServerError,
				&meta.ErrInternalServer{},
			)
			return false
		}
	}
	return true
}

func (b *BaseEndpoints) ServeRequest(req InboundRequest) {
	if req.ReqBodySchemaLoader != nil || req.ReqBodyObj != nil {
		if !b.readAndValidateRequestBody(
			req.W,
			req.R,
			req.ReqBodySchemaLoader,
			req.ReqBodyObj,
		) {
			return
		}
	}
	respBodyObj, err := req.EndpointLogic()
	if err != nil {
		switch e := errors.Cause(err).(type) {
		case *meta.ErrAuthentication:
			b.WriteAPIResponse(req.W, http.StatusUnauthorized, e)
		case *meta.ErrAuthorization:
			b.WriteAPIResponse(req.W, http.StatusForbidden, e)
		case *meta.ErrBadRequest:
			b.WriteAPIResponse(req.W, http.StatusBadRequest, e)
		case *meta.ErrNotFound:
			b.WriteAPIResponse(req.W, http.StatusNotFound, e)
		case *meta.ErrConflict:
			b.WriteAPIResponse(req.W, http.StatusConflict, e)
		case *meta.ErrNotSupported:
			b.WriteAPIResponse(req.W, http.StatusNotImplemented, e)
		case *meta.ErrInternalServer:
			b.WriteAPIResponse(req.W, http.StatusInternalServerError, e)
		default:
			log.Println(err)
			b.WriteAPIResponse(
				req.W,
				http.StatusInternalServerError,
				&meta.ErrInternalServer{},
			)
		}
		return
	}
	b.WriteAPIResponse(req.W, req.SuccessCode, respBodyObj)
}

func (b *BaseEndpoints) WriteAPIResponse(
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

type HumanRequest struct {
	W             http.ResponseWriter
	EndpointLogic func() (interface{}, error)
	SuccessCode   int
}

func (b *BaseEndpoints) ServeHumanRequest(humanReq HumanRequest) {
	respBodyObj, err := humanReq.EndpointLogic()
	if err != nil {
		switch e := errors.Cause(err).(type) {
		case *meta.ErrAuthentication:
			http.Error(humanReq.W, e.Error(), http.StatusUnauthorized)
		case *meta.ErrAuthorization:
			http.Error(humanReq.W, e.Error(), http.StatusForbidden)
		case *meta.ErrBadRequest:
			http.Error(humanReq.W, e.Error(), http.StatusBadRequest)
		case *meta.ErrNotFound:
			http.Error(humanReq.W, e.Error(), http.StatusNotFound)
		case *meta.ErrConflict:
			http.Error(humanReq.W, e.Error(), http.StatusConflict)
		case *meta.ErrNotSupported:
			http.Error(humanReq.W, e.Error(), http.StatusNotImplemented)
		case *meta.ErrInternalServer:
			http.Error(humanReq.W, e.Error(), http.StatusInternalServerError)
		default:
			log.Println(e)
			http.Error(humanReq.W, e.Error(), http.StatusInternalServerError)
		}
		return
	}
	humanReq.W.Header().Set("Content-Type", "text/plain; charset=utf-8")
	humanReq.W.WriteHeader(humanReq.SuccessCode)
	var responseBody []byte
	switch r := respBodyObj.(type) {
	case []byte:
		responseBody = r
	case string:
		responseBody = []byte(r)
	case fmt.Stringer:
		responseBody = []byte(r.String())
	}
	if _, err := humanReq.W.Write(responseBody); err != nil {
		log.Println(errors.Wrap(err, "error writing response body"))
	}
}
