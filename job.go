package brignext

// type JobStatus string

// const (
// 	JobPending   JobStatus = "Pending"
// 	JobRunning   JobStatus = "Running"
// 	JobSucceeded JobStatus = "Succeeded"
// 	JobFailed    JobStatus = "Failed"
// 	JobUnknown   JobStatus = "Unknown"
// )

// // Job is a single job that is executed by the worker that processes and event.
// type Job struct {
// 	ID           string    `json:"id,omitempty" bson:"_id,omitempty"`
// 	Name         string    `json:"name,omitempty" bson:"name,omitempty"`
// 	EventID      string    `json:"eventID,omitempty" bson:"eventID,omitempty"`
// 	Image        string    `json:"image,omitempty" bson:"image,omitempty"`
// 	CreationTime time.Time `json:"creationTime,omitempty" bson:"creationTime,omitempty"`
// 	StartTime    time.Time `json:"startTime,omitempty" bson:"startTime,omitempty"`
// 	EndTime      time.Time `json:"endTime,omitempty" bson:"endTime,omitempty"`
// 	ExitCode     int32     `json:"exitCode,omitempty" bson:"exitCode,omitempty"`
// 	Status       JobStatus `json:"status,omitempty" bson:"status,omitempty"`
// }
