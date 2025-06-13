package main

import (
	"errors"
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
			Destination: &cfg.Service.Listen.Host,
			Usage:       "listen host",
			Sources:     cli.EnvVars("LISTEN_HOST"),
		},
		&cli.IntFlag{
			Name:        "listen-port",
			Aliases:     []string{"p"},
			Destination: &cfg.Service.Listen.Port,
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
		Destination: &cfg.Kafka.Producer.BatchSize,
		Value:       2000,
		Usage:       "kafka producer batch size",
		Sources:     cli.EnvVars("KAFKA_PRODUCER_BATCH_SIZE"),
	})
	flags = append(flags, &cli.DurationFlag{
		Name:        "kafka-producer-batch-timeout",
		Destination: &cfg.Kafka.Producer.BatchTimeout,
		Value:       1 * time.Second,
		Usage:       "kafka producer batch size",
		Sources:     cli.EnvVars("KAFKA_PRODUCER_BATCH_TIMEOUT"),
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "kafka-producer-max-attempts",
		Destination: &cfg.Kafka.Producer.MaxAttempts,
		Value:       100,
		Usage:       "kafka producer batch send max attempts",
		Sources:     cli.EnvVars("KAFKA_PRODUCER_MAX_ATTEMPTS"),
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
		Destination: &cfg.Kafka.Consumer.GroupID,
		Value:       "track",
		Usage:       "kafka consumer group id",
		Sources:     cli.EnvVars("KAFKA_CONSUMER_GROUP_ID"),
	})
	flags = append(flags, &cli.DurationFlag{
		Name:        "kafka-consumer-max-wait",
		Destination: &cfg.Kafka.Consumer.MaxWait,
		Value:       time.Duration(100 * time.Millisecond),
		Usage:       "maximum amount of time to wait for new data to come when fetching batches of messages from kafka",
		Sources:     cli.EnvVars("KAFKA_CONSUMER_MAX_WAIT"),
	})
	flags = append(flags, &cli.DurationFlag{
		Name:        "kafka-consumer-commit-interval",
		Destination: &cfg.Kafka.Consumer.CommitInterval,
		Value:       0,
		Usage:       "commit interval indicates the interval at which offsets are committed to the broker,  if 0 commits will be handled synchronously",
		Sources:     cli.EnvVars("KAFKA_CONSUMER_COMMIT_INTERVAL"),
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "kafka-consumer-max-bytes",
		Destination: &cfg.Kafka.Consumer.MaxBytes,
		Value:       1 << 22,
		Usage:       "max bytes indicates to the broker the maximum batch size that the consumer will accept",
		Sources:     cli.EnvVars("KAFKA_CONSUMER_MAX_BYTES"),
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "kafka-consumer-queue-capacity",
		Destination: &cfg.Kafka.Consumer.QueueCapacity,
		Value:       20000,
		Usage:       "the capacity of the internal message queue",
		Sources:     cli.EnvVars("KAFKA_CONSUMER_MAX_BYTES"),
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "flush-batch-size",
		Destination: &cfg.Worker.FlushBatchSize,
		Value:       2000,
		Usage:       "flush batch size",
		Sources:     cli.EnvVars("FLUSH_BATCH_SIZE"),
	})
	flags = append(flags, &cli.DurationFlag{
		Name:        "flush-batch-timeout",
		Destination: &cfg.Worker.FlushBatchTimeout,
		Value:       time.Duration(1 * time.Second),
		Usage:       "flush batch timeout (seconds)",
		Sources:     cli.EnvVars("FLUSH_BATCH_TIMEOUT"),
		Validator: func(d time.Duration) error {
			if d == 0 {
				return errors.New("duration must be positive")
			}
			return nil
		},
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "dlq-buffer-size",
		Destination: &cfg.Worker.DLQBufferSize,
		Value:       2000,
		Usage:       "dead letter queue size",
		Sources:     cli.EnvVars("DQL_QUEUE_SIZE"),
	})
	flags = append(flags, &cli.IntFlag{
		Name:        "dlq-buffer-size",
		Destination: &cfg.Worker.DLQBufferSize,
		Value:       100,
		Usage:       "ack queue size per partition",
		Sources:     cli.EnvVars("ACK_QUEUE_SIZE"),
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
