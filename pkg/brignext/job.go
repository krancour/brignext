package brignext

import "time"

type JobStatus string

const (
	JobPending   JobStatus = "Pending"
	JobRunning   JobStatus = "Running"
	JobSucceeded JobStatus = "Succeeded"
	JobFailed    JobStatus = "Failed"
	JobUnknown   JobStatus = "Unknown"
)

// Job is a single job that is executed by the worker that processes and event.
type Job struct {
	ID           string    `json:"id" bson:"id"`
	Name         string    `json:"name" bson:"name"`
	EventID      string    `json:"eventID" bson:"eventID"`
	Image        string    `json:"image" bson:"image"`
	CreationTime time.Time `json:"creationTime" bson:"creationTime"`
	StartTime    time.Time `json:"startTime" bson:"startTime"`
	EndTime      time.Time `json:"endTime" bson:"endTime"`
	ExitCode     int32     `json:"exitCode" bson:"exitCode"`
	Status       JobStatus `json:"status" bson:"status"`
}
