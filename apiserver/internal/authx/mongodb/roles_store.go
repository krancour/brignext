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

func (r *rolesStore) GrantToUser(
	ctx context.Context,
	userID string,
	roles ...authx.Role,
) error {
	for _, role := range roles {
		if _, err := r.usersCollection.UpdateOne(
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

func (r *rolesStore) RevokeFromUser(
	ctx context.Context,
	userID string,
	roles ...authx.Role,
) error {
	for _, role := range roles {
		if _, err := r.usersCollection.UpdateOne(
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

func (r *rolesStore) GrantToServiceAccount(
	ctx context.Context,
	serviceAccountID string,
	roles ...authx.Role,
) error {
	for _, role := range roles {
		if _, err := r.serviceAccountsCollection.UpdateOne(
			ctx,
			bson.M{
				"id": serviceAccountID,
				"roles": bson.M{
					"$nin": []authx.Role{role},
				},
			},
			bson.M{
				"$push": bson.M{
					"roles": bson.M{
						"$each": []authx.Role{role},
						"$sort": bson.M{
							"name":  1,
							"scope": 1,
						},
					},
				},
			},
		); err != nil {
			return errors.Wrapf(
				err,
				"error updating service account %q",
				serviceAccountID,
			)
		}
	}
	return nil
}

func (r *rolesStore) RevokeFromServiceAccount(
	ctx context.Context,
	serviceAccountID string,
	roles ...authx.Role,
) error {
	for _, role := range roles {
		if _, err := r.serviceAccountsCollection.UpdateOne(
			ctx,
			bson.M{
				"id": serviceAccountID,
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
				"error updating service account %q",
				serviceAccountID,
			)
		}
	}
	return nil
}
