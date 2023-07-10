// Package helloworld provides a set of Cloud Functions samples.
package serverless

import (
	"context"
	"fmt"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"log"
)

func init() {
	functions.CloudEvent("HelloPubSub2", helloPubSub2)
	functions.CloudEvent("HelloPubSubMetadata", helloPubSubMetadata)

	svc := Service{client: nil}
	functions.CloudEvent("HandleMetadataUpdated", svc.HandleMetadataUpdated)
}

type CloudStorageEventData struct {
	Bucket                  string `json:"bucket"`
	ContentType             string `json:"contentType"`
	CRC32C                  string `json:"crc32c"`
	ETag                    string `json:"etag"`
	Generation              string `json:"generation"`
	ID                      string `json:"id"`
	Kind                    string `json:"kind"`
	MD5Hash                 string `json:"md5Hash"`
	MediaLink               string `json:"mediaLink"`
	Metageneration          string `json:"metageneration"`
	Name                    string `json:"name"`
	SelfLink                string `json:"selfLink"`
	Size                    string `json:"size"`
	StorageClass            string `json:"storageClass"`
	TimeCreated             string `json:"timeCreated"`
	TimeStorageClassUpdated string `json:"timeStorageClassUpdated"`
	Updated                 string `json:"updated"`
}

// helloPubSub consumes a CloudEvent message and extracts the Pub/Sub message.
func helloPubSub2(ctx context.Context, e event.Event) error {
	var msg CloudStorageEventData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %w", err)
	}

	log.Print("File uploaded")
	log.Printf("Type %s", e.Type())
	log.Printf("Source %s", e.Source())
	log.Printf("Subject %s", e.Subject())
	log.Printf("ID %s", e.ID())
	log.Printf("Time %s", e.Time())
	log.Printf("%+v", msg)

	log.Print("--------------", msg)
	log.Println("Bucket", msg.Bucket)
	log.Println("Name", msg.Name)
	log.Println("Content Type", msg.ContentType)
	log.Println("Time created", msg.TimeCreated)
	return nil
}

// helloPubSub consumes a CloudEvent message and extracts the Pub/Sub message.
func helloPubSubMetadata(ctx context.Context, e event.Event) error {
	var msg CloudStorageEventData
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %w", err)
	}

	log.Print("File uploaded")
	log.Printf("Type %s", e.Type())
	log.Printf("Source %s", e.Source())
	log.Printf("Subject %s", e.Subject())
	log.Printf("ID %s", e.ID())
	log.Printf("Time %s", e.Time())
	log.Printf("%+v", msg)

	//log.Printf("--------------", msg)
	//log.Println("Bucket", msg.Bucket)
	//log.Println("Name", msg.Name)
	//log.Println("Content Type", msg.ContentType)
	//log.Println("Time created", msg.TimeCreated)
	return nil
}

// helloPubSub consumes a CloudEvent message and extracts the Pub/Sub message.

type MetadataUpdatedEvent struct {
	Metadata map[string]string `json:"metadata"`
}

func (s *Service) HandleMetadataUpdated(ctx context.Context, e event.Event) error {

	// helloPubSub consumes a CloudEvent message and extracts the Pub/Sub message.
	//msg := make(map[string]interface{})
	var msg MetadataUpdatedEvent
	if err := e.DataAs(&msg); err != nil {
		return fmt.Errorf("event.DataAs: %w", err)
	}

	err := s.client.StoreMetadata(msg.Metadata["key1"], msg.Metadata["key2"])
	if err != nil {
		return err
	}

	//metadata := msg["metadata"]
	//if metadata == nil {
	//	log.Println("No metadata")
	//	return nil
	//}
	return nil
}

type Service struct {
	client ApiClient
}

//go:generate moq -out apiclient_moq_test.go . ApiClient
type ApiClient interface {
	StoreMetadata(key1 string, key2 string) error
}
