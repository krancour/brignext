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

// Job is a single job that is executed when a build is triggered for an event.
type Job struct {
	ID           string    `json:"id" bson:"id"`
	Name         string    `json:"name" bson:"name"`
	BuildID      string    `json:"buildID" bson:"buildID"`
	Image        string    `json:"image" bson:"image"`
	CreationTime time.Time `json:"creationTime" bson:"creationTime"`
	StartTime    time.Time `json:"startTime" bson:"startTime"`
	EndTime      time.Time `json:"endTime" bson:"endTime"`
	ExitCode     int32     `json:"exitCode" bson:"exitCode"`
	Status       JobStatus `json:"status" bson:"status"`
}
