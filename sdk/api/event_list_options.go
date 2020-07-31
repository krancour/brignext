package api

import "github.com/krancour/brignext/v2/sdk"

// EventListOptions represents useful filter criteria when selecting multiple
// options for API group operations like list, cancel, or delete.
type EventListOptions struct {
	// ProjectID specifies that Events belonging to the indicated Project should
	// be selected.
	ProjectID string
	// WorkerPhases specifies that Events with their Worker's in any of the
	// indicated phases should be selected.
	WorkerPhases []sdk.WorkerPhase
}
