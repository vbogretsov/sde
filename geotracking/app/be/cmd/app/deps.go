package main

import (
	"log/slog"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"

	"app/pkg/config"
	"app/pkg/track"
)

func onKafkaWrite(msgs []kafka.Message, err error) {
	if err != nil {
		slog.Error("failed to produce batch of messages", "err", err)
		return
	}
	slog.Info("committed batch of messages", "len", len(msgs))
}

func newMongo(cfg config.Mongo, wc *writeconcern.WriteConcern) (*mongo.Client, error) {
	auth := options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      cfg.User,
		Password:      cfg.Password,
	}
	opts := options.Client().
		ApplyURI(cfg.Address).
		SetAuth(auth).
		SetMaxPoolSize(uint64(cfg.PoolSize)).
		SetWriteConcern(wc)
	return mongo.Connect(opts)
}

func newProducer(cfg config.Kafka) *kafka.Writer {
	brokers := strings.Split(cfg.BrokerURL, ",")
	slog.Info("connecting to kafka brokers", "brokers", brokers)
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Async:        true,
		RequiredAcks: kafka.RequireAll,
		BatchSize:    cfg.BatchSize,
		BatchTimeout: cfg.BatchTimeout,
		Balancer:     &kafka.Hash{},
		MaxAttempts:  100,
		Completion:   onKafkaWrite,
	}
}

func newConsumer(cfg config.Kafka) *kafka.Reader {
	brokers := strings.Split(cfg.BrokerURL, ",")
	slog.Info("connecting to kafka brokers", "brokers", brokers)
	return kafka.NewReader(kafka.ReaderConfig{
		GroupID:        cfg.GroupID,
		Brokers:        brokers,
		Topic:          track.TopicLocation,
		CommitInterval: 2 * time.Second, // Async Commit
		QueueCapacity:  20000,
		MinBytes:       1 << 10,
		MaxBytes:       1 << 22,
		MaxWait:        50 * time.Millisecond,
	})
}

func newKafkaAdmin(cfg config.Kafka) (*kafka.Conn, error) {
	addr := kafka.TCP(strings.Split(cfg.BrokerURL, ",")[0])
	return kafka.Dial(addr.Network(), addr.String())
}
