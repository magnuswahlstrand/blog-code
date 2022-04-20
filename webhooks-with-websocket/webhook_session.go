package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/magnuswahlstrand/blog/webhooks-with-websocket/db"
	"gopkg.in/olahol/melody.v1"
	"log"
)

type WebhookSession struct {
	websocket      *melody.Session
	newCh          chan webhooksdb.Webhook
	acknowledgedCh chan uuid.UUID

	listWebhooks            func(background context.Context, count int32) ([]webhooksdb.Webhook, error)
	markWebhookAcknowledged func(background context.Context, id uuid.UUID) (int64, error)

	unackedMessages map[uuid.UUID]bool
}

const maxCapacity = 5

func (s *WebhookSession) start() {
	fmt.Println("starting session")
	webhooks, err := s.listWebhooks(context.TODO(), maxCapacity)
	if err != nil {
		log.Fatalln(err)
	}

	sendCh := make(chan webhooksdb.Webhook, maxCapacity)
	for _, webhook := range webhooks {
		sendCh <- webhook
	}

	// First send all outstanding messages from the database

	// Listen to new messages

	for {
		isEmpty := len(s.unackedMessages) < 1 && len(sendCh) < 1

		fmt.Println("current unackedMessages", len(s.unackedMessages), isEmpty)
		// TODO: handle if client disconnects here
		// TODO: handle if server shutdown
		select {
		case wh := <-sendCh:
			if err := s.sendIndentedMessage(wh.Payload); err != nil {
				fmt.Println("something went wrong:-(", err)
				continue
			}
			s.unackedMessages[wh.ID] = true
		case ackID := <-s.acknowledgedCh:

			// Received ack message
			fmt.Println("received", ackID)
			i, _ := s.markWebhookAcknowledged(context.TODO(), ackID)
			if i != 1 {
				fmt.Println("did not update any webhooks :-(")
			}
			// We need to delete here, just in case other process has marked webhook as acked
			delete(s.unackedMessages, ackID)

		// New webhook received
		case wh := <-s.newCh:
			if !isEmpty {
				fmt.Println("unprocessed messages is not zero, skip message", wh)
				continue
			}

			fmt.Println("received, adding to queue", wh)
			sendCh <- wh
		}
	}
}

func (s *WebhookSession) sendIndentedMessage(payload json.RawMessage) error {
	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	if err := s.websocket.Write(b); err != nil {
		return err
	}
	return nil
}

func (s *WebhookSession) shutdown() {
	// TODO: Do proper shutdown here
	fmt.Println("some shutdown here")
}
