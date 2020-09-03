package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/krancour/brignext/v2/apiserver/internal/authx"
	"github.com/krancour/brignext/v2/apiserver/internal/authx/users"
	"github.com/krancour/brignext/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const createIndexTimeout = 5 * time.Second

type store struct {
	collection         *mongo.Collection
	sessionsCollection *mongo.Collection
}

func NewStore(database *mongo.Database) (users.Store, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("users")
	if _, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"id": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to users collection")
	}
	return &store{
		collection:         collection,
		sessionsCollection: database.Collection("sessions"),
	}, nil
}

func (s *store) Create(ctx context.Context, user authx.User) error {
	if _, err :=
		s.collection.InsertOne(ctx, user); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &meta.ErrConflict{
					Type:   "User",
					ID:     user.ID,
					Reason: fmt.Sprintf("A user with the ID %q already exists.", user.ID),
				}
			}
		}
		return errors.Wrapf(err, "error inserting new user %q", user.ID)
	}
	return nil
}

func (s *store) Count(ctx context.Context) (int64, error) {
	count, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, errors.Wrap(err, "error counting users")
	}
	return count, nil
}

func (s *store) List(
	ctx context.Context,
	_ authx.UsersSelector,
	opts meta.ListOptions,
) (authx.UserList, error) {
	users := authx.UserList{}

	criteria := bson.M{}
	if opts.Continue != "" {
		criteria["id"] = bson.M{"$gt": opts.Continue}
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"id": 1})
	findOptions.SetLimit(opts.Limit)
	cur, err := s.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return users, errors.Wrap(err, "error finding users")
	}
	if err := cur.All(ctx, &users.Items); err != nil {
		return users, errors.Wrap(err, "error decoding users")
	}

	if int64(len(users.Items)) == opts.Limit {
		continueID := users.Items[opts.Limit-1].ID
		criteria["id"] = bson.M{"$gt": continueID}
		remaining, err := s.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return users, errors.Wrap(err, "error counting remaining users")
		}
		if remaining > 0 {
			users.Continue = continueID
			users.RemainingItemCount = remaining
		}
	}

	return users, nil
}

func (s *store) Get(
	ctx context.Context,
	id string,
) (authx.User, error) {
	user := authx.User{}
	res := s.collection.FindOne(ctx, bson.M{"id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return user, &meta.ErrNotFound{
			Type: "User",
			ID:   id,
		}
	}
	if res.Err() != nil {
		return user, errors.Wrapf(res.Err(), "error finding user %q", id)
	}
	if err := res.Decode(&user); err != nil {
		return user, errors.Wrapf(err, "error decoding user %q", id)
	}
	return user, nil
}

func (s *store) Lock(ctx context.Context, id string) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"locked": time.Now(),
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "User",
			ID:   id,
		}
	}

	// Now delete all the user's sessions.
	if _, err = s.sessionsCollection.DeleteMany(
		ctx,
		bson.M{
			"userID": id,
		},
	); err != nil {
		return errors.Wrapf(err, "error deleting sessions for user %q", id)
	}

	return nil
}

func (s *store) Unlock(ctx context.Context, id string) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
		bson.M{
			"$set": bson.M{
				"locked": nil,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating user %q", id)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "User",
			ID:   id,
		}
	}
	return nil
}

func (s *store) GrantRole(
	ctx context.Context,
	userID string,
	roles ...authx.Role,
) error {
	for _, role := range roles {
		if _, err := s.collection.UpdateOne(
			ctx,
			bson.M{
				"id": userID,
				"roles": bson.M{
					"$nin": []authx.Role{role},
				},
			},
			bson.M{
				"$push": bson.M{
					"roles": bson.M{
						"$each": []authx.Role{role},
						"$sort": bson.M{
							"type":  1,
							"name":  1,
							"scope": 1,
						},
					},
				},
			},
		); err != nil {
			return errors.Wrapf(err, "error updating user %q", userID)
		}
	}
	return nil
}

func (s *store) RevokeRole(
	ctx context.Context,
	userID string,
	roles ...authx.Role,
) error {
	for _, role := range roles {
		if _, err := s.collection.UpdateOne(
			ctx,
			bson.M{
				"id": userID,
			},
			bson.M{
				"$pull": bson.M{
					"roles": bson.M{
						"$in": []authx.Role{role},
					},
				},
			},
		); err != nil {
			return errors.Wrapf(err, "error updating user %q", userID)
		}
	}
	return nil
}
