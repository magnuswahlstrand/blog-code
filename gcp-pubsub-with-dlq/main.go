package main

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
)

// Inspired by https://github.com/GoogleCloudPlatform/golang-samples/blob/main/pubsub/subscriptions/sync_pull.go
func processMessages(ctx context.Context, projectID, subID string) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	fmt.Println("Waiting for messages")
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		fmt.Printf("Got message: %q with %d delivery attempts\n", string(msg.Data), *msg.DeliveryAttempt)

		if bytes.HasPrefix(msg.Data, []byte("DEAD-LETTER")) {
			msg.Nack()
		} else {
			msg.Ack()
		}

	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %v", err)
	}
	return nil
}

func main() {
	projectID := "b32-demo-projects"
	subID := "app.user-created"

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := processMessages(ctx, projectID, subID)
	if err != nil {
		log.Fatalln("error", err)
	}
}
