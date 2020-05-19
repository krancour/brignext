package brignext

type apiRequest struct {
	method      string
	path        string
	queryParams map[string]string
	authHeaders map[string]string
	headers     map[string]string
	reqBodyObj  interface{}
	successCode int
	respObj     interface{}
	errObjs     map[int]error
}
