package store

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/ecos/ingestion/internal/model"
)

// ReadingsStore defines the interface for reading data access.
type ReadingsStore interface {
	Put(ctx context.Context, reading model.SensorReading) error
	GetByStation(ctx context.Context, stationID string, q model.ReadingsQuery) ([]model.SensorReading, error)
	GetByReadingType(ctx context.Context, readingType model.ReadingType, q model.ReadingsQuery) ([]model.SensorReading, error)
	List(ctx context.Context, q model.ReadingsQuery) ([]model.SensorReading, error)
}

type readingsStore struct {
	client    *dynamodb.Client
	tableName string
	indexName string
}

// NewReadingsStore creates a ReadingsStore backed by DynamoDB.
func NewReadingsStore(client *dynamodb.Client, tableName, indexName string) ReadingsStore {
	return &readingsStore{
		client:    client,
		tableName: tableName,
		indexName: indexName,
	}
}

func (s *readingsStore) Put(ctx context.Context, reading model.SensorReading) error {
	reading.SortKey = reading.Timestamp + "#" + reading.ID
	item, err := attributevalue.MarshalMap(reading)
	if err != nil {
		return fmt.Errorf("marshal reading: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put reading: %w", err)
	}

	return nil
}

func (s *readingsStore) GetByStation(ctx context.Context, stationID string, q model.ReadingsQuery) ([]model.SensorReading, error) {
	keyCondition := expression.Key("stationId").Equal(expression.Value(stationID))
	keyCondition = applySortKeyCondition(keyCondition, q)

	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, fmt.Errorf("build expression: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 &s.tableName,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false),
	}
	if q.Limit > 0 {
		input.Limit = &q.Limit
	}

	return s.query(ctx, input)
}

func (s *readingsStore) GetByReadingType(ctx context.Context, readingType model.ReadingType, q model.ReadingsQuery) ([]model.SensorReading, error) {
	keyCondition := expression.Key("readingType").Equal(expression.Value(string(readingType)))
	keyCondition = applySortKeyCondition(keyCondition, q)

	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		return nil, fmt.Errorf("build expression: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 &s.tableName,
		IndexName:                 &s.indexName,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ScanIndexForward:          aws.Bool(false),
	}
	if q.Limit > 0 {
		input.Limit = &q.Limit
	}

	return s.query(ctx, input)
}

func (s *readingsStore) List(ctx context.Context, q model.ReadingsQuery) ([]model.SensorReading, error) {
	// If we have a stationId, use a query instead of scan.
	if q.StationID != "" {
		return s.GetByStation(ctx, q.StationID, q)
	}
	// If we have a readingType, use the GSI.
	if q.ReadingType != "" {
		return s.GetByReadingType(ctx, q.ReadingType, q)
	}

	slog.Warn("performing table scan on readings — consider using stationId or readingType filter")

	input := &dynamodb.ScanInput{
		TableName: &s.tableName,
	}
	if q.Limit > 0 {
		input.Limit = &q.Limit
	}

	result, err := s.client.Scan(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("scan readings: %w", err)
	}

	var readings []model.SensorReading
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &readings); err != nil {
		return nil, fmt.Errorf("unmarshal readings: %w", err)
	}

	return readings, nil
}

func (s *readingsStore) query(ctx context.Context, input *dynamodb.QueryInput) ([]model.SensorReading, error) {
	result, err := s.client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("query readings: %w", err)
	}

	var readings []model.SensorReading
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &readings); err != nil {
		return nil, fmt.Errorf("unmarshal readings: %w", err)
	}

	return readings, nil
}

func applySortKeyCondition(kc expression.KeyConditionBuilder, q model.ReadingsQuery) expression.KeyConditionBuilder {
	if q.StartTime != nil && q.EndTime != nil {
		return kc.And(expression.Key("sk").Between(
			expression.Value(q.StartTime.Format("2006-01-02T15:04:05Z07:00")),
			expression.Value(q.EndTime.Format("2006-01-02T15:04:05Z07:00")+"~"),
		))
	}
	if q.StartTime != nil {
		return kc.And(expression.Key("sk").GreaterThanEqual(
			expression.Value(q.StartTime.Format("2006-01-02T15:04:05Z07:00")),
		))
	}
	if q.EndTime != nil {
		return kc.And(expression.Key("sk").LessThanEqual(
			expression.Value(q.EndTime.Format("2006-01-02T15:04:05Z07:00")+"~"),
		))
	}
	return kc
}
