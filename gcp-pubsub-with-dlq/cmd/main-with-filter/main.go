package main

import (
	"bytes"
	"cloud.google.com/go/pubsub"
	"context"
	"flag"
	pubsub_with_dlq "github.com/magnuswahlstrand/blog/gco-pubsub-with-dlq"
	"log"
	"os"
	"os/signal"
)

func main() {
	projectID := "b32-demo-projects"
	subscriptionID := "app.user-created"
	filterFunc := func(msg *pubsub.Message) bool { return false }

	sendToDLQ := flag.Bool("send-to-dlq", true, "send messages starting with \"dead\" to DLQ")
	flag.Parse()
	if *sendToDLQ {
		filterFunc = func(msg *pubsub.Message) bool {
			return bytes.HasPrefix(msg.Data, []byte("dead"))
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err := pubsub_with_dlq.ProcessMessagesWithFilter(ctx, projectID, subscriptionID, filterFunc)
	if err != nil {
		log.Fatalln("error", err)
	}
}
