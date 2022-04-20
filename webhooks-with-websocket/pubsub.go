package main

import (
	"fmt"
	"github.com/google/uuid"
	webhooksdb "github.com/magnuswahlstrand/blog/webhooks-with-websocket/db"
	"sync"
)

type Pubsub struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]chan webhooksdb.Webhook
}

func NewPubsub() *Pubsub {
	ps := &Pubsub{}
	ps.sessions = make(map[uuid.UUID]chan webhooksdb.Webhook)
	return ps
}

func (ps *Pubsub) Publish(subID uuid.UUID, w webhooksdb.Webhook) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	whSess := ps.sessions[subID]
	if whSess == nil {
		fmt.Println("no active listener to subscription, dropping")
		return
	}
	ps.sessions[subID] <- w
}

func (ps *Pubsub) Subscribe(subID uuid.UUID) chan webhooksdb.Webhook {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan webhooksdb.Webhook, 1)
	ps.sessions[subID] = ch
	return ch
}

func (ps *Pubsub) Unsubscribe(subID uuid.UUID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	close(ps.sessions[subID])
	delete(ps.sessions, subID)
}
