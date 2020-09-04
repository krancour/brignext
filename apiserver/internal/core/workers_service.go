package core

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
)

type WorkersService interface {
	// Start starts the indicated Event's Worker on BrigNext's workload
	// execution substrate.
	Start(ctx context.Context, eventID string) error
	// GetStatus returns an Event's Worker's status.
	GetStatus(
		ctx context.Context,
		eventID string,
	) (WorkerStatus, error)
	// WatchStatus returns a channel over which an Event's Worker's status
	// is streamed. The channel receives a new WorkerStatus every time there is
	// any change in that status.
	WatchStatus(
		ctx context.Context,
		eventID string,
	) (<-chan WorkerStatus, error)
	// UpdateStatus updates the status of an Event's Worker.
	UpdateStatus(
		ctx context.Context,
		eventID string,
		status WorkerStatus,
	) error
}

type workersService struct {
	authorize    authx.AuthorizeFn
	eventsStore  EventsStore
	workersStore WorkersStore
	substrate    Substrate
}

func NewWorkersService(
	eventsStore EventsStore,
	workersStore WorkersStore,
	substrate Substrate,
) WorkersService {
	return &workersService{
		authorize:    authx.Authorize,
		eventsStore:  eventsStore,
		workersStore: workersStore,
		substrate:    substrate,
	}
}

func (w *workersService) Start(ctx context.Context, eventID string) error {
	if err := w.authorize(ctx, authx.RoleScheduler()); err != nil {
		return err
	}

	event, err := w.eventsStore.Get(ctx, eventID)
	if err != nil {
		return errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}

	if event.Worker.Status.Phase != WorkerPhasePending {
		return &meta.ErrConflict{
			Type: "Event",
			ID:   event.ID,
			Reason: fmt.Sprintf(
				"Event %q worker has already been started.",
				event.ID,
			),
		}
	}

	if err = w.substrate.StartWorker(ctx, event); err != nil {
		return errors.Wrapf(err, "error starting worker for event %q", event.ID)
	}
	return nil
}

func (w *workersService) GetStatus(
	ctx context.Context,
	eventID string,
) (WorkerStatus, error) {
	if err := w.authorize(ctx, authx.RoleReader()); err != nil {
		return WorkerStatus{}, err
	}

	event, err := w.eventsStore.Get(ctx, eventID)
	if err != nil {
		return WorkerStatus{},
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	return event.Worker.Status, nil
}

// TODO: Should we put some kind of timeout on this function?
func (w *workersService) WatchStatus(
	ctx context.Context,
	eventID string,
) (<-chan WorkerStatus, error) {
	if err := w.authorize(ctx, authx.RoleReader()); err != nil {
		return nil, err
	}

	// Read the event up front to confirm it exists.
	if _, err := w.eventsStore.Get(ctx, eventID); err != nil {
		return nil,
			errors.Wrapf(err, "error retrieving event %q from store", eventID)
	}
	statusCh := make(chan WorkerStatus)
	go func() {
		defer close(statusCh)
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			event, err := w.eventsStore.Get(ctx, eventID)
			if err != nil {
				log.Printf("error retrieving event %q from store: %s", eventID, err)
				return
			}
			select {
			case statusCh <- event.Worker.Status:
			case <-ctx.Done():
				return
			}
		}
	}()
	return statusCh, nil
}

func (w *workersService) UpdateStatus(
	ctx context.Context,
	eventID string,
	status WorkerStatus,
) error {
	if err := w.authorize(ctx, authx.RoleObserver()); err != nil {
		return err
	}

	if err := w.workersStore.UpdateStatus(
		ctx,
		eventID,
		status,
	); err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker in store",
			eventID,
		)
	}
	return nil
}
