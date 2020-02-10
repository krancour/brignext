package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type projectStore struct {
	projectsCollection *mongo.Collection
	eventsCollection   *mongo.Collection
	// jobsCollection     *mongo.Collection
}

func NewProjectStore(database *mongo.Database) (storage.ProjectStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	eventsCollection := database.Collection("events")
	if _, err := eventsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"projectID": 1,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to events collection")
	}

	// jobsCollection := database.Collection("jobs")
	// if _, err := jobsCollection.Indexes().CreateOne(
	// 	ctx,
	// 	mongo.IndexModel{
	// 		Keys: bson.M{
	// 			"eventID": 1,
	// 		},
	// 	},
	// ); err != nil {
	// 	return nil, errors.Wrap(err, "error adding indexes to jobs collection")
	// }

	return &projectStore{
		projectsCollection: database.Collection("projects"),
		eventsCollection:   eventsCollection,
		// jobsCollection:     jobsCollection,
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

	cur, err := p.projectsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving projects")
	}
	projects := []brignext.Project{}
	for cur.Next(ctx) {
		project := brignext.Project{}
		err := cur.Decode(&project)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding projects")
		}
		projects = append(projects, project)
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

	if _, err :=
		p.projectsCollection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return errors.Wrapf(err, "error deleting project %q", id)
	}
	return nil
}

func (p *projectStore) CreateEvent(event brignext.Event) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	event.ID = uuid.NewV4().String()
	now := time.Now()
	event.Created = &now
	event.Status = brignext.EventStatusCreated

	if _, err := p.eventsCollection.InsertOne(ctx, event); err != nil {
		return "", errors.Wrapf(err, "error creating event %q", event.ID)
	}
	return event.ID, nil
}

func (p *projectStore) GetEvents() ([]brignext.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.eventsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving events")
	}

	events := []brignext.Event{}
	for cur.Next(ctx) {
		event := brignext.Event{}
		err := cur.Decode(&event)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding events")
		}
		events = append(events, event)
	}

	return events, nil
}

func (p *projectStore) GetEventsByProjectID(
	projectID string,
) ([]brignext.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.eventsCollection.Find(ctx, bson.M{"projectID": projectID})
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving events for project %q",
			projectID,
		)
	}
	events := []brignext.Event{}
	for cur.Next(ctx) {
		event := brignext.Event{}
		err := cur.Decode(&event)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"error decoding events for project %q",
				projectID,
			)
		}
		events = append(events, event)
	}
	return events, nil
}

func (p *projectStore) GetEvent(id string) (brignext.Event, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	event := brignext.Event{}

	result := p.eventsCollection.FindOne(ctx, bson.M{"_id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return event, false, nil
	}
	if result.Err() != nil {
		return event, false, errors.Wrapf(
			result.Err(),
			"error retrieving event %q",
			id,
		)
	}
	if err := result.Decode(&event); err != nil {
		return event, false, errors.Wrapf(err, "error decoding event %q", id)
	}

	return event, true, nil
}

func (p *projectStore) DeleteEventsByProjectID(
	projectID string,
	options storage.DeleteEventOptions,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	criteria := bson.M{"projectID": projectID}
	if !options.DeleteEventsWithRunningWorkers {
		// TODO: Amend the criteria appropriately
	}
	if _, err := p.eventsCollection.DeleteMany(ctx, criteria); err != nil {
		return errors.Wrapf(
			err,
			"error deleting events for project %q",
			projectID,
		)
	}

	return nil
}

func (p *projectStore) DeleteEvent(
	id string,
	options storage.DeleteEventOptions,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	criteria := bson.M{"_id": id}
	if !options.DeleteEventsWithRunningWorkers {
		// TODO: Amend the criteria appropriately
	}
	if _, err :=
		p.eventsCollection.DeleteOne(ctx, criteria); err != nil {
		return errors.Wrapf(err, "error deleting event %q", id)
	}

	return nil
}

// func (p *projectStore) CreateJob(job brignext.Job) (string, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
// 	defer cancel()

// 	job.ID = uuid.NewV4().String()

// 	if _, err := p.jobsCollection.InsertOne(ctx, job); err != nil {
// 		return "", errors.Wrapf(err, "error creating job %q", job.ID)
// 	}
// 	return job.ID, nil
// }

// func (p *projectStore) GetJobsByEventID(
// 	eventID string,
// ) ([]brignext.Job, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
// 	defer cancel()

// 	cur, err := p.jobsCollection.Find(ctx, bson.M{"eventID": eventID})
// 	if err != nil {
// 		return nil, errors.Wrapf(err, "error retrieving jobs for event %q", eventID)
// 	}
// 	jobs := []brignext.Job{}
// 	for cur.Next(ctx) {
// 		job := brignext.Job{}
// 		err := cur.Decode(&job)
// 		if err != nil {
// 			return nil, errors.Wrapf(err, "error decoding job for event %q", eventID)
// 		}
// 		jobs = append(jobs, job)
// 	}

// 	return jobs, nil
// }

// func (p *projectStore) UpdateJobStatus(jobID string, status string) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
// 	defer cancel()

// 	if _, err :=
// 		p.jobsCollection.UpdateOne(
// 			ctx,
// 			bson.M{
// 				"_id": jobID,
// 			},
// 			bson.M{
// 				"$set": bson.M{"status": status},
// 			},
// 		); err != nil {
// 		return errors.Wrapf(err, "error updating status for job %q", jobID)
// 	}
// 	return nil
// }

// func (p *projectStore) GetJob(id string) (brignext.Job, bool, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
// 	defer cancel()

// 	job := brignext.Job{}

// 	result := p.jobsCollection.FindOne(ctx, bson.M{"_id": id})
// 	if result.Err() == mongo.ErrNoDocuments {
// 		return job, false, nil
// 	}
// 	if result.Err() != nil {
// 		return job, false, errors.Wrapf(result.Err(), "error retrieving job %q", id)
// 	}

// 	if err := result.Decode(&job); err != nil {
// 		return job, false, errors.Wrapf(err, "error decoding job %q", id)
// 	}
// 	return job, true, nil
// }

// func (p *projectStore) DeleteJobsByEventID(eventID string) error {
// 	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
// 	defer cancel()

// 	if _, err :=
// 		p.eventsCollection.DeleteMany(ctx, bson.M{"everntID": eventID}); err != nil {
// 		return errors.Wrapf(err, "error deleting jobs for event %q", eventID)
// 	}

// 	return nil
// }
