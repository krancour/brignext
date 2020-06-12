package api

import (
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

type InboundRequest struct {
	W                   http.ResponseWriter
	R                   *http.Request
	ReqBodySchemaLoader gojsonschema.JSONLoader
	ReqBodyObj          interface{}
	EndpointLogic       func() (interface{}, error)
	SuccessCode         int
}
