package track

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/hamba/avro"
)

func serialize(dto any, schema schema) ([]byte, error) {
	payload, err := avro.Marshal(schema.value, dto)
	if err != nil {
		return nil, fmt.Errorf("Failed to serialize: %w", err)
	}

	var buf bytes.Buffer
	buf.WriteByte(0)
	if err := binary.Write(&buf, binary.BigEndian, int32(schema.id)); err != nil {
		return nil, fmt.Errorf("Failed to write schema ID: %w", err)
	}

	buf.Write(payload)
	return buf.Bytes(), nil
}

func deserialize[T any](payload []byte, schema avro.Schema) (T, error) {
	var dto T
	if err := avro.Unmarshal(schema, payload[5:], &dto); err != nil {
		return dto, fmt.Errorf("Failed to deserialize: %w", err)
	}
	return dto, nil
}
