package serverless

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test(t *testing.T) {
	mock := ApiClientMock{
		StoreMetadataFunc: func(key1 string, key2 string) error {
			assert.Equal(t, "value", key1)
			assert.Equal(t, "bar", key2)
			return nil
		},
	}

	// Arrange
	evt, err := EventWithMetadata(t, map[string]string{
		"key1": "value",
		"key2": "bar",
	})
	svc := Service{client: &mock}

	// Act
	err = svc.HandleMetadataUpdated(context.Background(), evt)
	assert.Nil(t, err)

	// Assert
	assert.Len(t, mock.StoreMetadataCalls(), 1)
}

func EventWithMetadata(t *testing.T, metadataMap map[string]string) (event.Event, error) {
	event := cloudevents.NewEvent()
	err := event.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"metadata": metadataMap,
	})
	assert.Nil(t, err)
	return event, err
}
