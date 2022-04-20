package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/magnuswahlstrand/blog/webhooks-with-websocket/db"
	"gopkg.in/olahol/melody.v1"
	"log"
	"sync"
)

type Service struct {
	mu      sync.RWMutex
	queries *webhooksdb.Queries
	pubsub  *Pubsub
}

func (s *Service) NewWebhookSession(sess *melody.Session, subscriptionID uuid.UUID, ch chan webhooksdb.Webhook) *WebhookSession {
	ackCh := make(chan uuid.UUID)

	list := func(ctx context.Context, limit int32) ([]webhooksdb.Webhook, error) {
		return s.queries.ListWebhooks(ctx, webhooksdb.ListWebhooksParams{
			Limit:          limit,
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
	return webhookSession
}

func (s *Service) handleWebsocketDisconnect(sess *melody.Session) {
	webhookSession := sess.MustGet("service_session").(*WebhookSession)
	subID := sess.MustGet("subscription_id").(uuid.UUID)
	s.pubsub.Unsubscribe(subID)
	webhookSession.shutdown()
}

func (s *Service) handleDBNotifications(payload string) error {
	var ws webhooksdb.Webhook
	if err := json.Unmarshal([]byte(payload), &ws); err != nil {
		fmt.Println("Error processing JSON: ", err)
		return nil // Ignore errors for now
	}

	fmt.Println("received a webhook for subscription", ws.SubscriptionID)
	s.pubsub.Publish(ws.SubscriptionID, ws)
	return nil
}

func (s *Service) handleWebsocketConnect(sess *melody.Session) {
	fmt.Println("New connection")
	subID := sess.MustGet("subscription_id").(uuid.UUID)
	whCh := s.pubsub.Subscribe(subID)
	//s.AddWebhookSession(sess, subID)

	webhookSession := s.NewWebhookSession(sess, subID, whCh)

	sess.Set("service_session", webhookSession)
	go webhookSession.start()
}

func (s *Service) handleWebsocketMessage(sess *melody.Session, received []byte) {
	fmt.Println("Receive message")

	// TODO: handle JSON events
	id, err := uuid.Parse(string(received))
	if err != nil {
		log.Println("error when parsing received message", err)
		return
	}

	// Forward acknowledgement to session
	webhookSession := sess.MustGet("service_session").(*WebhookSession)

	go func() {
		webhookSession.acknowledgedCh <- id
	}()
}
