package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync/atomic"
	"time"
)

// Copied from https://github.com/GoogleCloudPlatform/golang-samples/blob/main/pubsub/subscriptions/sync_pull.go

// [START pubsub_subscriber_sync_pull]
import (
	"cloud.google.com/go/pubsub"
)

func pullMsgsSync(w io.Writer, projectID, subID string) error {

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("pubsub.NewClient: %v", err)
	}
	defer client.Close()

	sub := client.Subscription(subID)

	// Turn on synchronous mode. This makes the subscriber use the Pull RPC rather
	// than the StreamingPull RPC, which is useful for guaranteeing MaxOutstandingMessages,
	// the max number of messages the client will hold in memory at a time.
	sub.ReceiveSettings.Synchronous = true
	sub.ReceiveSettings.MaxOutstandingMessages = 10

	// Receive messages for 10 seconds, which simplifies testing.
	// Comment this out in production, since `Receive` should
	// be used as a long running operation.
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var received int32
	fmt.Fprintln(w, "Waiting for messages")
	err = sub.Receive(ctx, func(_ context.Context, msg *pubsub.Message) {
		fmt.Fprintf(w, "Got message: %q\n", string(msg.Data))
		atomic.AddInt32(&received, 1)
		msg.Ack()
	})
	if err != nil {
		return fmt.Errorf("sub.Receive: %v", err)
	}
	fmt.Fprintf(w, "Received %d messages\n", received)

	return nil
}

func main() {
	projectID := "b32-demo-projects"
	subID := "user-created-sub"
	err := pullMsgsSync(os.Stdout, projectID, subID)
	if err != nil {
		log.Fatalln("error", err)
	}
}
