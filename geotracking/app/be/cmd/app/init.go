package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"

	"app/pkg/track"
)

func runInit(ctx context.Context, _ *cli.Command) error {
	l := slog.New(slog.NewJSONHandler(os.Stdout, cfg.Logging.ToOptions()))
	slog.SetDefault(l)

	slog.Info("Starting initialization")
	slog.Debug("Application configuration", "cfg", cfg)

	mongodb, err := newMongo(cfg.Mongo, &writeconcern.WriteConcern{W: "majority"})
	if err != nil {
		slog.Error("Failed to connect to MongoDB", "err", err)
		return err
	}
	defer mongodb.Disconnect(ctx)

	kafkacn, err := newKafkaAdmin(cfg.Kafka)
	if err != nil {
		slog.Error("Failed to connect to Kafka Cluster", "err", err)
		return err
	}
	defer kafkacn.Close()

	slog.Info("Creating Kafka Topics")
	if err := track.CreateTopics(kafkacn, cfg.Topic); err != nil {
		slog.Error("Failed to create Kafka Topics", "err", err)
		return err
	}

	slog.Info("Initializing MongoDB Collections")
	if err := track.CreateLocationIndexes(ctx, cfg, mongodb); err != nil {
		slog.Error("Failed to initialize MongoDB collections", "err", err)
		return err
	}

	return nil
}
