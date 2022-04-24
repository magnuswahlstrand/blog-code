package main

import (
	"context"
	pubsub_with_dlq "github.com/magnuswahlstrand/blog/gco-pubsub-with-dlq"
	"log"
	"os"
	"os/signal"
)

func main() {
	projectID := "b32-demo-projects"
	dlqSubscriptionID := "app.user-created.dlq"
	targetTopic := "user-created"

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := pubsub_with_dlq.RepublishMessages(ctx, projectID, dlqSubscriptionID, targetTopic)
	if err != nil {
		log.Fatalln("error", err)
	}
}
