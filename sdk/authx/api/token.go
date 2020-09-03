package api

// Token represents an opaque bearer token used to authenticate to the BrigNext
// API.
type Token struct {
	Value string `json:"value,omitempty"`
}
