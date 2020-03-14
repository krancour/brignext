package controller

type workerContext struct {
	EventID  string `json:"event"`
	WorkerID string `json:"worker"`
	doneCh   chan struct{}
}
