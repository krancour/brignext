package mongodb

import (
	"context"

	"github.com/krancour/brignext/pkg/brignext"
	"github.com/krancour/brignext/pkg/crypto"
	"github.com/krancour/brignext/pkg/storage"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type userStore struct {
	usersCollection      *mongo.Collection
	userTokensCollection *mongo.Collection
}

func NewUserStore(database *mongo.Database) storage.UserStore {
	return &userStore{
		usersCollection:      database.Collection("users"),
		userTokensCollection: database.Collection("userTokens"),
	}
}

func (u *userStore) CreateUser(username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	// Prevent duplicate username
	result := u.usersCollection.FindOne(
		ctx,
		bson.M{
			"username": username,
		},
	)
	if result.Err() == nil {
		return errors.Errorf(
			"a user %q already exists",
			username,
		)
	} else if result.Err() != mongo.ErrNoDocuments {
		return errors.Wrapf(
			result.Err(),
			"error checking for existing user %q",
			username,
		)
	}

	if _, err :=
		u.usersCollection.InsertOne(
			ctx,
			bson.M{
				"username":       username,
				"hashedpassword": crypto.ShortSHA(username, password),
			},
		); err != nil {
		return errors.Wrapf(
			err,
			"error creating user %q",
			username,
		)
	}
	return nil
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

func (u *userStore) GetUser(username string) (*brignext.User, error) {
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
			"error retrieving user %q",
			username,
		)
	}
	user := &brignext.User{}
	if err := result.Decode(user); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding user %q",
			username,
		)
	}
	return user, nil
}

func (u *userStore) GetUserByUsernameAndPassword(
	username string,
	password string,
) (*brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := u.usersCollection.FindOne(
		ctx,
		bson.M{
			"username":       username,
			"hashedpassword": crypto.ShortSHA(username, password),
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrapf(
			result.Err(),
			"error retrieving user %q with password [REDACTED]",
			username,
		)
	}
	user := &brignext.User{}
	if err := result.Decode(user); err != nil {
		return nil, errors.Wrapf(
			err,
			"error decoding user %q",
			username,
		)
	}
	return user, nil
}

func (u *userStore) GetUserByToken(token string) (*brignext.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	result := u.userTokensCollection.FindOne(
		ctx,
		bson.M{
			// TODO: What can we use as a salt?
			"hashedtoken": crypto.ShortSHA("", token),
		},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if result.Err() != nil {
		return nil, errors.Wrap(
			result.Err(),
			"error retrieving user token [REDACTED]",
		)
	}
	userToken := &brignext.UserToken{}
	if err := result.Decode(userToken); err != nil {
		return nil, errors.Wrap(
			err,
			"error decoding user token [REDACTED] ",
		)
	}
	return u.GetUser(userToken.Username)
}

func (u *userStore) UpdateUserPassword(username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.usersCollection.UpdateOne(
			ctx,
			bson.M{
				"username": username,
			},
			bson.M{
				"$set": bson.M{
					"hashedpassword": crypto.ShortSHA(username, password),
				},
			},
		); err != nil {
		return errors.Wrapf(
			err,
			"error updating user %q hashed password",
			username,
		)
	}
	return nil
}

func (u *userStore) DeleteUser(username string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.usersCollection.DeleteOne(
			ctx,
			bson.M{
				"username": username,
			},
		); err != nil {
		return errors.Wrapf(
			err,
			"error deleting user %q",
			username,
		)
	}
	return nil
}

func (u *userStore) CreateUserToken(username string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	token := crypto.NewToken()

	if _, err :=
		u.userTokensCollection.InsertOne(
			ctx,
			bson.M{
				"username": username,
				// TODO: What can we use as a salt?
				"hashedtoken": crypto.ShortSHA("", token),
			},
		); err != nil {
		return "", errors.Wrapf(
			err,
			"error creating new token for user %q",
			username,
		)
	}
	return token, nil
}

func (u *userStore) DeleteUserToken(token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.userTokensCollection.DeleteOne(
			ctx,
			bson.M{
				// TODO: What can we use as a salt?
				"hashedtoken": crypto.ShortSHA("", token),
			},
		); err != nil {
		return errors.Wrap(
			err,
			"error deleting user token [REDACTED]",
		)
	}
	return nil
}
