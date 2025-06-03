package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	slogecho "github.com/samber/slog-echo"
	"github.com/segmentio/kafka-go"
	"github.com/urfave/cli/v3"
	"go.mongodb.org/mongo-driver/v2/mongo/writeconcern"
	"golang.org/x/sync/errgroup"

	"app/pkg/health"
	"app/pkg/track"
)

func getNumPartitions(topic string) (int, error) {
	conn, err := newKafkaAdmin(cfg.Kafka)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to Kafka: %w", err)
	}
	defer conn.Close()

	parts, err := conn.ReadPartitions(topic)
	if err != nil {
		return 0, fmt.Errorf("failed to read topics partitions: %w", err)
	}

	return len(parts), err
}

func runWrk(ctx context.Context, _ *cli.Command) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	g, ctx := errgroup.WithContext(ctx)

	l := slog.New(slog.NewJSONHandler(os.Stdout, cfg.Logging.ToOptions()))
	slog.SetDefault(l)

	e := echo.New()
	e.Use(slogecho.New(l))
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.GET("/healthz", health.Healthz)

	nparts, err := getNumPartitions(track.TopicLocation)
	if err != nil {
		return err
	}

	journal := false
	mongo, err := newMongo(cfg.Mongo, &writeconcern.WriteConcern{W: 0, Journal: &journal})
	if err != nil {
		return err
	}
	defer mongo.Disconnect(ctx)

	consumer := newConsumer(cfg.Kafka)
	defer consumer.Close()

	locations := mongo.Database(cfg.Mongo.Database).Collection(track.CollectionLocations)

	slog.Info("Starting background worker", "cfg", cfg)
	if err := track.RegisterSchemas(cfg.SchemaRegistry.URL); err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)
	
	g.Go(func() error {
		select {
		case sig := <- sigs:
			slog.Warn("received signal", "signal", sig)
			cancel()
			return nil
		case <- ctx.Done():
			slog.Warn("context canceled", "goroutine", "sig")
			return ctx.Err()
		}
	})

	ack := make(chan []kafka.Message, 200000 * cfg.Topic.TopicNumPartitions)
	g.Go(func() error {
		for {
			select {
			case msgs := <-ack:
				commit := make([]kafka.Message, 0, len(msgs))
				for _, msg := range msgs {
					if msg.Topic == "" {
						continue
					}
					commit = append(commit, msg)
				}


				if err := consumer.CommitMessages(ctx, commit...); err != nil {
					slog.Error("failed to commit offsets", slog.Any("err", err))
					cancel()
					return err
				}

				slog.Info("offsets commited")
			case <-ctx.Done():
				slog.Warn("context canceled", "goroutine", "ack")
				return ctx.Err()
			}
		}
	})

	dlq := make(chan kafka.Message, 100)
	g.Go(func() error {
		for {
			select {
			case msg := <-dlq:
				slog.Warn(
					"failed to process message",
					slog.Any("topic", msg.Topic),
					slog.Any("partition", msg.Partition),
					slog.Any("offset", msg.Offset),
				)
			case <-ctx.Done():
				slog.Warn("context canceled", "goroutine", "dlq")
				return ctx.Err()
			}
		}
	})

	in := make(chan kafka.Message, cfg.Kafka.BatchSize)

	tic := time.NewTicker(cfg.Kafka.BatchTimeout)
	defer tic.Stop()

	g.Go(func() error {
		if err := track.Consume(ctx, locations, cfg.Mongo.WriteTimeout, nparts, tic, in, ack, dlq); err != nil {
			slog.Error("consumer failed", "err", err)
			cancel()
			return err
		}
		return nil
	})

	g.Go(func() error {
		slog.Info("starting consumer loop")
		for {
			msg, err := consumer.FetchMessage(ctx)
			if err != nil {
				slog.Error("failed to fetch message", slog.Any("err", err))
				cancel()
				return err
			}

			l.Debug(
				"consumed message",
				slog.Any("topic", msg.Topic),
				slog.Any("partition", msg.Partition),
				slog.Any("offset", msg.Offset),
			)

			select {
			case in <- msg:
			case <-ctx.Done():
				slog.Warn("context canceled", "goroutine", "rcv")
				return ctx.Err()
			}
		}
	})

	slog.Info("background worker exiting gracefully")
	return g.Wait()
}
