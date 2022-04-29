package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var _ db = &db2{}

type db interface {
	get(idempotencyKey string) (RecordedRequestResponse, bool, error)
	update(idempotencyKey string, record RecordedRequestResponse) error
}
type db2 struct {
	db map[string][]byte
	mu sync.Mutex
}

func (d *db2) get(idempotencyKey string) (RecordedRequestResponse, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, exists := d.db[idempotencyKey]
	if !exists {
		return RecordedRequestResponse{}, false, nil
	}

	// TODO: handle error
	var record RecordedRequestResponse
	if err := json.Unmarshal(data, &record); err != nil {
		return RecordedRequestResponse{}, false, errors.New("failed to unmarshal response")
	}
	return record, true, nil
}

func (d *db2) update(idempotencyKey string, record RecordedRequestResponse) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// TODO: handle error
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	d.db[idempotencyKey] = data
	return nil
}
