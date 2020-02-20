package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const envconfigPrefix = "MONGODB"

// config represents common configuration options for a MongoDB connection
type config struct {
	Host       string `envconfig:"HOST" required:"true"`
	Port       int    `envconfig:"PORT" default:"27017"`
	Database   string `envconfig:"DATABASE" required:"true"`
	ReplicaSet string `envconfig:"REPLICA_SET" required:"true"`
	Username   string `envconfig:"USERNAME" required:"true"`
	Password   string `envconfig:"PASSWORD" required:"true"`
}

// Database returns a connection to a MongoDB database specified by environment
// variables
func Database() (*mongo.Database, error) {
	c := config{}
	err := envconfig.Process(envconfigPrefix, &c)
	if err != nil {
		return nil, errors.Wrap(
			err,
			"error getting mongo configuration from environment",
		)
	}

	connectCtx, connectCancel :=
		context.WithTimeout(context.Background(), 10*time.Second)
	defer connectCancel()
	client, err := mongo.Connect(
		connectCtx,
		options.Client().ApplyURI(
			fmt.Sprintf(
				"mongodb://%s:%s@%s:%d/%s?replicaSet=%s",
				c.Username,
				c.Password,
				c.Host,
				c.Port,
				c.Database,
				c.ReplicaSet,
			),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to mongo")
	}

	// Test connection
	pingCtx, pingCancel :=
		context.WithTimeout(context.Background(), 2*time.Second)
	defer pingCancel()
	err = client.Ping(pingCtx, readpref.Primary())
	if err != nil {
		return nil, errors.Wrap(err, "error pinging mongo")
	}

	return client.Database(c.Database), nil
}
