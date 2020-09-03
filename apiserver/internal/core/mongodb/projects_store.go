package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/core"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type projectsStore struct {
	collection       *mongo.Collection
	eventsCollection *mongo.Collection
}

func NewProjectsStore(database *mongo.Database) (core.ProjectsStore, error) {
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
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to projects collection",
		)
	}
	return &projectsStore{
		collection:       collection,
		eventsCollection: database.Collection("events"),
	}, nil
}

func (p *projectsStore) Create(
	ctx context.Context,
	project core.Project,
) error {
	if _, err := p.collection.InsertOne(ctx, project); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &meta.ErrConflict{
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

func (p *projectsStore) List(
	ctx context.Context,
	_ core.ProjectsSelector,
	opts meta.ListOptions,
) (core.ProjectList, error) {
	projects := core.ProjectList{}

	criteria := bson.M{}
	if opts.Continue != "" {
		criteria["id"] = bson.M{"$gt": opts.Continue}
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"id": 1})
	findOptions.SetLimit(opts.Limit)
	cur, err := p.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return projects, errors.Wrap(err, "error finding projects")
	}
	if err := cur.All(ctx, &projects.Items); err != nil {
		return projects, errors.Wrap(err, "error decoding projects")
	}

	if int64(len(projects.Items)) == opts.Limit {
		continueID := projects.Items[opts.Limit-1].ID
		criteria["id"] = bson.M{"$gt": continueID}
		remaining, err := p.collection.CountDocuments(ctx, criteria)
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

func (p *projectsStore) ListSubscribers(
	ctx context.Context,
	event core.Event,
) (core.ProjectList, error) {
	projects := core.ProjectList{}
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
	cur, err := p.collection.Find(
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

func (p *projectsStore) Get(
	ctx context.Context,
	id string,
) (core.Project, error) {
	project := core.Project{}
	res := p.collection.FindOne(ctx, bson.M{"id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return project, &meta.ErrNotFound{
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

func (p *projectsStore) Update(
	ctx context.Context, project core.Project,
) error {
	res, err := p.collection.UpdateOne(
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
		return &meta.ErrNotFound{
			Type: "Project",
			ID:   project.ID,
		}
	}
	return nil
}

func (p *projectsStore) Delete(ctx context.Context, id string) error {
	res, err := p.collection.DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return errors.Wrapf(err, "error deleting project %q", id)
	}
	if res.DeletedCount == 0 {
		return &meta.ErrNotFound{
			Type: "Project",
			ID:   id,
		}
	}

	// TODO: The service should worry about cascading deletes
	// Cascade the delete to the project's events
	if _, err := p.eventsCollection.DeleteMany(
		ctx,
		bson.M{
			"projectID": id,
		},
	); err != nil {
		return errors.Wrapf(err, "error deleting events for project %q", id)
	}

	return nil
}
