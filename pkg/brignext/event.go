package brignext

import "time"

type Event struct {
	ID          string `json:"id,omitempty" bson:"id"`
	ProjectName string `json:"projectName" bson:"projectName"`
	Provider    string `json:"provider" bson:"provider"`
	Type        string `json:"type" bson:"type"`
	// ShortTitle  string    `json:"shortTitle" bson:"shortTitle"`
	// LongTitle   string    `json:"longTitle" bson:"longTitle"`
	// CloneURL string    `json:"cloneURL" bson:"cloneURL"`
	// Revision *Revision `json:"revision" bson:"revision"`
	// Payload  []byte    `json:"payload" bson:"payload"`
	// Script   []byte    `json:"script" bson:"script"`
	// Config   []byte    `json:"config" bson:"config"`
	Worker *Worker `json:"worker,omitempty" bson:"worker,omitempty"`
	// LogLevel string    `json:"logLevel" bson:"logLevel"`
}

// type Revision struct {
// 	Commit string `json:"commit" bson:"commit"`
// 	Ref    string `json:"ref" bson:"ref"`
// }

type Worker struct {
	Status    JobStatus `json:"status" bson:"status"`
	StartTime time.Time `json:"startTime" bson:"startTime"`
	EndTime   time.Time `json:"endTime" bson:"endTime"`
	ExitCode  int32     `json:"exitCode" bson:"exitCode"`
}
