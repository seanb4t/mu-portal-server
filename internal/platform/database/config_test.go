package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(t *testing.T) {
	tests := map[string]Config{
		"database host is required": {
			User: "root",
			Pass: "",
			Name: "database",
		},
		"database port is required": {
			Host: "localhost",
			User: "root",
			Pass: "",
			Name: "database",
		},
		"database user is required": {
			Host: "localhost",
			Pass: "",
			Name: "database",
		},
		"database name is required": {
			Host: "localhost",
			User: "root",
			Pass: "",
		},
	}

	for name, test := range tests {
		name, test := name, test

		t.Run(name, func(t *testing.T) {
			err := test.Validate()

			assert.EqualError(t, err, name)
		})
	}
}

func TestConfig_URI(t *testing.T) {
	config := Config{
		Host: "host",
		User: "root",
		Pass: "pass",
		Name: "database",
		Params: map[string]string{
			"parseTime": "true",
		},
	}

	uri := config.URI()

	assert.Equal(t, "mongodb+srv://root:pass@host/database?parseTime=true&retryWrites=true&w=majority", uri)
}
