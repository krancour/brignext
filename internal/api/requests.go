package api

import (
	"net/http"

	"github.com/xeipuuv/gojsonschema"
)

type OutboundRequest struct {
	Method      string
	Path        string
	QueryParams map[string]string
	AuthHeaders map[string]string
	Headers     map[string]string
	ReqBodyObj  interface{}
	SuccessCode int
	RespObj     interface{}
}

type InboundRequest struct {
	W                   http.ResponseWriter
	R                   *http.Request
	ReqBodySchemaLoader gojsonschema.JSONLoader
	ReqBodyObj          interface{}
	EndpointLogic       func() (interface{}, error)
	SuccessCode         int
}
