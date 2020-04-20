package mongodb

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/krancour/brignext"
	"github.com/krancour/brignext/apiserver/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var mongodbTimeout = 5 * time.Second

type logStore struct {
	logsCollection *mongo.Collection
}

func NewLogStore(database *mongo.Database) storage.LogStore {
	return &logStore{
		logsCollection: database.Collection("logs"),
	}
}

func (l *logStore) GetWorkerLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) ([]brignext.LogEntry, error) {
	return l.getLogs(
		ctx,
		bson.M{
			"component": "worker",
			"event":     eventID,
			"worker":    workerName,
			"container": strings.ToLower(workerName),
		},
	)
}

func (l *logStore) StreamWorkerLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) (<-chan brignext.LogEntry, error) {
	return l.streamLogs(
		ctx,
		bson.M{
			"component": "worker",
			"event":     eventID,
			"worker":    workerName,
			"container": strings.ToLower(workerName),
		},
	)
}

func (l *logStore) GetWorkerInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) ([]brignext.LogEntry, error) {
	return l.getLogs(
		ctx,
		bson.M{
			"component": "worker",
			"event":     eventID,
			"worker":    workerName,
			"container": "vcs",
		},
	)
}

func (l *logStore) StreamWorkerInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
) (<-chan brignext.LogEntry, error) {
	return l.streamLogs(
		ctx,
		bson.M{
			"component": "worker",
			"event":     eventID,
			"worker":    workerName,
			"container": "vcs",
		},
	)
}

func (l *logStore) GetJobLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) ([]brignext.LogEntry, error) {
	return l.getLogs(
		ctx,
		bson.M{
			"component": "job",
			"event":     eventID,
			"worker":    workerName,
			"job":       jobName,
			"container": strings.ToLower(jobName),
		},
	)
}

func (l *logStore) StreamJobLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	return l.streamLogs(
		ctx,
		bson.M{
			"component": "job",
			"event":     eventID,
			"worker":    workerName,
			"job":       jobName,
			"container": strings.ToLower(jobName),
		},
	)
}

func (l *logStore) GetJobInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) ([]brignext.LogEntry, error) {
	return l.getLogs(
		ctx,
		bson.M{
			"component": "job",
			"event":     eventID,
			"worker":    workerName,
			"job":       jobName,
			"container": "vcs",
		},
	)
}

func (l *logStore) StreamJobInitLogs(
	ctx context.Context,
	eventID string,
	workerName string,
	jobName string,
) (<-chan brignext.LogEntry, error) {
	return l.streamLogs(
		ctx,
		bson.M{
			"component": "job",
			"event":     eventID,
			"worker":    workerName,
			"job":       jobName,
			"container": "vcs",
		},
	)
}

func (l *logStore) getLogs(
	ctx context.Context,
	criteria bson.M,
) ([]brignext.LogEntry, error) {
	cursor, err := l.logsCollection.Find(ctx, criteria)
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
	criteria bson.M,
) (<-chan brignext.LogEntry, error) {
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
			logEntry := brignext.LogEntry{}
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
