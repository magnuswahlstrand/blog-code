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

	listWebhooks            func(background context.Context) ([]webhooksdb.Webhook, error)
	markWebhookAcknowledged func(background context.Context, id uuid.UUID) (int64, error)

	unackedMessages map[uuid.UUID]bool
}

const maxCapacity = 5

func (s *WebhookSession) start() {
	fmt.Println("starting session")
	webhooks, err := s.listWebhooks(context.TODO())
	if err != nil {
		log.Fatalln(err)
	}

	sendCh := make(chan webhooksdb.Webhook, maxCapacity)
	for _, webhook := range webhooks {
		sendCh <- webhook
	}

	for {
		isEmpty := len(s.unackedMessages) < 1 && len(sendCh) < 1

		fmt.Println("current unackedMessages", len(s.unackedMessages), isEmpty)
		// TODO: handle if client disconnects here
		// TODO: handle if server shutdown
		select {
		case msg0 := <-sendCh:
			if err := s.sendIndentedMessage(msg0.Payload); err != nil {
				fmt.Println("something went wrong:-(", err)
				continue
			}
			s.unackedMessages[msg0.ID] = true
		case msg1 := <-s.acknowledgedCh:

			// Received ack message
			fmt.Println("received", msg1)
			i, _ := s.markWebhookAcknowledged(context.TODO(), msg1)
			if i != 1 {
				fmt.Println("did not update any webhooks :-(")
			}
			// We need to delete here, just in case other process has marked webhook as acked
			delete(s.unackedMessages, msg1)

		// New webhook received
		case newMessage := <-s.newCh:
			if !isEmpty {
				fmt.Println("unprocessed messages is not zero, skip message", newMessage)
				continue
			}

			fmt.Println("received, adding to queue", newMessage)
			sendCh <- newMessage
		}

		//for _, hook := range webhooks {
		//	if err := s.sendIndentedMessage(hook.Payload); err != nil {
		//		fmt.Println("failed something", err)
		//		continue
		//	}
		//}
		//// We have now parsed all the webhooks in the unackedMessages, we accept all new webhooks
		//
		//for hook := range s.newCh {
		//	fmt.Println("received message")
		//	if err := s.sendIndentedMessage(hook.Payload); err != nil {
		//		fmt.Println("failed something", err)
		//		continue
		//	}
		//}
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
