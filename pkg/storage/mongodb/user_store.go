package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/crypto"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userStore struct {
	usersCollection           *mongo.Collection
	serviceAccountsCollection *mongo.Collection
}

func NewUserStore(database *mongo.Database) (storage.UserStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	unique := true
	serviceAccountsCollection := database.Collection("service-accounts")
	if _, err := serviceAccountsCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"hashedToken": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(
			err,
			"error adding indexes to service accounts collection",
		)
	}

	return &userStore{
		usersCollection:           database.Collection("users"),
		serviceAccountsCollection: serviceAccountsCollection,
	}, nil
}

func (u *userStore) CreateUser(user brignext.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	user.FirstSeen = time.Now()

	if _, err :=
		u.usersCollection.InsertOne(ctx, user); err != nil {
		return errors.Wrapf(err, "error creating user %q", user.ID)
	}
	return nil
}

func (u *userStore) GetUsers() ([]brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := u.usersCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving users")
	}
	users := []brignext.User{}
	for cur.Next(ctx) {
		user := brignext.User{}
		err := cur.Decode(&user)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding users")
		}
		users = append(users, user)
	}
	return users, nil
}

func (u *userStore) GetUser(id string) (brignext.User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	user := brignext.User{}

	result := u.usersCollection.FindOne(ctx, bson.M{"_id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return user, false, nil
	}
	if result.Err() != nil {
		return user, false, errors.Wrapf(
			result.Err(),
			"error retrieving user %q",
			id,
		)
	}
	if err := result.Decode(&user); err != nil {
		return user, false, errors.Wrapf(err, "error decoding user %q", id)
	}
	return user, true, nil
}

func (u *userStore) LockUser(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.usersCollection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$set": bson.M{"locked": true},
			},
		); err != nil {
		return errors.Wrapf(err, "error locking user %q", id)
	}

	// TODO: Delete all of the locked user's sessions

	return nil
}

func (u *userStore) UnlockUser(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.usersCollection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$unset": bson.M{"locked": 1},
			},
		); err != nil {
		return errors.Wrapf(err, "error unlocking user %q", id)
	}
	return nil
}

func (u *userStore) CreateServiceAccount(
	serviceAccount brignext.ServiceAccount,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	now := time.Now()
	serviceAccount.Created = &now

	token := crypto.NewToken(256)
	hashedToken := crypto.ShortSHA("", token)

	if _, err :=
		u.serviceAccountsCollection.InsertOne(
			ctx,
			struct {
				brignext.ServiceAccount `bson:",inline"`
				HashedToken             string `bson:"hashedToken"`
			}{
				ServiceAccount: serviceAccount,
				HashedToken:    hashedToken,
			},
		); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating service account %q",
			serviceAccount.ID,
		)
	}
	return token, nil
}

func (u *userStore) GetServiceAccounts() ([]brignext.ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := u.serviceAccountsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving service accounts")
	}
	serviceAccounts := []brignext.ServiceAccount{}
	for cur.Next(ctx) {
		serviceAccount := brignext.ServiceAccount{}
		err := cur.Decode(&serviceAccount)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding service accounts")
		}
		serviceAccounts = append(serviceAccounts, serviceAccount)
	}
	return serviceAccounts, nil
}

func (u *userStore) GetServiceAccount(
	id string,
) (brignext.ServiceAccount, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	serviceAccount := brignext.ServiceAccount{}

	result := u.serviceAccountsCollection.FindOne(ctx, bson.M{"_id": id})
	if result.Err() == mongo.ErrNoDocuments {
		return serviceAccount, false, nil
	}
	if result.Err() != nil {
		return serviceAccount, false, errors.Wrapf(
			result.Err(),
			"error retrieving service account %q",
			id,
		)
	}
	if err := result.Decode(&serviceAccount); err != nil {
		return serviceAccount, false, errors.Wrapf(
			err,
			"error decoding service account %q",
			id,
		)
	}

	return serviceAccount, true, nil
}

func (u *userStore) GetServiceAccountByToken(
	token string,
) (brignext.ServiceAccount, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	serviceAccount := brignext.ServiceAccount{}

	result := u.serviceAccountsCollection.FindOne(
		ctx,
		bson.M{"hashedToken": crypto.ShortSHA("", token)},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return serviceAccount, false, nil
	}
	if result.Err() != nil {
		return serviceAccount, false, errors.Wrap(
			result.Err(),
			"error retrieving service account with token [REDACTED]",
		)
	}
	if err := result.Decode(&serviceAccount); err != nil {
		return serviceAccount, false, errors.Wrap(
			err,
			"error decoding service account with token [REDACTED]",
		)
	}

	return serviceAccount, true, nil
}

func (u *userStore) DeleteServiceAccount(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.serviceAccountsCollection.DeleteOne(
			ctx,
			bson.M{"_id": id},
		); err != nil {
		return errors.Wrapf(err, "error deleting service account %q", id)
	}
	return nil
}
