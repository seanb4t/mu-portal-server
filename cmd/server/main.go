package main

import (
  "github.com/pkg/errors"
	"go.uber.org/zap"
)

func main() {
  run(nil, nil)
}

func run(env string, address string) (<-chan error, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	defer logger.Sync()

	logger.Info("Starting server", zap.String("address", address))

	return nil, nil
}
