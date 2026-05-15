package rabbitmq

import (
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestNewJSONPublishingUsesPersistentDeliveryMode(t *testing.T) {
	message := []byte(`{"operation":"deposit"}`)

	publishing := newJSONPublishing(message, true)

	if publishing.ContentType != "application/json" {
		t.Fatalf("expected JSON content type, got %q", publishing.ContentType)
	}
	if publishing.DeliveryMode != amqp.Persistent {
		t.Fatalf("expected persistent delivery mode, got %d", publishing.DeliveryMode)
	}
	if string(publishing.Body) != string(message) {
		t.Fatalf("expected body %q, got %q", message, publishing.Body)
	}
}

func TestNewJSONPublishingCanUseTransientDeliveryMode(t *testing.T) {
	publishing := newJSONPublishing([]byte("{}"), false)

	if publishing.DeliveryMode != amqp.Transient {
		t.Fatalf("expected transient delivery mode, got %d", publishing.DeliveryMode)
	}
}
