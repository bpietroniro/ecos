package events

import (
	"context"
	"testing"

	"github.com/ecos/ingestion/internal/model"
)

func TestNoopPublisher_DataEvent(t *testing.T) {
	pub := &noopPublisher{}

	event := model.NewDataEvent(model.EventNewDataReceived, model.SensorReading{
		ID:        "r-1",
		StationID: "station-1",
	}, "test data event")

	err := pub.PublishDataEvent(context.Background(), event)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNoopPublisher_AlertEvent(t *testing.T) {
	pub := &noopPublisher{}

	event := model.NewAlertEvent(model.EventExtremeCondition, "station-1", "test alert", nil)

	err := pub.PublishAlertEvent(context.Background(), event)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNewPublisher_NoopWhenNoARNs(t *testing.T) {
	pub := NewPublisher(nil, "", "")
	if _, ok := pub.(*noopPublisher); !ok {
		t.Error("expected noopPublisher when no ARNs configured")
	}
}

func TestNewPublisher_SNSWhenARNsProvided(t *testing.T) {
	// Pass nil client — we won't actually publish in this test.
	pub := NewPublisher(nil, "arn:aws:sns:us-east-1:123456789:test-topic", "arn:aws:sns:us-east-1:123456789:alert-topic")
	if _, ok := pub.(*snsPublisher); !ok {
		t.Error("expected snsPublisher when ARNs configured")
	}
}
