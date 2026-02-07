package store

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/ecos/ingestion/internal/model"
)

// StationsStore defines the interface for station data access.
type StationsStore interface {
	Get(ctx context.Context, id string) (*model.Station, error)
	Put(ctx context.Context, station model.Station) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]model.Station, error)
}

type stationsStore struct {
	client    *dynamodb.Client
	tableName string
}

// NewStationsStore creates a StationsStore backed by DynamoDB.
func NewStationsStore(client *dynamodb.Client, tableName string) StationsStore {
	return &stationsStore{
		client:    client,
		tableName: tableName,
	}
}

func (s *stationsStore) Get(ctx context.Context, id string) (*model.Station, error) {
	result, err := s.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get station: %w", err)
	}
	if result.Item == nil {
		return nil, nil
	}

	var station model.Station
	if err := attributevalue.UnmarshalMap(result.Item, &station); err != nil {
		return nil, fmt.Errorf("unmarshal station: %w", err)
	}

	return &station, nil
}

func (s *stationsStore) Put(ctx context.Context, station model.Station) error {
	item, err := attributevalue.MarshalMap(station)
	if err != nil {
		return fmt.Errorf("marshal station: %w", err)
	}

	_, err = s.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.tableName,
		Item:      item,
	})
	if err != nil {
		return fmt.Errorf("put station: %w", err)
	}

	return nil
}

func (s *stationsStore) Delete(ctx context.Context, id string) error {
	_, err := s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: &s.tableName,
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		return fmt.Errorf("delete station: %w", err)
	}

	return nil
}

func (s *stationsStore) List(ctx context.Context) ([]model.Station, error) {
	result, err := s.client.Scan(ctx, &dynamodb.ScanInput{
		TableName: &s.tableName,
	})
	if err != nil {
		return nil, fmt.Errorf("scan stations: %w", err)
	}

	var stations []model.Station
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &stations); err != nil {
		return nil, fmt.Errorf("unmarshal stations: %w", err)
	}

	return stations, nil
}
