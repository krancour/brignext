package core

import (
	"context"
)

type JobsStore interface {
	Create(
		ctx context.Context,
		eventID string,
		jobName string,
		jobSpec JobSpec,
	) error
	// TODO: Add get status, watch status, etc.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		jobName string,
		status JobStatus,
	) error
}
