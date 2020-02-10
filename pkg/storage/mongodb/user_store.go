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
	"go.mongodb.org/mongo-driver/mongo/options"
)

type userStore struct {
	usersCollection           *mongo.Collection
	sessionsCollection        *mongo.Collection
	serviceAccountsCollection *mongo.Collection
}

func NewUserStore(database *mongo.Database) (storage.UserStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	unique := true

	sessionsCollection := database.Collection("sessions")
	if _, err := sessionsCollection.Indexes().CreateMany(
		ctx,
		[]mongo.IndexModel{
			{
				Keys: bson.M{
					"hashedOAuth2State": 1,
				},
				Options: &options.IndexOptions{
					Unique: &unique,
					PartialFilterExpression: bson.M{
						"hashedOAuth2State": bson.M{"exists": true},
					},
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
		return nil, errors.Wrap(err, "error adding indexes to sessions collection")
	}

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
		sessionsCollection:        sessionsCollection,
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

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
	cur, err := u.usersCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving users")
	}

	users := []brignext.User{}
	if err := cur.All(ctx, &users); err != nil {
		return nil, errors.Wrap(err, "error decoding users")
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

	if err := u.DeleteSessionsByUserID(id); err != nil {
		return errors.Wrapf(err, "error deleting sessions for user %q", id)
	}

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

func (u *userStore) CreateSession(session brignext.Session) (string, string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	session.ID = uuid.NewV4().String()
	session.Created = time.Now()

	var oauth2State, hashedOAuth2State string
	if !session.Root {
		oauth2State = crypto.NewToken(30)
		hashedOAuth2State = crypto.ShortSHA("", oauth2State)
	}

	token := crypto.NewToken(256)
	hashedToken := crypto.ShortSHA("", token)

	if _, err := u.sessionsCollection.InsertOne(
		ctx,
		struct {
			brignext.Session  `bson:",inline"`
			HashedOAuth2State string `bson:"hashedOAuth2State,omitempty"`
			HashedToken       string `bson:"hashedToken"`
		}{
			Session:           session,
			HashedOAuth2State: hashedOAuth2State,
			HashedToken:       hashedToken,
		},
	); err != nil {
		return "", "", "", errors.Wrap(err, "error creating new session")
	}

	return session.ID, oauth2State, token, nil
}

func (u *userStore) GetSessionByOAuth2State(
	oauth2State string,
) (brignext.Session, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	session := brignext.Session{}

	result := u.sessionsCollection.FindOne(
		ctx,
		bson.M{"hashedOAuth2State": crypto.ShortSHA("", oauth2State)},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return session, false, nil
	}
	if result.Err() != nil {
		return session, false, errors.Wrap(
			result.Err(),
			"error retrieving session with OAuth2 state [REDACTED]",
		)
	}
	if err := result.Decode(&session); err != nil {
		return session, false, errors.Wrap(
			err,
			"error decoding session with OAuth2 state [REDACTED]",
		)
	}
	return session, true, nil
}

func (u *userStore) GetSessionByToken(
	token string,
) (brignext.Session, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	session := brignext.Session{}

	result := u.sessionsCollection.FindOne(
		ctx,
		bson.M{"hashedToken": crypto.ShortSHA("", token)},
	)
	if result.Err() == mongo.ErrNoDocuments {
		return session, false, nil
	}
	if result.Err() != nil {
		return session, false, errors.Wrap(
			result.Err(),
			"error retrieving session with token [REDACTED]",
		)
	}
	if err := result.Decode(&session); err != nil {
		return session, false, errors.Wrap(
			err,
			"error decoding session with token [REDACTED]",
		)
	}
	return session, true, nil
}

func (u *userStore) AuthenticateSession(sessionID, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.sessionsCollection.UpdateOne(
			ctx,
			bson.M{
				"_id": sessionID,
			},
			bson.M{
				"$set": bson.M{
					"userID":        userID,
					"authenticated": true,
					"expires":       time.Now().Add(time.Hour),
				},
			},
		); err != nil {
		return errors.Wrap(err, "error updating session")
	}
	return nil
}

func (u *userStore) DeleteSession(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.sessionsCollection.DeleteOne(ctx, bson.M{"_id": id}); err != nil {
		return errors.Wrap(err, "error deleting session")
	}
	return nil
}

func (u *userStore) DeleteSessionsByUserID(userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err := u.sessionsCollection.DeleteMany(
		ctx,
		bson.M{"userID": userID},
	); err != nil {
		return errors.Wrapf(err, "error deleting sessions for user %q", userID)
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

	findOptions := options.Find()
	findOptions.SetSort(bson.M{"_id": 1})
	cur, err := u.serviceAccountsCollection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving service accounts")
	}

	serviceAccounts := []brignext.ServiceAccount{}
	if err := cur.All(ctx, &serviceAccounts); err != nil {
		return nil, errors.Wrap(err, "error decoding service accounts")
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

func (u *userStore) LockServiceAccount(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	if _, err :=
		u.serviceAccountsCollection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$set": bson.M{"locked": true},
			},
		); err != nil {
		return errors.Wrapf(err, "error locking service account %q", id)
	}
	return nil
}

func (u *userStore) UnlockServiceAccount(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mongodbTimeout)
	defer cancel()

	token := crypto.NewToken(256)
	hashedToken := crypto.ShortSHA("", token)

	if _, err :=
		u.serviceAccountsCollection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$unset": bson.M{"locked": 1},
				"$set":   bson.M{"hashedToken": hashedToken},
			},
		); err != nil {
		return "", errors.Wrapf(err, "error unlocking service account %q", id)
	}
	return token, nil
}
