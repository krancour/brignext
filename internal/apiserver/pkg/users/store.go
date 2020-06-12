package users

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
	Create(context.Context, brignext.User) error
	List(context.Context) (brignext.UserList, error)
	Get(context.Context, string) (brignext.User, error)
	Lock(context.Context, string) error
	Unlock(context.Context, string) error

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
	collection := database.Collection("users")
	if _, err := collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"metadata.id": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to users collection")
	}
	return &store{
		collection: collection,
	}, nil
}

func (s *store) Create(ctx context.Context, user brignext.User) error {
	now := time.Now()
	user.Created = &now
	if _, err :=
		s.collection.InsertOne(ctx, user); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return brignext.NewErrConflict(
					"User",
					user.ID,
					fmt.Sprintf("A user with the ID %q already exists.", user.ID),
				)
			}
		}
		return errors.Wrapf(err, "error inserting new user %q", user.ID)
	}
	return nil
}

func (s *store) List(ctx context.Context) (brignext.UserList, error) {
	userList := brignext.NewUserList()
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.id": 1})
	cur, err := s.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return userList, errors.Wrap(err, "error finding users")
	}
	if err := cur.All(ctx, &userList.Items); err != nil {
		return userList, errors.Wrap(err, "error decoding users")
	}
	return userList, nil
}

func (s *store) Get(
	ctx context.Context,
	id string,
) (brignext.User, error) {
	user := brignext.User{}
	res := s.collection.FindOne(ctx, bson.M{"metadata.id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return user, brignext.NewErrNotFound("User", id)
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
		bson.M{"metadata.id": id},
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
		return brignext.NewErrNotFound("User", id)
	}
	return nil
}

func (s *store) Unlock(ctx context.Context, id string) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"metadata.id": id},
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
		return brignext.NewErrNotFound("User", id)
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
