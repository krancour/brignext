package mongodb

import (
	"context"
	"fmt"

	"github.com/brigadecore/brigade/v2/apiserver/internal/core"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type jobsStore struct {
	eventsCollection *mongo.Collection
}

func NewJobsStore(database *mongo.Database) (core.JobsStore, error) {
	return &jobsStore{
		eventsCollection: database.Collection("events"),
	}, nil
}

func (j *jobsStore) Create(
	ctx context.Context,
	eventID string,
	jobName string,
	job core.Job,
) error {
	res, err := j.eventsCollection.UpdateOne(
		ctx,
		bson.M{"id": eventID},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("worker.jobs.%s", jobName): job,
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

func (j *jobsStore) UpdateStatus(
	ctx context.Context,
	eventID string,
	jobName string,
	status core.JobStatus,
) error {
	res, err := j.eventsCollection.UpdateOne(
		ctx,
		bson.M{
			"id": eventID,
		},
		bson.M{
			"$set": bson.M{
				fmt.Sprintf("worker.jobs.%s.status", jobName): status,
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
