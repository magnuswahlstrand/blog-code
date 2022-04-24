package gcp_pubsub_with_dlq

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"fmt"
)

func ProcessMessagesWithFilter(ctx context.Context, projectID, subID string, filterFunc func(msg *pubsub.Message) bool) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	subscription := client.Subscription(subID)
	return subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		txt := fmt.Sprintf("Received message: %q (attempt %d)", string(msg.Data), *msg.DeliveryAttempt)

		if filterFunc(msg) {
			fmt.Println(txt, " - NACK")
			msg.Nack()
			return
		}

		fmt.Println(txt, " - ACK")
		msg.Ack()
	})
}

// Inspired by https://github.com/GoogleCloudPlatform/golang-samples/blob/main/pubsub/subscriptions/sync_pull.go
func ProcessMessages(ctx context.Context, projectID, subID string) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	subscription := client.Subscription(subID)
	return subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		txt := fmt.Sprintf("Received message: %q (attempt %d)", string(msg.Data), *msg.DeliveryAttempt)

		if bytes.HasPrefix(msg.Data, []byte("dead")) {
			fmt.Println(txt, " - NACK")
			msg.Nack()
			return
		}

		fmt.Println(txt, " - ACK")
		msg.Ack()
	})
}

func RepublishMessages(ctx context.Context, projectID, dlqSubID, toTopicID string) error {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	subscription := client.Subscription(dlqSubID)
	topic := client.Topic(toTopicID)

	return subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		// Republish message
		result := topic.Publish(ctx, msg)
		if _, err := result.Get(ctx); err != nil {
			msg.Nack()
			return
		}

		fmt.Printf("Republished message: %q\n", string(msg.Data))
		msg.Ack()
	})
}
