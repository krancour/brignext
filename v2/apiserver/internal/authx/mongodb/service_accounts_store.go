package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/brigadecore/brigade/v2/apiserver/internal/meta"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type serviceAccountsStore struct {
	collection *mongo.Collection
}

func NewServiceAccountsStore(
	database *mongo.Database,
) (authx.ServiceAccountsStore, error) {
	ctx, cancel :=
		context.WithTimeout(context.Background(), createIndexTimeout)
	defer cancel()
	unique := true
	collection := database.Collection("service-accounts")
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
	return &serviceAccountsStore{
		collection: collection,
	}, nil
}

func (s *serviceAccountsStore) Create(
	ctx context.Context,
	serviceAccount authx.ServiceAccount,
) error {
	if _, err := s.collection.InsertOne(
		ctx,
		serviceAccount,
	); err != nil {
		if writeException, ok := err.(mongo.WriteException); ok {
			if len(writeException.WriteErrors) == 1 &&
				writeException.WriteErrors[0].Code == 11000 {
				return &meta.ErrConflict{
					Type: "ServiceAccount",
					ID:   serviceAccount.ID,
					Reason: fmt.Sprintf(
						"A service account with the ID %q already exists.",
						serviceAccount.ID,
					),
				}
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

func (s *serviceAccountsStore) List(
	ctx context.Context,
	opts meta.ListOptions,
) (authx.ServiceAccountList, error) {
	serviceAccounts := authx.ServiceAccountList{}

	criteria := bson.M{}
	if opts.Continue != "" {
		criteria["id"] = bson.M{"$gt": opts.Continue}
	}

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"id": 1})
	findOptions.SetLimit(opts.Limit)
	cur, err := s.collection.Find(ctx, criteria, findOptions)
	if err != nil {
		return serviceAccounts,
			errors.Wrap(err, "error finding service accounts")
	}
	if err := cur.All(ctx, &serviceAccounts.Items); err != nil {
		return serviceAccounts,
			errors.Wrap(err, "error decoding service accounts")
	}

	if int64(len(serviceAccounts.Items)) == opts.Limit {
		continueID := serviceAccounts.Items[opts.Limit-1].ID
		criteria["id"] = bson.M{"$gt": continueID}
		remaining, err := s.collection.CountDocuments(ctx, criteria)
		if err != nil {
			return serviceAccounts,
				errors.Wrap(err, "error counting remaining service accounts")
		}
		if remaining > 0 {
			serviceAccounts.Continue = continueID
			serviceAccounts.RemainingItemCount = remaining
		}
	}

	return serviceAccounts, nil
}

func (s *serviceAccountsStore) Get(
	ctx context.Context,
	id string,
) (authx.ServiceAccount, error) {
	serviceAccount := authx.ServiceAccount{}
	res := s.collection.FindOne(ctx, bson.M{"id": id})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, &meta.ErrNotFound{
			Type: "ServiceAccount",
			ID:   id,
		}
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

func (s *serviceAccountsStore) GetByHashedToken(
	ctx context.Context,
	hashedToken string,
) (authx.ServiceAccount, error) {
	serviceAccount := authx.ServiceAccount{}
	res :=
		s.collection.FindOne(ctx, bson.M{"hashedToken": hashedToken})
	if res.Err() == mongo.ErrNoDocuments {
		return serviceAccount, &meta.ErrNotFound{
			Type: "ServiceAccount",
		}
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

func (s *serviceAccountsStore) Lock(ctx context.Context, id string) error {
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
		return errors.Wrapf(err, "error updating service account %q", id)
	}
	if res.MatchedCount == 0 {
		return &meta.ErrNotFound{
			Type: "ServiceAccount",
			ID:   id,
		}
	}
	// Note, unlike the case of locking a user, there are no sessions to delete
	// because service accounts use non-expiring, sessionless tokens.
	return nil
}

func (s *serviceAccountsStore) Unlock(
	ctx context.Context,
	id string,
	newHashedToken string,
) error {
	res, err := s.collection.UpdateOne(
		ctx,
		bson.M{"id": id},
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
		return &meta.ErrNotFound{
			Type: "ServiceAccount",
			ID:   id,
		}
	}
	return nil
}
