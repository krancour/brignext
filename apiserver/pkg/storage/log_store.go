package storage

import (
	"context"

	"github.com/krancour/brignext/v2"
)

type LogStore interface {
	GetWorkerLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) ([]brignext.LogEntry, error)
	StreamWorkerLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) (<-chan brignext.LogEntry, error)
	GetWorkerInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) ([]brignext.LogEntry, error)
	StreamWorkerInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
	) (<-chan brignext.LogEntry, error)

	GetJobLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) ([]brignext.LogEntry, error)
	StreamJobLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
	GetJobInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) ([]brignext.LogEntry, error)
	StreamJobInitLogs(
		ctx context.Context,
		eventID string,
		workerName string,
		jobName string,
	) (<-chan brignext.LogEntry, error)
}
