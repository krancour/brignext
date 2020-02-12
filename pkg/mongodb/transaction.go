package mongodb

import (
	"context"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func DoTx(
	ctx context.Context,
	database *mongo.Database,
	fn func(mongo.SessionContext) error,
) error {
	if err := database.Client().UseSession(
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
