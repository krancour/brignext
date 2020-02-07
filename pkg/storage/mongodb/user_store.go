package mongodb

import (
	"context"
	"time"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/crypto"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type userStore struct {
	usersCollection           *mongo.Collection
	serviceAccountsCollection *mongo.Collection
}

func NewUserStore(database *mongo.Database) storage.UserStore {
	return &userStore{
		usersCollection:           database.Collection("users"),
		serviceAccountsCollection: database.Collection("service-accounts"),
	}
}

func (u *userStore) CreateUser(username string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	// Prevent duplicate username
	// TODO: Do this with a unique index instead?
	result := u.usersCollection.FindOne(
		ctx,
		bson.M{
			"username": username,
		},
	)
	if result.Err() == nil {
		return "", errors.Errorf(
			"a user with the username %q already exists",
			username,
		)
	} else if result.Err() != mongo.ErrNoDocuments {
		return "", errors.Wrapf(
			result.Err(),
			"error checking for existing user with the username %q",
			username,
		)
	}

	id := uuid.NewV4().String()

	if _, err :=
		u.usersCollection.InsertOne(
			ctx,
			bson.M{
				"id":        id,
				"username":  username,
				"firstseen": time.Now(),
			},
		); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating user with username %q",
			username,
		)
	}
	return id, nil
}

func (u *userStore) GetUsers() ([]*brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := u.usersCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving users")
	}
	users := []*brignext.User{}
	for cur.Next(ctx) {
		user := &brignext.User{}
		err := cur.Decode(user)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding users")
		}
		users = append(users, user)
	}
	return users, nil
}

func (u *userStore) GetUser(id string) (*brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := u.usersCollection.FindOne(
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
			"error retrieving user %q",
			id,
		)
	}
	user := &brignext.User{}
	if err := result.Decode(user); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding user %q",
			id,
		)
	}
	return user, nil
}

func (u *userStore) GetUserByUsername(username string) (*brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := u.usersCollection.FindOne(
		ctx,
		bson.M{
			"username": username,
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrapf(
			result.Err(),
			"error retrieving user with username %q",
			username,
		)
	}
	user := &brignext.User{}
	if err := result.Decode(user); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding user with username %q",
			username,
		)
	}
	return user, nil
}

func (u *userStore) DeleteUser(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.usersCollection.DeleteOne(
			ctx,
			bson.M{
				"id": id,
			},
		); err != nil {
		return errors.Wrapf(err, "error deleting user %q", id)
	}
	return nil
}

func (u *userStore) DeleteUserByUsername(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.usersCollection.DeleteOne(
			ctx,
			bson.M{
				"username": username,
			},
		); err != nil {
		return errors.Wrapf(err, "error deleting user with username %q", username)
	}
	return nil
}

func (u *userStore) CreateServiceAccount(
	name string,
	description string,
) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	// Prevent duplicate name
	// TODO: Do this with a unique index instead?
	result := u.serviceAccountsCollection.FindOne(
		ctx,
		bson.M{
			"name": name,
		},
	)
	if result.Err() == nil {
		return "", "", errors.Errorf(
			"a service account with the name %q already exists",
			name,
		)
	} else if result.Err() != mongo.ErrNoDocuments {
		return "", "", errors.Wrapf(
			result.Err(),
			"error checking for existing service account with the name %q",
			name,
		)
	}

	id := uuid.NewV4().String()
	token := crypto.NewToken(256)

	if _, err :=
		u.serviceAccountsCollection.InsertOne(
			ctx,
			bson.M{
				"id":          id,
				"name":        name,
				"description": description,
				"hashedtoken": crypto.ShortSHA("", token),
				"created":     time.Now(),
			},
		); err != nil {
		return "", "", errors.Wrapf(
			err,
			"error creating service account with name %q",
			name,
		)
	}
	return id, token, nil
}

func (u *userStore) GetServiceAccounts() ([]*brignext.ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	cur, err := u.serviceAccountsCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving service accounts")
	}
	serviceAccounts := []*brignext.ServiceAccount{}
	for cur.Next(ctx) {
		serviceAccount := &brignext.ServiceAccount{}
		err := cur.Decode(serviceAccount)
		if err != nil {
			return nil, errors.Wrap(err, "error decoding service accounts")
		}
		serviceAccounts = append(serviceAccounts, serviceAccount)
	}
	return serviceAccounts, nil
}

func (u *userStore) GetServiceAccount(
	id string,
) (*brignext.ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := u.serviceAccountsCollection.FindOne(
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
			"error retrieving service account %q",
			id,
		)
	}
	serviceAccount := &brignext.ServiceAccount{}
	if err := result.Decode(serviceAccount); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding service account %q",
			id,
		)
	}
	return serviceAccount, nil
}

func (u *userStore) GetServiceAccountByName(
	name string,
) (*brignext.ServiceAccount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := u.serviceAccountsCollection.FindOne(
		ctx,
		bson.M{
			"name": name,
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrapf(
			result.Err(),
			"error retrieving service account with name %q",
			name,
		)
	}
	serviceAccount := &brignext.ServiceAccount{}
	if err := result.Decode(serviceAccount); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding service account with name %q",
			name,
		)
	}
	return serviceAccount, nil
}

func (u *userStore) DeleteServiceAccount(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.serviceAccountsCollection.DeleteOne(
			ctx,
			bson.M{
				"id": id,
			},
		); err != nil {
		return errors.Wrapf(err, "error deleting service account %q", id)
	}
	return nil
}

func (u *userStore) DeleteServiceAccountByName(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.serviceAccountsCollection.DeleteOne(
			ctx,
			bson.M{
				"name": name,
			},
		); err != nil {
		return errors.Wrapf(
			err,
			"error deleting service account with name %q",
			name,
		)
	}
	return nil
}
