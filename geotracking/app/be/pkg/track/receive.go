package track

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func onReceive(ctx context.Context, dto LocationDTO, producer *kafka.Writer) error {
	dto.UpdatedAt = time.Now().UTC()

	payload, err := serialize(&dto, locationSchema)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Topic: TopicLocation,
		Key:   []byte(dto.UserID),
		Value: payload,
	}

	return producer.WriteMessages(ctx, msg)
}

func onConsume(ctx context.Context, col *mongo.Collection, dto *LocationDTO) error {
	if err := validate.Struct(dto); err != nil {
		return err
	}

	model := dto.toModel()

	filter := bson.M{"_id": dto.UserID}
	opts := options.Replace().SetUpsert(true)

	_, err := col.ReplaceOne(ctx, filter, model, opts)
	if err != nil {
		return fmt.Errorf("Failed to update MongoDB document: %w", err)
	}

	return nil
}

func readLocation(payload []byte) (LocationModel, error) {
	var res LocationModel

	dto, err := deserialize[LocationDTO](payload, locationSchema.value)
	if err != nil {
		slog.Error("failed to deserialize location", "err", err)
		return res, err
	}

	if err := validate.Struct(&dto); err != nil {
		slog.Error("consumed invali message", "err", err)
		return res, err
	}

	return dto.toModel(), nil
}

func copyOffsets(offsets []kafka.Message) []kafka.Message {
	result := []kafka.Message{}
	for _, m := range offsets {
		if m.Topic == "" {
			continue
		}
		result = append(result, m)
	}
	return result
}

func flushLocations(
	ctx context.Context,
	col *mongo.Collection,
	timeout time.Duration,
	ack chan<- []kafka.Message,
	dlq chan<- kafka.Message,
	batchSize int,
	records [][]kafka.Message,
	offsets []kafka.Message,
) error {
	if batchSize == 0 {
		return nil
	}

	var result error

	var sem sync.WaitGroup
	sem.Add(len(records))

	started := time.Now()
	slog.Info("flusing records batch")
	for part, recs := range records {
		if len(recs) == 0 {
			sem.Done()
			continue
		}

		go func(p int, recs []kafka.Message) {
			defer sem.Done()

			log := slog.With("partition", p)
			log.Info("flusing partition batch", slog.Any("len", len(recs)))

			inmap := make([]int, len(recs))
			batch := make([]mongo.WriteModel, 0, len(recs))
			for j, msg := range recs {
				rec, err := readLocation(msg.Value)
				if err != nil {
					log.Error(
						"record validation failed",
						slog.Any("err", err),
						slog.Any("offset", msg.Offset),
					)
					dlq <- msg
					continue
				}

				item := mongo.NewReplaceOneModel().
					SetFilter(bson.M{"_id": rec.UserID}).
					SetReplacement(rec).
					SetUpsert(true)

				inmap[j] = len(batch)
				batch = append(batch, item)
			}

			opts := options.BulkWrite().SetOrdered(false)

			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			_, err := col.BulkWrite(ctx, batch, opts)
			if err != nil {
				bwErr, ok := err.(mongo.BulkWriteException)
				if ok {
					for _, we := range bwErr.WriteErrors {
						msg := recs[inmap[we.Index]]

						slog.Error(
							"failed to upsert record",
							slog.Any("err", we.Message),
							slog.Any("code", we.Code),
							slog.Any("offset", msg.Offset),
						)

						dlq <- msg
					}
				} else {
					result = err
				}
			}

			slog.Info(
				"committed partition batch",
				slog.Any("size", len(batch)),
			)

			clear(records[p])
			records[p] = records[p][:0]
		}(part, recs)
	}

	slog.Info("waiting for all goroutines to perform flush")
	sem.Wait()

	flushed := time.Now()

	select {
	case ack <- copyOffsets(offsets):
	case <-ctx.Done():
		return ctx.Err()
	}

	clear(offsets)
	completed := time.Now()
	slog.Info(
		"all goroutines completed",
		"flushedIn", float64(flushed.Sub(started)) / float64(time.Second),
		"completedIn", float64(completed.Sub(started)) / float64(time.Second),
		"batchSize", batchSize,
	)
	return result
}

func Consume(
	ctx context.Context,
	col *mongo.Collection,
	mongoWriteTimeout time.Duration,
	nparts int,
	tic *time.Ticker,
	in <-chan kafka.Message,
	ack chan<- []kafka.Message,
	dlq chan<- kafka.Message,
) error {
	slog.Debug("starting locations consumer", "channelCapacity", cap(in))

	offsets := make([]kafka.Message, nparts+1)
	records := make([][]kafka.Message, nparts+1)
	for i := range records {
		records[i] = make([]kafka.Message, 0, cap(in))
	}

	batchSize := 0
	for {
		select {
		case msg := <-in:
			records[msg.Partition] = append(records[msg.Partition], msg)
			if offsets[msg.Partition].Offset < msg.Offset {
				offsets[msg.Partition] = msg
			}
			batchSize += 1

			if batchSize < cap(in) {
				continue
			}

			slog.Info("flush by batch size", "batchSize", batchSize)
			if err := flushLocations(ctx, col, mongoWriteTimeout, ack, dlq, batchSize, records, offsets); err != nil {
				return err
			}
			batchSize = 0
		case <-tic.C:
			slog.Info("flush by timeout", "batchSize", batchSize)
			if err := flushLocations(ctx, col, mongoWriteTimeout, ack, dlq, batchSize, records, offsets); err != nil {
				return err
			}
			batchSize = 0
		case <-ctx.Done():
			return ctx.Err()
		}

	}
}

func OnConsume(ctx context.Context, col *mongo.Collection, key, payload []byte) error {
	dto, err := deserialize[LocationDTO](payload, locationSchema.value)
	if err != nil {
		return err
	}

	return onConsume(ctx, col, &dto)
}

func setRoute(ctx context.Context, dto *RouteDTO, col *mongo.Collection) error {
	dto.UpdatedAt = time.Now().UTC()

	filter := bson.M{"_id": dto.UserID}
	opts := options.Replace().SetUpsert(true)

	_, err := col.ReplaceOne(ctx, filter, dto, opts)
	if err != nil {
		return fmt.Errorf("Failed to update MongoDB document: %w", err)
	}

	return nil
}
