package sregistry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const contentType = "application/vnd.schemaregistry.v1+json"

type ResponseDTO struct {
	ID int `json:"id"`
}

func RegisterSchema(registryURL, subject, schema string) (int, error) {
	payload := fmt.Sprintf(`{"schema": %q}`, schema)

	res, err := http.Post(
		registryURL+"/subjects/"+subject+"/versions",
		contentType,
		bytes.NewBuffer([]byte(payload)),
	)

	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	var version ResponseDTO
	if err := json.NewDecoder(res.Body).Decode(&version); err != nil {
		return 0, err
	}

	return version.ID, nil
}
