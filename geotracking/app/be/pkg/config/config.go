package config

import (
	"fmt"
	"log/slog"
	"time"
)

type Logging struct {
	Level string
}

func (l Logging) ToOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: LogLevelsMap[l.Level],
	}
}

type Mongo struct {
	Address      string
	Database     string
	User         string
	Password     string
	PoolSize     int
	WriteTimeout time.Duration
}

type Kafka struct {
	BrokerURL    string
	GroupID      string
	BatchSize    int
	BatchTimeout time.Duration
}

type Topic struct {
	TopicNumPartitions     int
	TopicReplicationFactor int
}

type SchemaRegistry struct {
	URL string
}

type Listen struct {
	Host string
	Port int
}

func (l Listen) Address() string {
	return fmt.Sprintf("%s:%d", l.Host, l.Port)
}

type JWT struct {
	Key string
}

type Config struct {
	Logging        Logging
	Listen         Listen
	Kafka          Kafka
	Topic          Topic
	Mongo          Mongo
	SchemaRegistry SchemaRegistry
	LocationTTL    int
}
