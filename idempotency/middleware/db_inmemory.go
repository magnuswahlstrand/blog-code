package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
)

var _ db = &inMemoryDB{}

type inMemoryDB struct {
	db map[string][]byte
	mu sync.Mutex
}

func (d *inMemoryDB) get(ctx context.Context, idempotencyKey string) (RequestRecording, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, exists := d.db[idempotencyKey]
	if !exists {
		return RequestRecording{}, false, nil
	}

	// TODO: handle error
	var record RequestRecording
	if err := json.Unmarshal(data, &record); err != nil {
		return RequestRecording{}, false, errors.New("failed to unmarshal response")
	}
	return record, true, nil
}

func (d *inMemoryDB) update(ctx context.Context, idempotencyKey string, record RequestRecording) error {
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

func NewInMemoryDB() *inMemoryDB {
	return &inMemoryDB{db: map[string][]byte{}}
}
