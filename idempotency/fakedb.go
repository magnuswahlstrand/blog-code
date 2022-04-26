package main

import (
	"encoding/json"
	"log"
	"sync"
)

type db2 struct {
	db          map[string][]byte
	mu          sync.Mutex
	shouldSleep bool
}

func (d *db2) get(idempotencyKey string) (RecordedRequestResponse, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()

	data, exists := d.db[idempotencyKey]
	if !exists {
		return RecordedRequestResponse{}, false
	}

	// TODO: handle error
	var record RecordedRequestResponse
	if err := json.Unmarshal(data, &record); err != nil {
		log.Fatalln(err)
	}
	return record, true
}

func (d *db2) update(idempotencyKey string, record RecordedRequestResponse) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// TODO: handle error
	data, err := json.Marshal(record)
	if err != nil {
		log.Fatalln(err)
	}

	d.db[idempotencyKey] = data
}
