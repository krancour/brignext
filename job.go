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
// 	ID           string    `json:"id,omitempty" bson:"_id,"`
// 	Name         string    `json:"name,omitempty" bson:"name,"`
// 	EventID      string    `json:"eventID,omitempty" bson:"eventID,"`
// 	Image        string    `json:"image,omitempty" bson:"image,"`
// 	CreationTime time.Time `json:"creationTime,omitempty" bson:"creationTime,"`
// 	StartTime    time.Time `json:"startTime,omitempty" bson:"startTime,"`
// 	EndTime      time.Time `json:"endTime,omitempty" bson:"endTime,"`
// 	ExitCode     int32     `json:"exitCode,omitempty" bson:"exitCode,"`
// 	Status       JobStatus `json:"status,omitempty" bson:"status,"`
// }
