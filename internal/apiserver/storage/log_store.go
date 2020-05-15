package storage

import (
	"context"

	"github.com/krancour/brignext/v2"
)

type LogStore interface {
	GetWorkerLogs(
		ctx context.Context,
		eventID string,
	) (brignext.LogEntryList, error)
	StreamWorkerLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)
	GetWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (brignext.LogEntryList, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
	) (<-chan brignext.LogEntry, error)

	GetJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (brignext.LogEntryList, error)
	StreamJobLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
	GetJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (brignext.LogEntryList, error)
	StreamJobInitLogs(
		ctx context.Context,
		eventID string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
}
