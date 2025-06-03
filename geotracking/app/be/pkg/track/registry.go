package track

import (
	_ "embed"
	"fmt"
	"log/slog"

	"github.com/hamba/avro"

	"app/pkg/sregistry"
)

type schema struct {
	id      int
	subject string
	value   avro.Schema
}

// TODO: move to parameters
const TopicLocation = "app.track.location"

var locationSchema schema

//go:embed location.avsc
var locationSchemaStr string

func RegisterSchemas(registryURL string) error {
	var err error

	locationSchema.subject = fmt.Sprintf("%s-value", TopicLocation)

	locationSchema.value, err = avro.Parse(locationSchemaStr)
	if err != nil {
		return fmt.Errorf("internal error: failed to parse location.avsc: %w", err)
	}

	locationSchema.id, err = sregistry.RegisterSchema(
		registryURL,
		locationSchema.subject,
		locationSchemaStr,
	)

	if err != nil {
		return fmt.Errorf("error: failed to register schema: %w", err)
	}

	slog.Debug(
		"Registered schema",
		slog.Any("name", TopicLocation),
		slog.Any("version", locationSchema.id),
	)

	return nil
}
