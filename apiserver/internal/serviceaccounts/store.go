package serviceaccounts

import (
	"context"
	"fmt"
	"time"

	errs "github.com/krancour/brignext/v2/internal/errors"
	"github.com/krancour/brignext/v2/internal/mongodb"
	brignext "github.com/krancour/brignext/v2/sdk"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store interface {
	Create(context.Context, brignext.ServiceAccount) error
	List(context.Context) (brignext.ServiceAccountList, error)
	Get(context.Context, string) (brignext.ServiceAccount, error)
	GetByHashedToken(context.Context, string) (brignext.ServiceAccount, error)
	Lock(context.Context, string) error
	Unlock(ctx context.Context, id string, newHashedToken string) error

	DoTx(context.Context, func(context.Context) error) error

	CheckHealth(context.Context) error
}

type store struct {
	*mongodb.BaseStore
	collection *mongo.Collection
}

func NewStore(database *mongo.Database) (Store, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), mongodb.CreateIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("service-accounts")
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
			// Fast lookup by bearer token
			{
				Keys: bson.M{
					"hashedToken": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to service accounts collection",
		)
	}
	return &store{
		BaseStore: &mongodb.BaseStore{
			Database: database,
		},
		collection: collection,
	}, nil
}

func (s *store) Create(
	ctx context.Context,
	serviceAccount brignext.ServiceAccount,
) error {
	now := time.Now()
	serviceAccount.Created = &now
	if _, err := s.collection.InsertOne(
		ctx,
		serviceAccount,
	); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return errs.NewErrConflict(
					"ServiceAccount",
					serviceAccount.ID,
					fmt.Sprintf(
						"A service account with the ID %q already exists.",
						serviceAccount.ID,
					),
				)
			}
		}
		return errors.Wrapf(
			err,
			"error inserting new service account %q",
			serviceAccount.ID,
		)
	}
	return nil
}

func (s *store) List(
	ctx context.Context,
) (brignext.ServiceAccountList, error) {
	serviceAccountList := brignext.NewServiceAccountList()
	findOptions := options.Find()
	findOptions.SetSort(bson.M{"metadata.id": 1})
	cur, err := s.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return serviceAccountList,
			errors.Wrap(err, "error finding service accounts")
	}
	if err := cur.All(ctx, &serviceAccountList.Items); err != nil {
		return serviceAccountList,
			errors.Wrap(err, "error decoding service accounts")
	}
	return serviceAccountList, nil
}

func (s *store) Get(
	ctx context.Context,
	id string,
) (brignext.ServiceAccount, error) {
	serviceAccount := brignext.ServiceAccount{}
	res := s.collection.FindOne(ctx, bson.M{"metadata.id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, errs.NewErrNotFound("ServiceAccount", id)
	}
	if res.Err() != nil {
		return serviceAccount, errors.Wrapf(
			res.Err(),
			"error finding service account %q",
			id,
		)
	}
	if err := res.Decode(&serviceAccount); err != nil {
		return serviceAccount, errors.Wrapf(
			err,
			"error decoding service account %q",
			id,
		)
	}
	return serviceAccount, nil
}

func (s *store) GetByHashedToken(
	ctx context.Context,
	hashedToken string,
) (brignext.ServiceAccount, error) {
	serviceAccount := brignext.ServiceAccount{}
	res :=
		s.collection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, errs.NewErrNotFound("ServiceAccount", "")
	}
	if res.Err() != nil {
		return serviceAccount, errors.Wrap(
			res.Err(),
			"error finding service account by hashed token",
		)
	}
	if err := res.Decode(&serviceAccount); err != nil {
		return serviceAccount, errors.Wrap(
			err,
			"error decoding service account",
		)
	}
	return serviceAccount, nil
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
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return errs.NewErrNotFound("ServiceAccount", id)
	}
	// Note, unlike the case of locking a user, there are no sessions to delete
	// because service accounts use non-expiring, sessionless tokens.
	return nil
}

func (s *store) Unlock(
	ctx context.Context,
	id string,
	newHashedToken string,
) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"metadata.id": id},
		bson.M{
			"$set": bson.M{
				"locked":      nil,
				"hashedToken": newHashedToken,
			},
		},
	)
	if err != nil {
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return errs.NewErrNotFound("ServiceAccount", id)
	}
	return nil
}