package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type eventsStore struct {
	collection *mongo.Collection
}

func NewEventsStore(database *mongo.Database) (core.EventsStore, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("events")
	if _, err := collection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"id": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
			// This facilitates sorting by event creation date/time
			{
				Keys: bson.M{
					"created": -1,
				},
			},
			// This facilitates quickly selecting all events for a given project
			{
				Keys: bson.M{
					"projectID": 1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to events collection")
	}
	return &eventsStore{
		collection: collection,
	}, nil
}

func (e *eventsStore) Create(ctx context.Context, event core.Event) error {
	if _, err := e.collection.InsertOne(ctx, event); err != nil {
		return errors.Wrapf(err, "error inserting new event %q", event.ID)
	}
	return nil
}

func (e *eventsStore) List(
	ctx context.Context,
	selector core.EventsSelector,
	opts meta.ListOptions,
) (core.EventList, error) {
	events := core.EventList{}

	criteria := bson.M{
		"worker.status.phase": bson.M{
			"$in": selector.WorkerPhases,
		},
		"deleted": bson.M{
			"$exists": false, // Don't grab logically deleted events
		},
	}
	if selector.ProjectID != "" {
		criteria["projectID"] = selector.ProjectID
	}
	if opts.Continue != "" {
		continueTime, err :=
			time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", opts.Continue)
		if err != nil {
			return events, errors.Wrap(err, "error parsing continue time")
		}
		criteria["created"] = bson.M{"$lt": continueTime}
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created": -1})
	findOptions.SetLimit(opts.Limit)
	cur, err := e.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return events, errors.Wrap(err, "error finding events")
	}
	if err := cur.All(ctx, &events.Items); err != nil {
		return events, errors.Wrap(err, "error decoding events")
	}

	if int64(len(events.Items)) == opts.Limit {
		continueTime := events.Items[opts.Limit-1].Created
		criteria["created"] = bson.M{"$lt": continueTime}
		remaining, err := e.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return events, errors.Wrap(err, "error counting remaining events")
		}
		if remaining > 0 {
			events.Continue = continueTime.String()
			events.RemainingItemCount = remaining
		}
	}

	return events, nil
}

func (e *eventsStore) Get(
	ctx context.Context,
	id string,
) (core.Event, error) {
	event := core.Event{}
	res := e.collection.FindOne(ctx, bson.M{"id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return event, &meta.ErrNotFound{
			Type: "Event",
			ID:   id,
		}
	}
	if res.Err() != nil {
		return event, errors.Wrapf(res.Err(), "error finding event %q", id)
	}
	if err := res.Decode(&event); err != nil {
		return event, errors.Wrapf(err, "error decoding event %q", id)
	}
	return event, nil
}

func (e *eventsStore) GetByHashedWorkerToken(
	ctx context.Context,
	hashedWorkerToken string,
) (core.Event, error) {
	event := core.Event{}
	res := e.collection.FindOne(
		ctx,
		bson.M{
			"worker.hashedToken": hashedWorkerToken,
		},
	)
	if res.Err() == mongo.ErrNoDocuments {
		return event, &meta.ErrNotFound{
			Type: "Event",
		}
	}
	if res.Err() != nil {
		return event, errors.Wrap(res.Err(), "error finding event")
	}
	if err := res.Decode(&event); err != nil {
		return event, errors.Wrap(err, "error decoding event")
	}
	return event, nil
}

func (e *eventsStore) Cancel(ctx context.Context, id string) error {
	if _, err := e.Get(ctx, id); err != nil {
		return err
	}
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{
			"id":                  id,
			"worker.status.phase": core.WorkerPhasePending,
		},
		bson.M{
			"$set": bson.M{
				"canceled":            time.Now(),
				"worker.status.phase": core.WorkerPhaseCanceled,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating status of event %q worker", id)
	}
	if res.MatchedCount == 1 {
		return nil
	}

	res, err = e.collection.UpdateOne(
		ctx,
		bson.M{
			"id":                  id,
			"worker.status.phase": core.WorkerPhaseRunning,
		},
		bson.M{
			"$set": bson.M{
				"canceled":            time.Now(),
				"worker.status.phase": core.WorkerPhaseAborted,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating status of event %q worker", id)
	}

	if res.MatchedCount == 0 {
		return &meta.ErrConflict{
			Type: "Event",
			ID:   id,
			Reason: fmt.Sprintf(
				"Event %q was not canceled because it was already in a terminal state.",
				id,
			),
		}
	}

	return nil
}

func (e *eventsStore) CancelMany(
	ctx context.Context,
	selector core.EventsSelector,
) (core.EventList, error) {
	events := core.EventList{}
	// It only makes sense to cancel events that are in a pending or running
	// state. We can ignore anything else.
	var cancelPending bool
	var cancelRunning bool
	for _, workerPhase := range selector.WorkerPhases {
		if workerPhase == core.WorkerPhasePending {
			cancelPending = true
		}
		if workerPhase == core.WorkerPhaseRunning {
			cancelRunning = true
		}
	}

	// Bail if we're not canceling pending or running events
	if !cancelPending && !cancelRunning {
		return events, nil
	}

	// The MongoDB driver for Go doesn't expose findAndModify(), which could be
	// used to select events and cancel them at the same time. As a workaround,
	// we'll cancel first, then select events based on cancellation time.

	cancellationTime := time.Now()

	criteria := bson.M{
		"projectID": selector.ProjectID,
	}

	if cancelPending {
		criteria["worker.status.phase"] = core.WorkerPhasePending
		if _, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				"$set": bson.M{
					"canceled":            cancellationTime,
					"worker.status.phase": core.WorkerPhaseCanceled,
				},
			},
		); err != nil {
			return events, errors.Wrap(err, "error updating events")
		}
	}

	if cancelRunning {
		criteria["worker.status.phase"] = core.WorkerPhaseRunning
		if _, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				"$set": bson.M{
					"canceled":            cancellationTime,
					"worker.status.phase": core.WorkerPhaseAborted,
				},
			},
		); err != nil {
			return events, errors.Wrap(err, "error updating events")
		}
	}

	delete(criteria, "worker.status.phase")
	criteria["canceled"] = cancellationTime
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created": -1})
	cur, err := e.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return events, errors.Wrapf(err, "error finding canceled events")
	}
	if err := cur.All(ctx, &events.Items); err != nil {
		return events, errors.Wrap(err, "error decoding canceled events")
	}

	return events, nil
}

func (e *eventsStore) Delete(ctx context.Context, id string) error {
	res, err := e.collection.DeleteOne(
		ctx,
		bson.M{
			"id": id,
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error deleting event %q", id)
	}
	if res.DeletedCount != 1 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   id,
		}
	}
	return nil
}

func (e *eventsStore) DeleteMany(
	ctx context.Context,
	selector core.EventsSelector,
) (core.EventList, error) {
	events := core.EventList{}

	// The MongoDB driver for Go doesn't expose findAndModify(), which could be
	// used to select events and delete them at the same time. As a workaround,
	// we'll perform a logical delete first, select the logically deleted events,
	// and then perform a real delete.

	deletedTime := time.Now()

	// Logical delete...
	criteria := bson.M{
		"projectID": selector.ProjectID,
		"worker.status.phase": bson.M{
			"$in": selector.WorkerPhases,
		},
		"deleted": bson.M{
			"$exists": false,
		},
	}
	if _, err := e.collection.UpdateMany(
		ctx,
		criteria,
		bson.M{
			"$set": bson.M{
				"deleted": deletedTime,
			},
		},
	); err != nil {
		return events, errors.Wrap(err, "error logically deleting events")
	}

	// Select the logically deleted documents...
	criteria["deleted"] = deletedTime
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created": -1})
	cur, err := e.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return events, errors.Wrapf(
			err,
			"error finding logically deleted events",
		)
	}
	if err := cur.All(ctx, &events.Items); err != nil {
		return events, errors.Wrap(
			err,
			"error decoding logically deleted events",
		)
	}

	// Final deletion
	if _, err := e.collection.DeleteMany(ctx, criteria); err != nil {
		return events, errors.Wrap(err, "error deleting events")
	}

	return events, nil
}

func (e *eventsStore) UpdateWorkerSpec(
	ctx context.Context,
	eventID string,
	spec core.WorkerSpec,
) error {
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{"id": eventID},
		bson.M{
			"$set": bson.M{
				"worker.spec": spec,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating spec of event %q worker",
			eventID,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   eventID,
		}
	}
	return nil
}

func (e *eventsStore) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status core.WorkerStatus,
) error {
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{"id": eventID},
		bson.M{
			"$set": bson.M{
				"worker.status": status,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker",
			eventID,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   eventID,
		}
	}
	return nil
}

func (e *eventsStore) CreateJob(
	ctx context.Context,
	eventID string,
	jobName string,
	jobSpec core.JobSpec,
) error {
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{"id": eventID},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("worker.jobs.%e.spec", jobName): jobSpec,
				fmt.Sprintf("worker.jobs.%e.status", jobName): core.JobStatus{
					Phase: core.JobPhasePending,
				},
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating spec of event %q job %q",
			eventID,
			jobName,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   eventID,
		}
	}
	return nil
}

func (e *eventsStore) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status core.JobStatus,
) error {
	job := core.Job{
		Status: status,
	}
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{
			"id": eventID,
		},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("worker.jobs.%s", jobName): job,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(
			err,
			"error updating status of event %q worker job %q",
			eventID,
			jobName,
		)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Event",
			ID:   eventID,
		}
	}
	return nil
}
