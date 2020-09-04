package core

import (
	"context"
)

type WorkersStore interface {
	UpdateSpec(
		ctx context.Context,
		eventID string,
		spec WorkerSpec,
	) error
	// TODO: Add get status, watch status, etc.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
}
