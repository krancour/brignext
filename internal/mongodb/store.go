package mongodb

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type BaseStore struct {
	Database *mongo.Database
}

// TODO: CosmosDB doesn't support this and distributed transactions are
// unreliable in authentic MongoDB as well. On both counts, we should trend
// toward a design that doesn't need to use any distributed transactions.
func (b *BaseStore) DoTx(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	// if err := b.Database.Client().UseSession(
	// 	ctx,
	// 	func(sc mongo.SessionContext) error {
	// 		if err := sc.StartTransaction(); err != nil {
	// 			return errors.Wrapf(err, "error starting transaction")
	// 		}
	// 		if err := fn(sc); err != nil {
	// 			return err
	// 		}
	// 		if err := sc.CommitTransaction(sc); err != nil {
	// 			return errors.Wrap(err, "error committing transaction")
	// 		}
	// 		return nil
	// 	},
	// ); err != nil {
	// 	return err
	// }
	// return nil
	return fn(ctx)
}

func (b *BaseStore) CheckHealth(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := b.Database.Client().Ping(
		pingCtx,
		readpref.Primary(),
	); err != nil {
		return errors.Wrap(err, "error pinging mongodb database")
	}
	return nil
}
