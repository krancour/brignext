package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/logic"
	"github.com/krancour/brignext/pkg/mongodb"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type projectStore struct {
	database           *mongo.Database
	projectsCollection *mongo.Collection
	eventsCollection   *mongo.Collection
	workersCollection  *mongo.Collection
}

func NewProjectStore(database *mongo.Database) (storage.ProjectStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	eventsCollection := database.Collection("events")
	if _, err := eventsCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"projectID": 1,
				},
			},
			{
				Keys: bson.M{
					"created": -1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to events collection")
	}

	workersCollection := database.Collection("workers")
	if _, err := workersCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"eventID": 1,
				},
			},
			{
				Keys: bson.M{
					"projectID": 1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to workers collection")
	}

	return &projectStore{
		database:           database,
		projectsCollection: database.Collection("projects"),
		eventsCollection:   eventsCollection,
		workersCollection:  workersCollection,
	}, nil
}

func (p *projectStore) CreateProject(project brignext.Project) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	now := time.Now()
	project.Created = &now

	if _, err := p.projectsCollection.InsertOne(ctx, project); err != nil {
		return "", errors.Wrapf(err, "error creating project %q", project.ID)
	}

	return project.ID, nil
}

func (p *projectStore) GetProjects() ([]brignext.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
	cur, err := p.projectsCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving projects")
	}

	projects := []brignext.Project{}
	if err := cur.All(ctx, &projects); err != nil {
		return nil, errors.Wrap(err, "error decoding projects")
	}

	return projects, nil
}

func (p *projectStore) GetProject(id string) (brignext.Project, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	project := brignext.Project{}

	result := p.projectsCollection.FindOne(ctx, bson.M{"_id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return project, false, nil
	}
	if result.Err() != nil {
		return project, false, errors.Wrapf(
			result.Err(),
			"error retrieving project %q",
			id,
		)
	}

	if err := result.Decode(&project); err != nil {
		return project, false, errors.Wrapf(err, "error decoding project %q", id)
	}

	return project, true, nil
}

func (p *projectStore) UpdateProject(project brignext.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.projectsCollection.ReplaceOne(
			ctx,
			bson.M{
				"_id": project.ID,
			},
			project,
		); err != nil {
		return errors.Wrapf(err, "error updating project %q", project.ID)
	}

	return nil
}

func (p *projectStore) DeleteProject(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	return mongodb.DoTx(ctx, p.database,
		func(sc mongo.SessionContext) error {

			if _, err :=
				p.projectsCollection.DeleteOne(sc, bson.M{"_id": id}); err != nil {
				return errors.Wrapf(err, "error deleting project %q", id)
			}

			if _, err :=
				p.eventsCollection.DeleteMany(sc, bson.M{"projectID": id}); err != nil {
				return errors.Wrapf(err, "error deleting events for project %q", id)
			}

			if _, err :=
				p.workersCollection.DeleteMany(
					sc,
					bson.M{"projectID": id},
				); err != nil {
				return errors.Wrapf(err, "error deleting workers for project %q", id)
			}

			return nil
		},
	)
}

func (p *projectStore) CreateEvent(event brignext.Event) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	event.ID = uuid.NewV4().String()
	workers := make([]interface{}, len(event.Workers))
	if len(event.Workers) == 0 {
		event.Status = brignext.EventStatusMoot
	} else {
		event.Status = brignext.EventStatusAccepted
		for i, worker := range event.Workers {
			worker.ID = uuid.NewV4().String()
			worker.EventID = event.ID
			worker.ProjectID = event.ProjectID
			worker.Status = brignext.WorkerStatusPending
			workers[i] = worker
		}
	}
	now := time.Now()
	event.Created = &now
	event.Workers = nil

	return event.ID, mongodb.DoTx(ctx, p.database,
		func(sc mongo.SessionContext) error {

			if _, err := p.eventsCollection.InsertOne(sc, event); err != nil {
				return errors.Wrapf(err, "error creating event %q", event.ID)
			}

			if len(workers) > 0 {
				if _, err := p.workersCollection.InsertMany(sc, workers); err != nil {
					return errors.Wrapf(
						err,
						"error creating workers for event %q",
						event.ID,
					)
				}
			}

			return nil
		},
	)
}

func (p *projectStore) GetEvents(
	criteria storage.GetEventsCriteria,
) ([]brignext.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	bsonCriteria := bson.M{}
	if criteria.ProjectID != "" {
		bsonCriteria["projectID"] = criteria.ProjectID
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"created": -1})
	cur, err := p.eventsCollection.Find(ctx, bsonCriteria, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving events")
	}

	events := []brignext.Event{}
	if err := cur.All(ctx, &events); err != nil {
		return nil, errors.Wrap(err, "error decoding events")
	}

	return events, nil
}

func (p *projectStore) GetEvent(id string) (brignext.Event, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	event := brignext.Event{}

	events := []brignext.Event{}
	cur, err := p.eventsCollection.Aggregate(
		ctx,
		[]bson.M{
			{
				"$match": bson.M{"_id": id},
			},
			{
				"$lookup": bson.M{ // Left outer join
					"from":         "workers",
					"localField":   "_id",
					"foreignField": "eventID",
					"as":           "workers",
				},
			},
		},
	)
	if err != nil {
		return event, false, errors.Wrapf(
			err,
			"error retrieving event %q",
			id,
		)
	}
	if err := cur.All(ctx, &events); err != nil {
		return event, false, errors.Wrapf(err, "error decoding event %q", id)
	}

	if len(events) == 0 {
		return event, false, nil
	}

	return events[0], true, nil
}

func (p *projectStore) DeleteEvents(
	criteria storage.DeleteEventsCriteria,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if !logic.ExactlyOne(
		criteria.EventID != "",
		criteria.ProjectID != "",
	) {
		return errors.New(
			"invalid criteria: only ONE of event ID or project ID must be specified",
		)
	}

	bsonCriteria := bson.M{}
	if criteria.EventID != "" {
		bsonCriteria["_id"] = criteria.EventID
	} else if criteria.ProjectID != "" {
		bsonCriteria["projectID"] = criteria.ProjectID
	}
	statusesToDelete := []brignext.EventStatus{
		brignext.EventStatusMoot,
		brignext.EventStatusCanceled,
		brignext.EventStatusAborted,
		brignext.EventStatusSucceeded,
		brignext.EventStatusFailed,
	}
	if criteria.DeleteAcceptedEvents {
		statusesToDelete = append(statusesToDelete, brignext.EventStatusAccepted)
	}
	if criteria.DeleteProcessingEvents {
		statusesToDelete = append(statusesToDelete, brignext.EventStatusProcessing)
	}
	bsonCriteria["status"] = bson.M{"$in": statusesToDelete}

	if _, err := p.eventsCollection.DeleteMany(ctx, bsonCriteria); err != nil {
		return errors.Wrap(err, "error deleting events")
	}

	// TODO: Cascade the delete to orphaned workers. Not yet sure how.

	return nil
}
