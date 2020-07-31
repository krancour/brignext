package api

type LogOptions struct {
	Job       string `json:"job,omitempty"`
	Container string `json:"container,omitempty"`
}
