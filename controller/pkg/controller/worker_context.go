package controller

type workerContext struct {
	EventID    string `json:"event"`
	WorkerName string `json:"worker"`
}
