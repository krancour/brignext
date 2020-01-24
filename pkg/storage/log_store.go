package storage

import (
	"context"

	"github.com/krancour/brignext/pkg/brignext"
)

type LogStore interface {
	GetWorkerLogs(buildID string) ([]brignext.LogEntry, error)
	GetWorkerInitLogs(buildID string) ([]brignext.LogEntry, error)
	GetJobLogs(jobID string, containerName string) ([]brignext.LogEntry, error)
	StreamWorkerLogs(ctx context.Context, buildID string) (<-chan brignext.LogEntry, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		buildID string,
	) (<-chan brignext.LogEntry, error)
	StreamJobLogs(
		ctx context.Context,
		jobID string,
		containerName string,
	) (<-chan brignext.LogEntry, error)
}
