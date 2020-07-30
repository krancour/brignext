package apimachinery

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
