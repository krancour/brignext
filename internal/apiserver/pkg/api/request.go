package api

import (
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

type Request struct {
	W                   http.ResponseWriter
	R                   *http.Request
	ReqBodySchemaLoader gojsonschema.JSONLoader
	ReqBodyObj          interface{}
	EndpointLogic       func() (interface{}, error)
	SuccessCode         int
}
