package mongodb

import (
	"context"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type projectStore struct {
	projectsCollection *mongo.Collection
	buildsCollection   *mongo.Collection
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

	buildsCollection := database.Collection("builds")
	if _, err := buildsCollection.Indexes().CreateMany(
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
			{
				Keys: bson.M{
					"projectName": 1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to builds collection")
	}

	jobsCollection := database.Collection("jobs")
	if _, err := jobsCollection.Indexes().CreateMany(
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
			{
				Keys: bson.M{
					"buildID": 1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to builds collection")
	}

	return &projectStore{
		projectsCollection: projectsCollection,
		buildsCollection:   buildsCollection,
		jobsCollection:     jobsCollection,
	}, nil
}

func (p *projectStore) CreateProject(project brignext.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()
	if _, err := p.projectsCollection.InsertOne(ctx, project); err != nil {
		return errors.Wrapf(err, "error creating project %q", project.Name)
	}
	return nil
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

func (p *projectStore) GetProject(name string) (brignext.Project, bool, error) {
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

func (p *projectStore) DeleteProject(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.projectsCollection.DeleteOne(ctx, bson.M{"name": name}); err != nil {
		return errors.Wrapf(err, "error deleting project %q", name)
	}
	return nil
}

func (p *projectStore) CreateBuild(build brignext.Build) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()
	if _, err := p.buildsCollection.InsertOne(ctx, build); err != nil {
		return errors.Wrapf(err, "error creating build %q", build.ID)
	}
	return nil
}

func (p *projectStore) GetBuilds() ([]brignext.Build, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.buildsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving builds")
	}
	builds := []brignext.Build{}
	for cur.Next(ctx) {
		build := brignext.Build{}
		err := cur.Decode(&build)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding builds")
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (p *projectStore) GetBuildsByProjectName(
	projectName string,
) ([]brignext.Build, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.buildsCollection.Find(ctx, bson.M{"projectname": projectName})
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving builds for project %q",
			projectName,
		)
	}
	builds := []brignext.Build{}
	for cur.Next(ctx) {
		build := brignext.Build{}
		err := cur.Decode(&build)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"error decoding builds for project %q",
				projectName,
			)
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (p *projectStore) GetBuild(id string) (brignext.Build, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	build := brignext.Build{}

	result := p.buildsCollection.FindOne(ctx, bson.M{"id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return build, false, nil
	}
	if result.Err() != nil {
		return build, false, errors.Wrapf(
			result.Err(),
			"error retrieving build %q",
			id,
		)
	}
	if err := result.Decode(&build); err != nil {
		return build, false, errors.Wrapf(err, "error decoding build %q", id)
	}
	return build, true, nil
}

func (p *projectStore) DeleteBuild(
	id string,
	options storage.DeleteBuildOptions,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	criteria := bson.M{
		"id": id,
	}
	if !options.DeleteRunningBuilds {
		// TODO: Amend the criteria appropriately
	}
	if _, err :=
		p.buildsCollection.DeleteOne(ctx, criteria); err != nil {
		return errors.Wrapf(err, "error deleting build %q", id)
	}

	return nil
}

func (p *projectStore) UpdateWorker(
	buildID string,
	worker brignext.Worker,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.buildsCollection.UpdateOne(
			ctx,
			bson.M{
				"id": buildID,
			},
			bson.M{
				"$set": bson.M{
					"worker": worker,
				},
			},
		); err != nil {
		return errors.Wrapf(err, "error updating worker for build %q", buildID)
	}

	return nil
}

func (p *projectStore) CreateJob(job brignext.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()
	if _, err := p.jobsCollection.InsertOne(ctx, job); err != nil {
		return errors.Wrapf(err, "error creating job %q", job.ID)
	}
	return nil
}

func (p *projectStore) GetJobsByBuildID(buildID string) ([]brignext.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.jobsCollection.Find(ctx, bson.M{"buildid": buildID})
	if err != nil {
		return nil, errors.Wrapf(err, "error retrieving jobs for build %q", buildID)
	}
	jobs := []brignext.Job{}
	for cur.Next(ctx) {
		job := brignext.Job{}
		err := cur.Decode(&job)
		if err != nil {
			return nil, errors.Wrapf(err, "error decoding job for build %q", buildID)
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
				"id": jobID,
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

	result := p.jobsCollection.FindOne(ctx, bson.M{"id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return job, false, nil
	}
	if result.Err() != nil {
		return job, false, errors.Wrapf(result.Err(), "error retrieving job %q", id)
	}

	if err := result.Decode(&job); err != nil {
		return job, false, errors.Wrapf(err, "error retrieving job %q", id)
	}
	return job, true, nil
}

func (p *projectStore) DeleteJobsByBuildID(buildID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.buildsCollection.DeleteMany(ctx, bson.M{"buildid": buildID}); err != nil {
		return errors.Wrapf(err, "error deleting jobs for build %q", buildID)
	}

	return nil
}
