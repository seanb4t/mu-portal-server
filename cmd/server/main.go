package main

import (
	"go.uber.org/zap"
)

func main() {
	run("dev", "localhost")
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
