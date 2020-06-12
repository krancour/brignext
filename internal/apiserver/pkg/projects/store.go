package projects

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// TODO: DRY this up
var mongodbTimeout = 5 * time.Second

type Store interface {
	Create(context.Context, brignext.Project) error
	List(context.Context) (brignext.ProjectList, error)
	ListSubscribers(
		ctx context.Context,
		event brignext.Event,
	) (brignext.ProjectList, error)
	Get(context.Context, string) (brignext.Project, error)
	Update(context.Context, brignext.Project) error
	Delete(context.Context, string) error

	// TODO: DRY this up
	DoTx(context.Context, func(context.Context) error) error

	CheckHealth(context.Context) error
}

type store struct {
	collection *mongo.Collection
}

func NewStore(database *mongo.Database) (Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("projects")
	if _, err := collection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"metadata.id": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
			// The next two indexes are involved in locating projects that have
			// subscribed to a given event. Two of the fields (types and labels) are
			// array fields. Indexes involving multiple array fields aren't permitted
			// by MongoDB, so we create two separate indexes and MongoDB *should*
			// utilize the intersection of the two indexes to support such queries.
			{
				Keys: bson.M{
					"spec.eventSubscriptions.source": 1,
					"spec.eventSubscriptions.types":  1,
				},
			},
			{
				Keys: bson.M{
					"spec.eventSubscriptions.labels": 1,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to projects collection",
		)
	}
	return &store{
		collection: collection,
	}, nil
}

func (s *store) Create(
	ctx context.Context,
	project brignext.Project,
) error {
	now := time.Now()
	project.Created = &now
	if _, err := s.collection.InsertOne(ctx, project); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return brignext.NewErrConflict(
					"Project",
					project.ID,
					fmt.Sprintf("A project with the ID %q already exists.", project.ID),
				)
			}
		}
		return errors.Wrapf(err, "error inserting new project %q", project.ID)
	}
	return nil
}

func (s *store) List(
	ctx context.Context,
) (brignext.ProjectList, error) {
	projectList := brignext.NewProjectList()
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.id": 1})
	cur, err := s.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return projectList, errors.Wrap(err, "error finding projects")
	}
	if err := cur.All(ctx, &projectList.Items); err != nil {
		return projectList, errors.Wrap(err, "error decoding projects")
	}
	return projectList, nil
}

func (s *store) ListSubscribers(
	ctx context.Context,
	event brignext.Event,
) (brignext.ProjectList, error) {
	projectList := brignext.NewProjectList()
	subscriptionMatchCriteria := bson.M{
		"source": event.Source,
		"types": bson.M{
			"$in": []string{event.Type, "*"},
		},
	}
	if len(event.Labels) > 0 {
		labelConditions := make([]bson.M, len(event.Labels))
		var i int
		for key, value := range event.Labels {
			labelConditions[i] = bson.M{
				"$elemMatch": bson.M{
					"key":   key,
					"value": value,
				},
			}
			i++
		}
		subscriptionMatchCriteria["labels"] = bson.M{
			"$all": labelConditions,
		}
	}
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.id": 1})
	cur, err := s.collection.Find(
		ctx,
		bson.M{
			"spec.eventSubscriptions": bson.M{
				"$elemMatch": subscriptionMatchCriteria,
			},
		},
		findOptions,
	)
	if err != nil {
		return projectList, errors.Wrap(err, "error finding projects")
	}
	if err := cur.All(ctx, &projectList.Items); err != nil {
		return projectList, errors.Wrap(err, "error decoding projects")
	}
	return projectList, nil
}

func (s *store) Get(
	ctx context.Context,
	id string,
) (brignext.Project, error) {
	project := brignext.Project{}
	res := s.collection.FindOne(ctx, bson.M{"metadata.id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return project, brignext.NewErrNotFound("Project", id)
	}
	if res.Err() != nil {
		return project, errors.Wrapf(res.Err(), "error finding project %q", id)
	}
	if err := res.Decode(&project); err != nil {
		return project, errors.Wrapf(err, "error decoding project %q", id)
	}
	return project, nil
}

func (s *store) Update(
	ctx context.Context, project brignext.Project,
) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{
			"metadata.id": project.ID,
		},
		bson.M{
			"$set": bson.M{
				"metadata.lastUpdated": time.Now(),
				"spec":                 project.Spec,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error replacing project %q", project.ID)
	}
	if res.MatchedCount == 0 {
		return brignext.NewErrNotFound("Project", project.ID)
	}
	return nil
}

func (s *store) Delete(ctx context.Context, id string) error {
	res, err := s.collection.DeleteOne(ctx, bson.M{"metadata.id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting project %q", id)
	}
	if res.DeletedCount == 0 {
		return brignext.NewErrNotFound("Project", id)
	}
	return nil
}

func (s *store) DoTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if err := s.collection.Database().Client().UseSession(
		ctx,
		func(sc mongo.SessionContext) error {
			if err := sc.StartTransaction(); err != nil {
				return errors.Wrapf(err, "error starting transaction")
			}
			if err := fn(sc); err != nil {
				return err
			}
			if err := sc.CommitTransaction(sc); err != nil {
				return errors.Wrap(err, "error committing transaction")
			}
			return nil
		},
	); err != nil {
		return err
	}
	return nil
}

func (s *store) CheckHealth(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := s.collection.Database().Client().Ping(
		pingCtx,
		readpref.Primary(),
	); err != nil {
		return errors.Wrap(err, "error pinging mongodb")
	}
	return nil
}
