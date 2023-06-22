package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"log"
)

func main() {
	projectID := "b32-demo-projects"
	topicID := "my-topic"

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("pubsub.NewClient: %v", err)
	}
	defer client.Close()
	topic := client.Topic(topicID)

	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("Hello World"),
	})
	if _, err := result.Get(ctx); err != nil {
		log.Fatalln("error", err)
		return
	}
}
