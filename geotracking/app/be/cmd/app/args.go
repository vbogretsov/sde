package main

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v3"

	"app/pkg/config"
)

func httpFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "listen-host",
			Aliases:     []string{"l"},
			Value:       "",
			Destination: &cfg.Listen.Host,
			Usage:       "listen host",
			Sources:     cli.EnvVars("LISTEN_HOST"),
		},
		&cli.IntFlag{
			Name:        "listen-port",
			Aliases:     []string{"p"},
			Destination: &cfg.Listen.Port,
			Value:       8000,
			Usage:       "listen host",
			Sources:     cli.EnvVars("LISTEN_PORT"),
		},
	}
}

func kafkaConnectionFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "schema-registry-url",
			Destination: &cfg.SchemaRegistry.URL,
			Required:    true,
			Usage:       "schema registry URL",
			Sources:     cli.EnvVars("SCHEMA_REGISTRY_URL"),
		},

		&cli.StringFlag{
			Name:        "kafka-broker-url",
			Destination: &cfg.Kafka.BrokerURL,
			Required:    true,
			Usage:       "kafka broker URL",
			Sources:     cli.EnvVars("KAFKA_BROKER_URL"),
		},
	}
}

func kafkaTopicsFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.IntFlag{
			Name:        "kafka-topic-partitions",
			Destination: &cfg.Topic.TopicNumPartitions,
			Value:       1,
			Usage:       "number of partitions for kafka topics",
			Sources:     cli.EnvVars("KAFKA_TOPIC_NUM_PARTITIONS"),
		},
		&cli.IntFlag{
			Name:        "kafka-topic-replication-factor",
			Destination: &cfg.Topic.TopicReplicationFactor,
			Value:       1,
			Usage:       "replication factor for kafka topics",
			Sources:     cli.EnvVars("KAFKA_TOPIC_REPLICATION_FACTOR"),
		},
	}
}

func mongoFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "mongo-address",
			Destination: &cfg.Mongo.Address,
			Required:    true,
			Usage:       "mongodb address",
			Sources:     cli.EnvVars("MONGO_ADDRESS"),
		},
		&cli.StringFlag{
			Name:        "mongo-user",
			Destination: &cfg.Mongo.User,
			Required:    true,
			Usage:       "mongodb username",
			Sources:     cli.EnvVars("MONGO_USER"),
		},
		&cli.StringFlag{
			Name:        "mongo-password",
			Destination: &cfg.Mongo.Password,
			Required:    true,
			Usage:       "mongodb password",
			Sources:     cli.EnvVars("MONGO_PASSWORD"),
		},
		&cli.StringFlag{
			Name:        "mongo-database",
			Destination: &cfg.Mongo.Database,
			Required:    true,
			Usage:       "mongodb database",
			Sources:     cli.EnvVars("MONGO_DATABASE"),
		},
		&cli.IntFlag{
			Name:        "mongo-pool-size",
			Destination: &cfg.Mongo.PoolSize,
			Value:       1000,
			Usage:       "mongo pool size",
			Sources:     cli.EnvVars("MONGO_POOL_SIZE"),
		},
		&cli.DurationFlag{
			Name:        "mongo-write-timeout",
			Destination: &cfg.Mongo.WriteTimeout,
			Value:       5 * time.Second,
			Usage:       "mongo write timeout",
			Sources:     cli.EnvVars("MONGO_WRITE_TIMEOUT"),
		},
	}
}

func commonFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Destination: &cfg.Logging.Level,
			Value:       config.LogLevels[1],
			Usage:       fmt.Sprintf("log level %v", config.LogLevels),
			Sources:     cli.EnvVars("LOG_LEVEL"),
			Validator: func(s string) error {
				if _, ok := config.LogLevelsMap[s]; !ok {
					return fmt.Errorf("unexpected log level %s", s)
				}
				return nil
			},
		},
	}
}

func newSvcFlags(cfg *config.Config) []cli.Flag {
	flags := []cli.Flag{}
	flags = append(flags, commonFlags(cfg)...)
	flags = append(flags, httpFlags(cfg)...)
	flags = append(flags, kafkaConnectionFlags(cfg)...)
	flags = append(flags, mongoFlags(cfg)...)
	flags = append(flags, &cli.IntFlag{
		Name:        "kafka-producer-batch-size",
		Destination: &cfg.Kafka.BatchSize,
		Value:       10000,
		Usage:       "kafka producer batch size",
		Sources:     cli.EnvVars("KAFKA_PRODUCER_BATCH_SIZE"),
	})
	flags = append(flags, &cli.DurationFlag{
		Name:        "kafka-producer-batch-timeout",
		Destination: &cfg.Kafka.BatchTimeout,
		Value:       1 * time.Second,
		Usage:       "kafka producer batch size",
		Sources:     cli.EnvVars("KAFKA_PRODUCER_BATCH_TIMEOUT"),
	})
	return flags
}

func newWrkFlags(cfg *config.Config) []cli.Flag {
	flags := []cli.Flag{}
	flags = append(flags, commonFlags(cfg)...)
	flags = append(flags, httpFlags(cfg)...)
	flags = append(flags, kafkaConnectionFlags(cfg)...)
	flags = append(flags, kafkaTopicsFlags(cfg)...)
	flags = append(flags, mongoFlags(cfg)...)
	flags = append(flags, &cli.StringFlag{
		Name:        "kafka-consumer-group-id",
		Destination: &cfg.Kafka.GroupID,
		Value:       "track",
		Usage:       "kafka consumer group id",
		Sources:     cli.EnvVars("KAFKA_CONSUMER_GROUP_ID"),
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "kafka-consumer-batch-size",
		Destination: &cfg.Kafka.BatchSize,
		Value:       10000,
		Usage:       "kafka consumer batch size",
		Sources:     cli.EnvVars("KAFKA_CONSUMER_BATCH_SIZE"),
	})
	flags = append(flags, &cli.DurationFlag{
		Name:        "kafka-consumer-batch-timeout",
		Destination: &cfg.Kafka.BatchTimeout,
		Value:       time.Duration(60 * time.Second),
		Usage:       "kafka consumer batch timeout (seconds)",
		Sources:     cli.EnvVars("KAFKA_CONSUMER_BATCH_TIMEOUT"),
	})
	return flags
}

func newInitFlags(cfg *config.Config) []cli.Flag {
	flags := []cli.Flag{}
	flags = append(flags, commonFlags(cfg)...)
	flags = append(flags, kafkaConnectionFlags(cfg)...)
	flags = append(flags, kafkaTopicsFlags(cfg)...)
	flags = append(flags, mongoFlags(cfg)...)
	flags = append(flags, &cli.IntFlag{
		Name:        "location-ttl",
		Destination: &cfg.LocationTTL,
		Value:       600,
		Usage:       "location TTL seconds",
		Sources:     cli.EnvVars("LOCATION_TTL"),
	})
	return flags
}
