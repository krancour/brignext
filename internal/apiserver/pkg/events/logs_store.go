package events

import (
	"context"
	"log"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type LogsStore interface {
	GetLogs(
		ctx context.Context,
		eventID string,
		opts brignext.LogOptions,
	) (brignext.LogEntryList, error)
	StreamLogs(
		ctx context.Context,
		eventID string,
		opts brignext.LogOptions,
	) (<-chan brignext.LogEntry, error)

	CheckHealth(context.Context) error
}

type logsStore struct {
	logsCollection *mongo.Collection
}

func NewLogsStore(database *mongo.Database) LogsStore {
	return &logsStore{
		logsCollection: database.Collection("logs"),
	}
}

func (l *logsStore) GetLogs(
	ctx context.Context,
	eventID string,
	opts brignext.LogOptions,
) (brignext.LogEntryList, error) {
	criteria := l.criteriaFromOptions(eventID, opts)

	logEntryList := brignext.NewLogEntryList()
	cursor, err := l.logsCollection.Find(ctx, criteria)
	if err != nil {
		return logEntryList, errors.Wrap(err, "error retrieving log entries")
	}
	for cursor.Next(ctx) {
		logEntry := brignext.NewLogEntry()
		err := cursor.Decode(&logEntry)
		if err != nil {
			return logEntryList, errors.Wrap(err, "error decoding log entries")
		}
		logEntryList.Items = append(logEntryList.Items, logEntry)
	}
	return logEntryList, nil
}

func (l *logsStore) StreamLogs(
	ctx context.Context,
	eventID string,
	opts brignext.LogOptions,
) (<-chan brignext.LogEntry, error) {
	criteria := l.criteriaFromOptions(eventID, opts)

	logEntryCh := make(chan brignext.LogEntry)
	go func() {
		defer close(logEntryCh)

		cursorType := options.Tailable
		var cursor *mongo.Cursor
		var err error
		// Any attempt to open a cursor that initially retrieves nothing will yield
		// a "dead" cursor which is no good for tailing the capped collection. We
		// need to retry this until we get a "live" cursor or the context is
		// canceled.
		for {
			cursor, err = l.logsCollection.Find(
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
			if cursor.ID() != 0 {
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
			available = cursor.TryNext(ctx)
			if !available {
				select {
				case <-time.After(time.Second): // Wait before retry
					continue
				case <-ctx.Done():
					return
				}
			}
			logEntry := brignext.NewLogEntry()
			err = cursor.Decode(&logEntry)
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

func (l *logsStore) criteriaFromOptions(
	eventID string,
	opts brignext.LogOptions,
) bson.M {
	criteria := bson.M{
		"event": eventID,
	}

	// If no job was specified, we want worker logs
	if opts.Job == "" {
		criteria["component"] = "worker"
		// If no container was specified, we want the "worker" container
		if opts.Container == "" {
			criteria["container"] = "worker"
		} else { // We want the one specified
			criteria["container"] = opts.Container
		}
	} else { // We want job logs
		criteria["component"] = "job"
		// TODO: Probably we shouldn't let users set the job's primary container's
		// name or else this assumption below doesn't hold.
		//
		// If no container was specified, we want the one with the same name as the
		// job
		if opts.Container == "" {
			criteria["container"] = opts.Job
		} else { // We want the one specified
			criteria["container"] = opts.Container
		}
	}

	return criteria
}

func (l *logsStore) CheckHealth(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := l.logsCollection.Database().Client().Ping(
		pingCtx,
		readpref.Primary(),
	); err != nil {
		return errors.Wrap(err, "error pinging mongo")
	}
	return nil
}
