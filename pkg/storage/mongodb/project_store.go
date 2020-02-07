package mongodb

import (
	"context"

	"github.com/brigadecore/brigade/pkg/brigade"
	brigStorage "github.com/brigadecore/brigade/pkg/storage"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type projectStore struct {
	projectsCollection *mongo.Collection
	buildsCollection   *mongo.Collection
	jobsCollection     *mongo.Collection
}

func NewProjectStore(database *mongo.Database) storage.ProjectStore {
	return &projectStore{
		projectsCollection: database.Collection("projects"),
		buildsCollection:   database.Collection("builds"),
		jobsCollection:     database.Collection("jobs"),
	}
}

func (p *projectStore) CreateProject(project *brigade.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	// Prevent duplicate IDs
	// TODO: Do this with a unique index instead?
	result := p.projectsCollection.FindOne(
		ctx,
		bson.M{
			"id": project.ID,
		},
	)
	if result.Err() == nil {
		return errors.Errorf(
			"a project with the id %q already exists",
			project.ID,
		)
	} else if result.Err() != mongo.ErrNoDocuments {
		return errors.Wrapf(
			result.Err(),
			"error checking for existing project with id %q",
			project.ID,
		)
	}

	if _, err :=
		p.projectsCollection.InsertOne(
			ctx,
			project,
		); err != nil {
		return errors.Wrapf(
			err,
			"error creating project with id %q",
			project.ID,
		)
	}
	return nil
}

func (p *projectStore) GetProjects() ([]*brigade.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.projectsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving projects")
	}
	projects := []*brigade.Project{}
	for cur.Next(ctx) {
		project := &brigade.Project{}
		err := cur.Decode(project)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding projects")
		}
		projects = append(projects, project)
	}
	return projects, nil
}

func (p *projectStore) GetProject(id string) (*brigade.Project, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := p.projectsCollection.FindOne(
		ctx,
		bson.M{
			"id": id,
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrapf(
			result.Err(),
			"error retrieving project with id %q",
			id,
		)
	}
	project := &brigade.Project{}
	if err := result.Decode(project); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding project with id %q",
			id,
		)
	}
	return project, nil
}

func (p *projectStore) UpdateProject(project *brigade.Project) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.projectsCollection.ReplaceOne(
			ctx,
			bson.M{
				"id": project.ID,
			},
			project,
		); err != nil {
		return errors.Wrapf(
			err,
			"error updating project with id %q",
			project.ID,
		)
	}
	return nil
}

func (p *projectStore) DeleteProject(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.projectsCollection.DeleteOne(
			ctx,
			bson.M{
				"id": id,
			},
		); err != nil {
		return errors.Wrapf(
			err,
			"error deleting project with id %q",
			id,
		)
	}
	return nil
}

func (p *projectStore) CreateBuild(build *brigade.Build) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	// Prevent duplicate IDs
	// TODO: Do this with a unique index instead?
	result := p.buildsCollection.FindOne(
		ctx,
		bson.M{
			"id": build.ID,
		},
	)
	if result.Err() == nil {
		return errors.Errorf("a build with the id %q already exists", build.ID)
	} else if result.Err() != mongo.ErrNoDocuments {
		return errors.Wrapf(
			result.Err(),
			"error checking for existing build %q",
			build.ID,
		)
	}

	if _, err :=
		p.buildsCollection.InsertOne(
			ctx,
			build,
		); err != nil {
		return errors.Wrapf(
			err,
			"error creating build %q",
			build.ID,
		)
	}
	return nil
}

func (p *projectStore) GetBuilds() ([]*brigade.Build, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.buildsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving builds")
	}
	builds := []*brigade.Build{}
	for cur.Next(ctx) {
		build := &brigade.Build{}
		err := cur.Decode(build)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding builds")
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (p *projectStore) GetProjectBuilds(projectID string) ([]*brigade.Build, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.buildsCollection.Find(
		ctx,
		bson.M{
			"projectid": projectID,
		},
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving builds for project %q",
			projectID,
		)
	}
	builds := []*brigade.Build{}
	for cur.Next(ctx) {
		build := &brigade.Build{}
		err := cur.Decode(build)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"error decoding builds for project %q",
				projectID,
			)
		}
		builds = append(builds, build)
	}
	return builds, nil
}

func (p *projectStore) GetBuild(id string) (*brigade.Build, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := p.buildsCollection.FindOne(
		ctx,
		bson.M{
			"id": id,
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrapf(
			result.Err(),
			"error retrieving build %q",
			id,
		)
	}
	build := &brigade.Build{}
	if err := result.Decode(build); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding build %q",
			id,
		)
	}
	return build, nil
}

func (p *projectStore) DeleteBuild(
	id string,
	options brigStorage.DeleteBuildOptions,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	criteria := bson.M{
		"id": id,
	}
	if options.SkipRunningBuilds {
		// TODO: Amend the criteria appropriately
	}
	if _, err :=
		p.buildsCollection.DeleteOne(ctx, criteria); err != nil {
		return errors.Wrapf(
			err,
			"error deleting build %q",
			id,
		)
	}
	return nil
}

func (p *projectStore) UpdateWorker(worker *brigade.Worker) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		p.buildsCollection.UpdateOne(
			ctx,
			bson.M{
				"id": worker.BuildID,
			},
			bson.M{
				"$set": bson.M{
					"worker": worker,
				},
			},
		); err != nil {
		return errors.Wrapf(
			err,
			"error updating worker for build %q",
			worker.BuildID,
		)
	}
	return nil
}

func (p *projectStore) CreateJob(buildID string, job *brigade.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	// Prevent duplicate IDs
	// TODO: Do this with a unique index instead?
	result := p.jobsCollection.FindOne(
		ctx,
		bson.M{
			"id": job.ID,
		},
	)
	if result.Err() == nil {
		return errors.Errorf("a job with the id %q already exists", job.ID)
	} else if result.Err() != mongo.ErrNoDocuments {
		return errors.Wrapf(
			result.Err(),
			"error checking for existing job %q",
			job.ID,
		)
	}

	brigNextJob := &brigNextJob{
		Job:     job,
		BuildID: buildID,
	}
	if _, err :=
		p.jobsCollection.InsertOne(
			ctx,
			brigNextJob,
		); err != nil {
		return errors.Wrapf(
			err,
			"error creating job %q",
			job.ID,
		)
	}
	return nil
}

func (p *projectStore) GetBuildJobs(buildID string) ([]*brigade.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := p.jobsCollection.Find(
		ctx,
		bson.M{
			"buildid": buildID,
		},
	)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving jobs for build %q",
			buildID,
		)
	}
	jobs := []*brigade.Job{}
	for cur.Next(ctx) {
		job := &brigade.Job{}
		err := cur.Decode(job)
		if err != nil {
			return nil, errors.Wrapf(
				err,
				"error decoding builds for build %q",
				buildID,
			)
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
		return errors.Wrapf(
			err,
			"error updating job %q status",
			jobID,
		)
	}
	return nil
}

func (p *projectStore) GetWorker(buildID string) (*brigade.Worker, error) {
	build, err := p.GetBuild(buildID)
	if err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving worker for build %q",
			buildID,
		)
	}
	return build.Worker, nil
}

func (p *projectStore) GetJob(id string) (*brigade.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := p.jobsCollection.FindOne(
		ctx,
		bson.M{
			"id": id,
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrapf(
			result.Err(),
			"error retrieving build %q",
			id,
		)
	}
	job := &brigade.Job{}
	if err := result.Decode(job); err != nil {
		return nil, errors.Wrapf(
			err,
			"error retrieving job %q",
			id,
		)
	}
	return job, nil
}

// This is a hack to associate jobs to builds. Current brigade doesn't do
// a good job of this.
type brigNextJob struct {
	*brigade.Job `bson:",inline"`
	BuildID      string
}
