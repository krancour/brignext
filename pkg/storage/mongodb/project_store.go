package mongodb

import (
	"context"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type projectStore struct {
	projectsCollection *mongo.Collection
	eventsCollection   *mongo.Collection
	jobsCollection     *mongo.Collection
}

func NewProjectStore(database *mongo.Database) (storage.ProjectStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	unique := true

	projectsCollection := database.Collection("projects")
	if _, err := projectsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"name": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding index to projects collection")
	}

	eventsCollection := database.Collection("events")
	if _, err := eventsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"projectName": 1,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to events collection")
	}

	jobsCollection := database.Collection("jobs")
	if _, err := jobsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"eventID": 1,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to jobs collection")
	}

	return &projectStore{
		projectsCollection: projectsCollection,
		eventsCollection:   eventsCollection,
		jobsCollection:     jobsCollection,
	}, nil
}

func (p *projectStore) CreateProject(project brignext.Project) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	project.ID = uuid.NewV4().String()

	if _, err := p.projectsCollection.InsertOne(ctx, project); err != nil {
		return "", errors.Wrapf(err, "error creating project %q", project.Name)
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

func (p *projectStore) GetProjectByName(
	name string,
) (brignext.Project, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	project := brignext.Project{}

	result := p.projectsCollection.FindOne(ctx, bson.M{"name": name})
	if result.Err() == mongo.ErrNoDocuments {
		return project, false, nil
	}
	if result.Err() != nil {
		return project, false, errors.Wrapf(
			result.Err(),
			"error retrieving project %q",
			name,
		)
	}
	if err := result.Decode(&project); err != nil {
		return project, false, errors.Wrapf(err, "error decoding project %q", name)
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
				"name": project.Name,
			},
			project,
		); err != nil {
		return errors.Wrapf(err, "error updating project %q", project.Name)
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

func (p *projectStore) DeleteProjectByName(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.projectsCollection.DeleteOne(ctx, bson.M{"name": name}); err != nil {
		return errors.Wrapf(err, "error deleting project %q", name)
	}
	return nil
}

func (p *projectStore) CreateEvent(event brignext.Event) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	event.ID = uuid.NewV4().String()

	// Supplement the event with the project ID before persisting. In the
	// underlying datastore, events are related to projects by project ID, but
	// externally, project names are used as identifiers.
	if project, ok, err := p.GetProjectByName(event.ProjectName); err != nil {
		return "", errors.Wrapf(
			err,
			"error retrieving project %q",
			event.ProjectName,
		)
	} else if !ok {
		return "", errors.Errorf("project %q not found", event.ProjectName)
	} else {
		event.ProjectID = project.ID
	}

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
	projectNamesByID := map[string]string{}
	for cur.Next(ctx) {
		event := brignext.Event{}
		err := cur.Decode(&event)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding events")
		}
		// Supplement the event with the project name before returning. In the
		// underlying datastore, events are related to projects by project ID, but
		// externally, project names are used as identifiers.
		if projectName, ok := projectNamesByID[event.ProjectID]; ok {
			event.ProjectName = projectName
		} else {
			if project, ok, err := p.GetProject(event.ProjectID); err != nil {
				return nil, errors.Wrapf(
					err,
					"error retrieving project %q",
					event.ProjectID,
				)
			} else if ok {
				projectNamesByID[event.ProjectID] = project.Name
				event.ProjectName = project.Name
			}
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

	var projectName string
	if project, ok, err := p.GetProject(projectID); err != nil {
		return nil, errors.Wrapf(err, "error retrieving project %q", projectID)
	} else if ok {
		projectName = project.Name
	}

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
		// Supplement the event with the project name before returning. In the
		// underlying datastore, events are related to projects by project ID, but
		// externally, project names are used as identifiers.
		event.ProjectName = projectName
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

	// Supplement the event with the project name before returning. In the
	// underlying datastore, events are related to projects by project ID, but
	// externally, project names are used as identifiers.
	if project, ok, err := p.GetProject(event.ProjectID); err != nil {
		return event, false, errors.Wrapf(
			err,
			"error retrieving project %q",
			event.ProjectID,
		)
	} else if ok {
		event.ProjectName = project.Name
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

func (p *projectStore) UpdateWorker(
	eventID string,
	worker brignext.Worker,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.eventsCollection.UpdateOne(
			ctx,
			bson.M{
				"_id": eventID,
			},
			bson.M{
				"$set": bson.M{
					"worker": worker,
				},
			},
		); err != nil {
		return errors.Wrapf(err, "error updating worker for event %q", eventID)
	}

	return nil
}

func (p *projectStore) CreateJob(job brignext.Job) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	job.ID = uuid.NewV4().String()

	if _, err := p.jobsCollection.InsertOne(ctx, job); err != nil {
		return "", errors.Wrapf(err, "error creating job %q", job.ID)
	}
	return job.ID, nil
}

func (p *projectStore) GetJobsByEventID(
	eventID string,
) ([]brignext.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.jobsCollection.Find(ctx, bson.M{"eventID": eventID})
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving jobs for event %q", eventID)
	}
	jobs := []brignext.Job{}
	for cur.Next(ctx) {
		job := brignext.Job{}
		err := cur.Decode(&job)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding job for event %q", eventID)
		}
		jobs = append(jobs, job)
	}

	return jobs, nil
}

func (p *projectStore) UpdateJobStatus(jobID string, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.jobsCollection.UpdateOne(
			ctx,
			bson.M{
				"_id": jobID,
			},
			bson.M{
				"$set": bson.M{"status": status},
			},
		); err != nil {
		return errors.Wrapf(err, "error updating status for job %q", jobID)
	}
	return nil
}

func (p *projectStore) GetJob(id string) (brignext.Job, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	job := brignext.Job{}

	result := p.jobsCollection.FindOne(ctx, bson.M{"_id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return job, false, nil
	}
	if result.Err() != nil {
		return job, false, errors.Wrapf(result.Err(), "error retrieving job %q", id)
	}

	if err := result.Decode(&job); err != nil {
		return job, false, errors.Wrapf(err, "error decoding job %q", id)
	}
	return job, true, nil
}

func (p *projectStore) DeleteJobsByEventID(eventID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.eventsCollection.DeleteMany(ctx, bson.M{"everntID": eventID}); err != nil {
		return errors.Wrapf(err, "error deleting jobs for event %q", eventID)
	}

	return nil
}
