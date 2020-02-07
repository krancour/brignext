package mongodb

import (
	"context"

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

	usersCollection := database.Collection("users")
	if _, err := usersCollection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"username": 1,
			},
			Options: &options.IndexOptions{
				Unique: &unique,
			},
		},
	); err != nil {
		return nil, errors.Wrap(err, "error adding indexes to users collection")
	}

	serviceAccountsCollection := database.Collection("service-accounts")
	if _, err := serviceAccountsCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"name": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
				},
			},
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

	return &userStore{
		usersCollection:           usersCollection,
		serviceAccountsCollection: serviceAccountsCollection,
	}, nil
}

func (u *userStore) CreateUser(user brignext.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()
	if _, err :=
		u.usersCollection.InsertOne(ctx, user); err != nil {
		return errors.Wrapf(err, "error creating user %q", user.Username)
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

func (u *userStore) GetUser(username string) (brignext.User, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	user := brignext.User{}

	result := u.usersCollection.FindOne(ctx, bson.M{"username": username})
	if result.Err() == mongo.ErrNoDocuments {
		return user, false, nil
	}
	if result.Err() != nil {
		return user, false, errors.Wrapf(
			result.Err(),
			"error retrieving user %q",
			username,
		)
	}
	if err := result.Decode(&user); err != nil {
		return user, false, errors.Wrapf(err, "error decoding user %q", username)
	}
	return user, true, nil
}

func (u *userStore) DeleteUser(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.usersCollection.DeleteOne(ctx, bson.M{"username": username}); err != nil {
		return errors.Wrapf(err, "error deleting user %q", username)
	}
	return nil
}

func (u *userStore) CreateServiceAccount(
	serviceAccount brignext.ServiceAccount,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	hashedToken := crypto.ShortSHA("", serviceAccount.Token)
	// The bson struct tags should stop this clear text field from being
	// persisted, but this is here for good measure.
	serviceAccount.Token = ""

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
		return errors.Wrapf(
			err,
			"error creating service account %q",
			serviceAccount.Name,
		)
	}
	return nil
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
	name string,
) (brignext.ServiceAccount, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	serviceAccount := brignext.ServiceAccount{}

	result := u.serviceAccountsCollection.FindOne(ctx, bson.M{"name": name})
	if result.Err() == mongo.ErrNoDocuments {
		return serviceAccount, false, nil
	}
	if result.Err() != nil {
		return serviceAccount, false, errors.Wrapf(
			result.Err(),
			"error retrieving service account %q",
			name,
		)
	}
	if err := result.Decode(&serviceAccount); err != nil {
		return serviceAccount, false, errors.Wrapf(
			err,
			"error decoding service account %q",
			name,
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

func (u *userStore) DeleteServiceAccount(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.serviceAccountsCollection.DeleteOne(
			ctx,
			bson.M{"name": name},
		); err != nil {
		return errors.Wrapf(err, "error deleting service account %q", name)
	}
	return nil
}
