package mongodb

import (
	"context"

	"github.com/brigadecore/brigade/v2/apiserver/internal/authx"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type rolesStore struct {
	usersCollection           *mongo.Collection
	serviceAccountsCollection *mongo.Collection
}

func NewRolesStore(database *mongo.Database) (authx.RolesStore, error) {
	return &rolesStore{
		usersCollection:           database.Collection("users"),
		serviceAccountsCollection: database.Collection("service-accounts"),
	}, nil
}

func (r *rolesStore) Grant(
	ctx context.Context,
	principalType authx.PrincipalType,
	principalID string,
	roles ...authx.Role,
) error {
	var collection *mongo.Collection
	if principalType == authx.PrincipalTypeUser {
		collection = r.usersCollection
	} else if principalType == authx.PrincipalTypeServiceAccount {
		collection = r.serviceAccountsCollection
	} else {
		return nil
	}
	for _, role := range roles {
		if _, err := collection.UpdateOne(
			ctx,
			bson.M{
				"id": principalID,
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
			return errors.Wrapf(
				err,
				"error updating user %s %q",
				principalType,
				principalID,
			)
		}
	}
	return nil
}

func (r *rolesStore) Revoke(
	ctx context.Context,
	principalType authx.PrincipalType,
	principalID string,
	roles ...authx.Role,
) error {
	var collection *mongo.Collection
	if principalType == authx.PrincipalTypeUser {
		collection = r.usersCollection
	} else if principalType == authx.PrincipalTypeServiceAccount {
		collection = r.serviceAccountsCollection
	} else {
		return nil
	}
	for _, role := range roles {
		if _, err := collection.UpdateOne(
			ctx,
			bson.M{
				"id": principalID,
			},
			bson.M{
				"$pull": bson.M{
					"roles": bson.M{
						"$in": []authx.Role{role},
					},
				},
			},
		); err != nil {
			return errors.Wrapf(
				err,
				"error updating user %s %q",
				principalType,
				principalID,
			)
		}
	}
	return nil
}
