package database

import (
	"go.mongodb.org/mongo-driver/mongo/options"

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
