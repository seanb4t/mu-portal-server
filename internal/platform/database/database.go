package database

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"emperror.dev/errors"
	mongowrapper "github.com/opencensus-integrations/gomongowrapper"
)

// NewConnector returns a new database connector for the application.
// TODO: Test w/ testcontainers ( https://golang.testcontainers.org )
func NewConnector(config Config) (*mongowrapper.WrappedClient, error) {

	clientOptions := options.Client().ApplyURI(config.URI())
	err := clientOptions.Validate()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client, err := mongowrapper.NewClient(clientOptions)
	if err != nil {
		return nil, errors.WrapIf(err, "failed to create mongo client")
	}

	return client, nil

}

func NewDbPingCheck(dbConnector *mongowrapper.WrappedClient) func(ctx context.Context) (details interface{}, err error) {
	return func(ctx context.Context) (details interface{}, err error) {
		err = dbConnector.Client().Ping(ctx, readpref.Primary())
		if err != nil {
			return nil, errors.WrapIf(err, "failed to ping mongo")
		}
		return nil, nil
	}
}
