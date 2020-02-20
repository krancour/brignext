package storage

import (
	"context"

	"github.com/krancour/brignext"
)

type LogStore interface {
	GetWorkerLogs(eventID string) ([]brignext.LogEntry, error)
	GetWorkerInitLogs(eventID string) ([]brignext.LogEntry, error)
	GetJobLogs(jobID string, containerName string) ([]brignext.LogEntry, error)
	StreamWorkerLogs(ctx context.Context, eventID string) (<-chan brignext.LogEntry, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)
	StreamJobLogs(
		ctx context.Context,
		jobID string,
		containerName string,
	) (<-chan brignext.LogEntry, error)
}
