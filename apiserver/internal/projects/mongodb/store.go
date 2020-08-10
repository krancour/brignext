package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/projects"
	brignext "github.com/krancour/brignext/v2/apiserver/internal/sdk"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const createIndexTimeout = 5 * time.Second

type store struct {
	collection       *mongo.Collection
	eventsCollection *mongo.Collection
}

func NewStore(database *mongo.Database) (projects.Store, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("projects")
	if _, err := collection.Indexes().CreateMany(
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
			// The next two indexes are involved in locating projects that have
			// subscribed to a given event. Two of the fields (types and labels) are
			// array fields. Indexes involving multiple array fields aren't permitted
			// by MongoDB, so we create two separate indexes and MongoDB *should*
			// utilize the intersection of the two indexes to support such queries.
			//
			// TODO: CosmosDB doesn't support these indices. We can probably live
			// without them because the number of projects in a cluster should usually
			// be pretty small in comparison to the number of events. What we should
			// probably do is make these indices optional through a configuration
			// setting.
			//
			// {
			// 	Keys: bson.M{
			// 		"spec.eventSubscriptions.source": 1,
			// 		"spec.eventSubscriptions.types":  1,
			// 	},
			// },
			// {
			// 	Keys: bson.M{
			// 		"spec.eventSubscriptions.labels": 1,
			// 	},
			// },
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to projects collection",
		)
	}
	return &store{
		collection:       collection,
		eventsCollection: database.Collection("events"),
	}, nil
}

func (s *store) Create(
	ctx context.Context,
	project brignext.Project,
) error {
	if _, err := s.collection.InsertOne(ctx, project); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &brignext.ErrConflict{
					Type: "Project",
					ID:   project.ID,
					Reason: fmt.Sprintf(
						"A project with the ID %q already exists.",
						project.ID,
					),
				}
			}
		}
		return errors.Wrapf(err, "error inserting new project %q", project.ID)
	}
	return nil
}

func (s *store) List(
	ctx context.Context,
	opts brignext.ProjectListOptions,
) (brignext.ProjectList, error) {
	const limit = 2 // TODO: Don't hard code this
	projects := brignext.ProjectList{}

	criteria := bson.M{}
	if opts.Continue != "" {
		criteria["id"] = bson.M{"$gt": opts.Continue}
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"id": 1})
	findOptions.SetLimit(limit)
	cur, err := s.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return projects, errors.Wrap(err, "error finding projects")
	}
	if err := cur.All(ctx, &projects.Items); err != nil {
		return projects, errors.Wrap(err, "error decoding projects")
	}

	if len(projects.Items) == limit {
		continueID := projects.Items[limit-1].ID
		criteria["id"] = bson.M{"$gt": continueID}
		remaining, err := s.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return projects, errors.Wrap(err, "error counting remaining projects")
		}
		if remaining > 0 {
			projects.Continue = continueID
			projects.RemainingItemCount = remaining
		}
	}

	return projects, nil
}

func (s *store) ListSubscribers(
	ctx context.Context,
	event brignext.Event,
) (brignext.ProjectList, error) {
	projects := brignext.ProjectList{}
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
	findOptions.SetSort(bson.M{"id": 1})
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
		return projects, errors.Wrap(err, "error finding projects")
	}
	if err := cur.All(ctx, &projects.Items); err != nil {
		return projects, errors.Wrap(err, "error decoding projects")
	}
	return projects, nil
}

func (s *store) Get(
	ctx context.Context,
	id string,
) (brignext.Project, error) {
	project := brignext.Project{}
	res := s.collection.FindOne(ctx, bson.M{"id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return project, &brignext.ErrNotFound{
			Type: "Project",
			ID:   id,
		}
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
			"id": project.ID,
		},
		bson.M{
			"$set": bson.M{
				"lastUpdated": time.Now(),
				"spec":        project.Spec,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error replacing project %q", project.ID)
	}
	if res.MatchedCount == 0 {
		return &brignext.ErrNotFound{
			Type: "Project",
			ID:   project.ID,
		}
	}
	return nil
}

func (s *store) Delete(ctx context.Context, id string) error {
	// TODO: We'd like to use transaction semantics here, but transactions in
	// MongoDB are dicey, so we should refine this strategy to where a
	// partially completed delete leaves us, overall, in a tolerable state.

	res, err := s.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting project %q", id)
	}
	if res.DeletedCount == 0 {
		return &brignext.ErrNotFound{
			Type: "Project",
			ID:   id,
		}
	}

	// Cascade the delete to the project's events
	if _, err := s.eventsCollection.DeleteMany(
		ctx,
		bson.M{
			"projectID": id,
		},
	); err != nil {
		return errors.Wrapf(err, "error deleting events for project %q", id)
	}

	return nil
}
