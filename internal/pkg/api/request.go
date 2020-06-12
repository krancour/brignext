package api

type Request struct {
	Method      string
	Path        string
	QueryParams map[string]string
	AuthHeaders map[string]string
	Headers     map[string]string
	ReqBodyObj  interface{}
	SuccessCode int
	RespObj     interface{}
}
