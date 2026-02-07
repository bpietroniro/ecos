package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"

	"github.com/ecos/ingestion/internal/model"
)

// Publisher defines the interface for publishing events.
type Publisher interface {
	PublishDataEvent(ctx context.Context, event model.DataEvent) error
	PublishAlertEvent(ctx context.Context, event model.AlertEvent) error
}

type snsPublisher struct {
	client              *sns.Client
	dataIngestionTopicARN string
	alertEventsTopicARN   string
}

// NewSNSClient creates an SNS client. If endpoint is non-empty, it overrides
// the default endpoint (for local development).
func NewSNSClient(ctx context.Context, region, endpoint string) (*sns.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	opts := []func(*sns.Options){}
	if endpoint != "" {
		opts = append(opts, func(o *sns.Options) {
			o.BaseEndpoint = aws.String(endpoint)
		})
	}

	return sns.NewFromConfig(cfg, opts...), nil
}

// NewPublisher creates a Publisher backed by SNS. If topic ARNs are empty,
// returns a no-op publisher that logs warnings.
func NewPublisher(client *sns.Client, dataIngestionTopicARN, alertEventsTopicARN string) Publisher {
	if dataIngestionTopicARN == "" && alertEventsTopicARN == "" {
		slog.Warn("SNS topic ARNs not configured — events will be logged but not published")
		return &noopPublisher{}
	}

	return &snsPublisher{
		client:                client,
		dataIngestionTopicARN: dataIngestionTopicARN,
		alertEventsTopicARN:   alertEventsTopicARN,
	}
}

func (p *snsPublisher) PublishDataEvent(ctx context.Context, event model.DataEvent) error {
	if p.dataIngestionTopicARN == "" {
		slog.Warn("data ingestion topic ARN not configured, skipping publish", "eventType", event.EventType)
		return nil
	}

	return p.publish(ctx, p.dataIngestionTopicARN, event.EventType, event)
}

func (p *snsPublisher) PublishAlertEvent(ctx context.Context, event model.AlertEvent) error {
	if p.alertEventsTopicARN == "" {
		slog.Warn("alert events topic ARN not configured, skipping publish", "eventType", event.EventType)
		return nil
	}

	return p.publish(ctx, p.alertEventsTopicARN, event.EventType, event)
}

func (p *snsPublisher) publish(ctx context.Context, topicARN string, eventType model.EventType, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	_, err = p.client.Publish(ctx, &sns.PublishInput{
		TopicArn: &topicARN,
		Message:  aws.String(string(body)),
		MessageAttributes: map[string]snstypes.MessageAttributeValue{
			"eventType": {
				DataType:    aws.String("String"),
				StringValue: aws.String(string(eventType)),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("publish to %s: %w", topicARN, err)
	}

	slog.Info("published event", "topicARN", topicARN, "eventType", eventType)
	return nil
}

// noopPublisher logs events but does not publish them.
type noopPublisher struct{}

func (p *noopPublisher) PublishDataEvent(_ context.Context, event model.DataEvent) error {
	slog.Warn("noop: would publish data event", "eventType", event.EventType, "readingId", event.Reading.ID)
	return nil
}

func (p *noopPublisher) PublishAlertEvent(_ context.Context, event model.AlertEvent) error {
	slog.Warn("noop: would publish alert event", "eventType", event.EventType, "stationId", event.StationID)
	return nil
}
