package track

import (
	"context"
	"log/slog"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"app/pkg/config"
)

const (
	CollectionRoutes    = "track_routes"
	CollectionLocations = "track_locations"
)

func CreateTopics(con *kafka.Conn, cfg config.Topic) error {
	locationTopic := kafka.TopicConfig{
		Topic:             TopicLocation,
		NumPartitions:     cfg.TopicNumPartitions,
		ReplicationFactor: cfg.TopicReplicationFactor,
	}

	l := slog.With(
		slog.Any("topic", locationTopic),
		slog.Any("partitions", cfg.TopicNumPartitions),
		slog.Any("replication.factor", cfg.TopicReplicationFactor),
	)

	l.Info("Creating Kafka Topic")
	if err := con.CreateTopics(locationTopic); err != nil {
		l.Error("Failed to create Kafka Topic", "err", err)
		return err
	}

	return nil
}

func CreateLocationIndexes(ctx context.Context, cfg config.Config, client *mongo.Client) error {
	coll := client.Database(cfg.Mongo.Database).Collection(CollectionLocations)

	slog.Info("Creating MongoDB collection", slog.Any("name", CollectionLocations))
	models := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "lat", Value: 1}, {Key: "lng", Value: 1}},
			Options: options.Index().
				SetName("idx_lat"),
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: 1}},
			Options: options.Index().
				SetExpireAfterSeconds(int32(cfg.LocationTTL)).
				SetName("idx_ttl_timestamp"),
		},
	}

	if _, err := coll.Indexes().CreateMany(ctx, models); err != nil {
		slog.Error("Failed to create MongoDB collection", "err", err)
		return err
	}

	return nil
}
