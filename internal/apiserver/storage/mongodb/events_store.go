package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/krancour/brignext/v2/internal/apiserver/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type eventsStore struct {
	collection *mongo.Collection
}

func NewEventsStore(database *mongo.Database) (storage.EventsStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("events")
	if _, err := collection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"metadata.id": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
			// This facilitates sorting by event creation date/time
			{
				Keys: bson.M{
					"metadata.created": -1,
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

func (e *eventsStore) Create(ctx context.Context, event brignext.Event) error {
	now := time.Now()
	event.Created = &now
	if _, err := e.collection.InsertOne(ctx, event); err != nil {
		return errors.Wrapf(err, "error inserting new event %q", event.ID)
	}
	return nil
}

func (e *eventsStore) List(ctx context.Context) (brignext.EventList, error) {
	eventList := brignext.NewEventList()
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.created": -1})
	cur, err := e.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return eventList, errors.Wrap(err, "error finding events")
	}
	if err := cur.All(ctx, &eventList.Items); err != nil {
		return eventList, errors.Wrap(err, "error decoding events")
	}
	return eventList, nil
}

func (e *eventsStore) ListByProject(
	ctx context.Context,
	projectID string,
) (brignext.EventList, error) {
	eventList := brignext.NewEventList()
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.created": -1})
	cur, err :=
		e.collection.Find(ctx, bson.M{"projectID": projectID}, findOptions)
	if err != nil {
		return eventList, errors.Wrapf(
			err,
			"error finding events for project %q",
			projectID,
		)
	}
	if err := cur.All(ctx, &eventList.Items); err != nil {
		return eventList, errors.Wrapf(
			err,
			"error decoding events for project %q",
			projectID,
		)
	}
	return eventList, nil
}

func (e *eventsStore) Get(
	ctx context.Context,
	id string,
) (brignext.Event, error) {
	event := brignext.Event{}
	res := e.collection.FindOne(ctx, bson.M{"metadata.id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return event, brignext.NewErrNotFound("Event", id)
	}
	if res.Err() != nil {
		return event, errors.Wrapf(res.Err(), "error finding event %q", id)
	}
	if err := res.Decode(&event); err != nil {
		return event, errors.Wrapf(err, "error decoding event %q", id)
	}
	return event, nil
}

func (e *eventsStore) Cancel(ctx context.Context, id string) error {
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{
			"metadata.id":               id,
			"status.workerStatus.phase": brignext.WorkerPhasePending,
		},
		bson.M{
			"$set": bson.M{
				"canceled":                  time.Now(),
				"status.workerStatus.phase": brignext.WorkerPhaseCanceled,
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
			"metadata.id":               id,
			"status.workerStatus.phase": brignext.WorkerPhaseRunning,
		},
		bson.M{
			"$set": bson.M{
				"canceled":                  time.Now(),
				"status.workerStatus.phase": brignext.WorkerPhaseAborted,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating status of event %q worker", id)
	}

	if res.MatchedCount == 0 {
		return brignext.NewErrConflict(
			"Event",
			id,
			fmt.Sprintf(
				"Event %q was not canceled because it was already in a terminal state.",
				id,
			),
		)
	}

	return nil
}

func (e *eventsStore) CancelCollection(
	ctx context.Context,
	opts brignext.EventListOptions,
) (brignext.EventReferenceList, error) {
	eventRefList := brignext.NewEventReferenceList()

	// For some reason, the MongoDB driver for Go doesn't implement the expose the
	// findAndModify() method. As a workaround, we'll do a straight updateMany()
	// and then we'll do a separate find() for all events that have the precise
	// moment of cancellation that we set in the updateMany() operation.

	var cancelPending bool
	var cancelRunning bool
	for _, workerPhase := range opts.WorkerPhases {
		if workerPhase == brignext.WorkerPhasePending {
			cancelPending = true
		}
		if workerPhase == brignext.WorkerPhaseRunning {
			cancelRunning = true
		}
	}

	if !cancelPending && !cancelRunning {
		return eventRefList, nil
	}

	cancellationTime := time.Now()

	if cancelPending {
		criteria := bson.M{
			"status.workerStatus.phase": brignext.WorkerPhasePending,
		}
		if opts.ProjectID != "" {
			criteria["projectID"] = opts.ProjectID
		}
		if _, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				"$set": bson.M{
					"canceled":                  cancellationTime,
					"status.workerStatus.phase": brignext.WorkerPhaseCanceled,
				},
			},
		); err != nil {
			return eventRefList, errors.Wrap(err, "error updating events")
		}
	}

	if cancelRunning {
		criteria := bson.M{
			"status.workerStatus.phase": brignext.WorkerPhaseRunning,
		}
		if opts.ProjectID != "" {
			criteria["projectID"] = opts.ProjectID
		}
		if _, err := e.collection.UpdateMany(
			ctx,
			criteria,
			bson.M{
				"$set": bson.M{
					"canceled":                  cancellationTime,
					"status.workerStatus.phase": brignext.WorkerPhaseAborted,
				},
			},
		); err != nil {
			return eventRefList, errors.Wrap(err, "error updating events")
		}
	}

	criteria := bson.M{
		"canceled": cancellationTime,
	}
	if opts.ProjectID != "" {
		criteria["projectID"] = opts.ProjectID
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.created": -1})
	cur, err := e.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return eventRefList, errors.Wrapf(err, "error finding canceled events")
	}
	if err := cur.All(ctx, &eventRefList.Items); err != nil {
		return eventRefList, errors.Wrap(err, "error decoding canceled events")
	}

	return eventRefList, nil
}

func (e *eventsStore) Delete(
	ctx context.Context,
	id string,
	deletePending bool,
	deleteRunning bool,
) (bool, error) {
	if _, err := e.Get(ctx, id); err != nil {
		return false, err
	}
	phasesToDelete := []brignext.WorkerPhase{
		brignext.WorkerPhaseCanceled,
		brignext.WorkerPhaseAborted,
		brignext.WorkerPhaseSucceeded,
		brignext.WorkerPhaseFailed,
		brignext.WorkerPhaseTimedOut,
	}
	if deletePending {
		phasesToDelete = append(phasesToDelete, brignext.WorkerPhasePending)
	}
	if deleteRunning {
		phasesToDelete = append(phasesToDelete, brignext.WorkerPhaseRunning)
	}
	res, err := e.collection.DeleteOne(
		ctx,
		bson.M{
			"metadata.id":               id,
			"status.workerStatus.phase": bson.M{"$in": phasesToDelete},
		},
	)
	if err != nil {
		return false, errors.Wrapf(err, "error deleting event %q", id)
	}
	return res.DeletedCount == 1, nil
}

func (e *eventsStore) DeleteByProject(
	ctx context.Context,
	projectID string,
) error {
	if _, err := e.collection.DeleteMany(
		ctx,
		bson.M{"projectID": projectID},
	); err != nil {
		return errors.Wrapf(err, "error deleting events for project %q", projectID)
	}
	return nil
}

func (e *eventsStore) UpdateWorkerStatus(
	ctx context.Context,
	eventID string,
	status brignext.WorkerStatus,
) error {
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{"metadata.id": eventID},
		bson.M{
			"$set": bson.M{
				"status.workerStatus": status,
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
		return brignext.NewErrNotFound("Event", eventID)
	}
	return nil
}

func (e *eventsStore) UpdateJobStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status brignext.JobStatus,
) error {
	res, err := e.collection.UpdateOne(
		ctx,
		bson.M{
			"metadata.id": eventID,
		},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("status.jobStatuses.%s", jobName): status,
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
		return brignext.NewErrNotFound("Event", eventID)
	}
	return nil
}
