package mongodb

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const envconfigPrefix = "MONGODB"

// config represents common configuration options for a MongoDB connection
type config struct {
	// TODO: This is getting complicated-- maybe just allow the connection string
	// to be passed in?
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
	connectionString := os.Getenv("MONGODB_CONNECTION_STRING")
	database := os.Getenv("MONGODB_DATABASE")
	if connectionString == "" {
		c := config{}
		err := envconfig.Process(envconfigPrefix, &c)
		if err != nil {
			return nil, errors.Wrap(
				err,
				"error getting mongo configuration from environment",
			)
		}
		connectionString = fmt.Sprintf(
			"mongodb://%s:%s@%s:%d/%s?replicaSet=%s",
			c.Username,
			c.Password,
			c.Host,
			c.Port,
			c.Database,
			c.ReplicaSet,
		)
		database = c.Database
	}

	connectCtx, connectCancel :=
		context.WithTimeout(context.Background(), 10*time.Second)
	defer connectCancel()
	// This client's settings favor consistency over speed
	client, err := mongo.Connect(
		connectCtx,
		options.Client().ApplyURI(connectionString).SetWriteConcern(
			writeconcern.New(writeconcern.WMajority()),
		).SetReadConcern(readconcern.Linearizable()),
	)
	if err != nil {
		return nil, errors.Wrap(err, "error connecting to mongo")
	}
	return client.Database(database), nil
}
