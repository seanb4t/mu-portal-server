package database

import (
	"fmt"

	"emperror.dev/errors"
)

// Config holds information necessary for connecting to a database.
type Config struct {
	Host string
	Port int
	User string
	Pass string
	Name string

	Params map[string]string
}

// Validate checks that the configuration is valid.
func (c Config) Validate() error {
	if c.Host == "" {
		return errors.New("database host is required")
	}

	if c.Port == 0 {
		return errors.New("database port is required")
	}

	if c.User == "" {
		return errors.New("database user is required")
	}

	if c.Name == "" {
		return errors.New("database name is required")
	}

	return nil
}

// DSN returns a Mongo driver compatible data source name.
// TODO: replace with mongo support
func (c Config) URI() string {
	var params string

	if len(c.Params) > 0 {
		var query string

		for key, value := range c.Params {
			if query != "" {
				query += "&"
			}

			query += key + "=" + value
		}

		params = "?" + query
	}

	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s%s",
		c.User,
		c.Pass,
		c.Host,
		c.Port,
		c.Name,
		params,
	)
}
