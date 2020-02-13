package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongodbTimeout = 5 * time.Second

type logStore struct {
	database *mongo.Database
}

func NewLogStore(database *mongo.Database) storage.LogStore {
	return &logStore{
		database: database,
	}
}

func (l *logStore) GetWorkerLogs(eventID string) ([]brignext.LogEntry, error) {
	collectionName := fmt.Sprintf("event.%s.brigade-runner", eventID)
	return l.getLogs(collectionName)
}

func (l *logStore) GetWorkerInitLogs(
	eventID string,
) ([]brignext.LogEntry, error) {
	collectionName := fmt.Sprintf("event.%s.vcs-sidecar", eventID)
	return l.getLogs(collectionName)
}

func (l *logStore) GetJobLogs(
	jobID string,
	containerName string,
) ([]brignext.LogEntry, error) {
	collectionName := fmt.Sprintf("job.%s.%s", jobID, containerName)
	return l.getLogs(collectionName)
}

func (l *logStore) StreamWorkerLogs(
	ctx context.Context,
	eventID string,
) (<-chan brignext.LogEntry, error) {
	collectionName := fmt.Sprintf("event.%s.brigade-runner", eventID)
	return l.streamLogs(ctx, collectionName)
}

func (l *logStore) StreamWorkerInitLogs(
	ctx context.Context,
	eventID string,
) (<-chan brignext.LogEntry, error) {
	collectionName := fmt.Sprintf("event.%s.vcs-sidecar", eventID)
	return l.streamLogs(ctx, collectionName)
}

func (l *logStore) StreamJobLogs(
	ctx context.Context,
	jobID string,
	containerName string,
) (<-chan brignext.LogEntry, error) {
	collectionName := fmt.Sprintf("job.%s.%s", jobID, containerName)
	return l.streamLogs(ctx, collectionName)
}

func (l *logStore) getLogs(collectionName string) ([]brignext.LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	collection := l.database.Collection(collectionName)
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving log entries")
	}
	logEntries := []brignext.LogEntry{}
	for cursor.Next(ctx) {
		logEntry := brignext.LogEntry{}
		err := cursor.Decode(&logEntry)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding log entries")
		}
		logEntries = append(logEntries, logEntry)
	}
	return logEntries, nil
}

func (l *logStore) streamLogs(
	ctx context.Context,
	collectionName string,
) (<-chan brignext.LogEntry, error) {
	collection := l.database.Collection(collectionName)
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
			cursor, err = collection.Find(
				ctx,
				&bson.D{},
				&options.FindOptions{CursorType: &cursorType},
			)
			if err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error getting cursor for collection %q",
						collectionName,
					),
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
			result := bson.M{}
			err = cursor.Decode(&result)
			if err != nil {
				log.Println(
					errors.Wrapf(
						err,
						"error decoding log entry from collection %q",
						collectionName,
					),
				)
				return
			}

			logEntry := brignext.LogEntry{
				Time:    result["time"].(primitive.DateTime).Time(),
				Message: result["log"].(string),
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
