package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/magnuswahlstrand/blog/webhooks-with-websocket/db"
	"gopkg.in/olahol/melody.v1"
	"sync"
	"time"
)

type Service struct {
	mu       sync.RWMutex
	sessions map[uuid.UUID]*WebhookSession
	queries  *webhooksdb.Queries
}

func (s *Service) AddWebhookSession(sess *melody.Session, subscriptionID uuid.UUID) {
	ch := make(chan webhooksdb.Webhook)
	ackCh := make(chan uuid.UUID)

	list := func(ctx context.Context) ([]webhooksdb.Webhook, error) {
		return s.queries.ListWebhooks(ctx, webhooksdb.ListWebhooksParams{
			Limit:          5,
			SubscriptionID: subscriptionID,
		})
	}
	markAcknowledged := func(ctx context.Context, id uuid.UUID) (int64, error) {
		return s.queries.SetPublished(ctx, webhooksdb.SetPublishedParams{
			ID:             id,
			SubscriptionID: subscriptionID,
		})
	}
	webhookSession := &WebhookSession{
		websocket:      sess,
		newCh:          ch,
		acknowledgedCh: ackCh,

		listWebhooks:            list,
		markWebhookAcknowledged: markAcknowledged,

		unackedMessages: map[uuid.UUID]bool{},
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: check if session already exists
	s.sessions[subscriptionID] = webhookSession

	go webhookSession.start()
}

func (s *Service) RemoveWebhookSession(subscriptionID uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: check if session already exists
	delete(s.sessions, subscriptionID)

	// TODO: Shutdown previous session
	//go webhookSession.start()
}

func (s *Service) ForwardAck(id uuid.UUID, subscriptionID uuid.UUID) {
	fmt.Println("forwarding webhook")
	s.mu.Lock()
	defer s.mu.Unlock()

	whSess := s.sessions[subscriptionID]
	if whSess == nil {
		fmt.Println("no active listener to subscription, dropping")
		return
	}
	go func() {
		whSess.acknowledgedCh <- id
	}()
}

func (s *Service) ForwardWebhook(ws webhooksdb.Webhook) {
	fmt.Println("forwarding webhook")
	s.mu.Lock()
	defer s.mu.Unlock()

	whSess := s.sessions[ws.SubscriptionID]
	if whSess == nil {
		fmt.Println("no active listener to subscription, dropping")
		return
	}

	go func() {
		whSess.newCh <- ws
	}()
	fmt.Println("webhook forwarded webhook")
}

func (s *Service) waitForNotification(l *pq.Listener) {
	for {
		select {
		case n := <-l.Notify:
			fmt.Println("Received data from channel [", n.Channel, "] :")
			// Prepare notification payload for pretty print
			var prettyJSON bytes.Buffer
			err := json.Indent(&prettyJSON, []byte(n.Extra), "", "\t")
			if err != nil {
				fmt.Println("Error processing JSON: ", err)
				return
			}

			fmt.Println(string(n.Extra))
			var ws webhooksdb.Webhook
			if err := json.Unmarshal([]byte(n.Extra), &ws); err != nil {
				fmt.Println("Error processing JSON: ", err)
				return
			}

			fmt.Println(string(prettyJSON.Bytes()))

			fmt.Println("received a webhook for subscription", ws.SubscriptionID)
			s.ForwardWebhook(ws)

			return
		case <-time.After(90 * time.Second):
			fmt.Println("Received no events for 90 seconds, checking connection")
			go func() {
				l.Ping()
			}()
			return
		}
	}
}
