package mongodb

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/events"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type logsStore struct {
	collection *mongo.Collection
}

func NewLogsStore(database *mongo.Database) events.LogsStore {
	return &logsStore{
		collection: database.Collection("logs"),
	}
}

func (l *logsStore) GetLogs(
	ctx context.Context,
	event brignext.Event,
	selector brignext.LogsSelector,
	opts meta.ListOptions,
) (brignext.LogEntryList, error) {
	logEntries := brignext.LogEntryList{
		Items: []brignext.LogEntry{},
	}

	criteria := l.criteriaFromSelector(event.ID, selector)
	if opts.Continue != "" {
		continueTime, err := time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", opts.Continue)
		if err != nil {
			return logEntries, errors.Wrap(err, "error parsing continue time")
		}
		criteria["time"] = bson.M{"$gt": continueTime}
	}

	findOptions := options.Find()
	// TODO: We might need this if we can't use capped collections in some environments
	// findOptions.SetSort(bson.M{"created": -1})
	findOptions.SetLimit(opts.Limit)
	cur, err := l.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return logEntries, errors.Wrap(err, "error retrieving log entries")
	}
	// TODO: Why aren't we using cur.All() here?
	for cur.Next(ctx) {
		logEntry := brignext.LogEntry{}
		err := cur.Decode(&logEntry)
		if err != nil {
			return logEntries, errors.Wrap(err, "error decoding log entries")
		}
		logEntries.Items = append(logEntries.Items, logEntry)
	}

	if int64(len(logEntries.Items)) == opts.Limit {
		continueTime := logEntries.Items[opts.Limit-1].Time
		criteria["time"] = bson.M{"$gt": continueTime}
		remaining, err := l.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return logEntries, errors.Wrap(err, "error counting remaining log entries")
		}
		if remaining > 0 {
			logEntries.Continue = continueTime.String()
			logEntries.RemainingItemCount = remaining
		}
	}

	return logEntries, nil
}

func (l *logsStore) StreamLogs(
	ctx context.Context,
	event brignext.Event,
	selector brignext.LogsSelector,
) (<-chan brignext.LogEntry, error) {
	criteria := l.criteriaFromSelector(event.ID, selector)

	logEntryCh := make(chan brignext.LogEntry)
	go func() {
		defer close(logEntryCh)

		cursorType := options.Tailable
		var cur *mongo.Cursor
		var err error
		// Any attempt to open a cursor that initially retrieves nothing will yield
		// a "dead" cursor which is no good for tailing the capped collection. We
		// need to retry this until we get a "live" cursor or the context is
		// canceled.
		for {
			cur, err = l.collection.Find(
				ctx,
				criteria,
				&options.FindOptions{CursorType: &cursorType},
			)
			if err != nil {
				log.Println(
					errors.Wrap(err, "error getting cursor for logs collection"),
				)
				return
			}
			if cur.ID() != 0 {
				// We got a live cursor.
				break
			}
			select {
			case <-time.After(time.Second): // Wait before retry
			case <-ctx.Done():
				return
			}
		}

		var available bool
		for {
			available = cur.TryNext(ctx)
			if !available {
				select {
				case <-time.After(time.Second): // Wait before retry
					continue
				case <-ctx.Done():
					return
				}
			}
			logEntry := brignext.LogEntry{}
			err = cur.Decode(&logEntry)
			if err != nil {
				log.Println(
					errors.Wrapf(err, "error decoding log entry from collection"),
				)
				return
			}

			select {
			case logEntryCh <- logEntry:
			case <-ctx.Done():
				return
			}
		}
	}()

	return logEntryCh, nil
}

func (l *logsStore) criteriaFromSelector(
	eventID string,
	selector brignext.LogsSelector,
) bson.M {
	criteria := bson.M{
		"event": eventID,
	}

	// If no job was specified, we want worker logs
	if selector.Job == "" {
		criteria["component"] = "worker"
		// If no container was specified, we want the "worker" container
		if selector.Container == "" {
			criteria["container"] = "worker"
		} else { // We want the one specified
			criteria["container"] = selector.Container
		}
	} else { // We want job logs
		criteria["component"] = "job"
		// TODO: Probably we shouldn't let users set the job's primary container's
		// name or else this assumption below doesn't hold.
		//
		// If no container was specified, we want the one with the same name as the
		// job
		if selector.Container == "" {
			criteria["container"] = selector.Job
		} else { // We want the one specified
			criteria["container"] = selector.Container
		}
	}

	return criteria
}
