package store

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/ecos/ingestion/internal/model"
)

func TestReadingsStore_MarshalRoundTrip(t *testing.T) {
	original := model.SensorReading{
		ID:          "r-1",
		StationID:   "station-1",
		Timestamp:   "2024-01-15T10:30:00Z",
		ReadingType: model.ReadingTypeTemperature,
		Value:       22.5,
		Unit:        "celsius",
		Quality:     model.Quality{Status: model.QualityRaw, Flags: []string{}},
		Location:    model.Location{Latitude: 40.7128, Longitude: -74.0060, Elevation: 10},
		Metadata:    map[string]string{"source": "test"},
	}

	item, err := attributevalue.MarshalMap(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Verify key attributes exist.
	if item["stationId"] == nil {
		t.Error("expected stationId attribute")
	}
	if item["timestamp"] == nil {
		t.Error("expected timestamp attribute")
	}
	if item["readingType"] == nil {
		t.Error("expected readingType attribute")
	}

	var restored model.SensorReading
	if err := attributevalue.UnmarshalMap(item, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("ID: got %s, want %s", restored.ID, original.ID)
	}
	if restored.StationID != original.StationID {
		t.Errorf("StationID: got %s, want %s", restored.StationID, original.StationID)
	}
	if restored.Value != original.Value {
		t.Errorf("Value: got %f, want %f", restored.Value, original.Value)
	}
	if string(restored.ReadingType) != string(original.ReadingType) {
		t.Errorf("ReadingType: got %s, want %s", restored.ReadingType, original.ReadingType)
	}
	if restored.Location.Latitude != original.Location.Latitude {
		t.Errorf("Latitude: got %f, want %f", restored.Location.Latitude, original.Location.Latitude)
	}
}

func TestApplyTimestampCondition(t *testing.T) {
	base := expression.Key("stationId").Equal(expression.Value("test"))

	// No time filters.
	result := applyTimestampCondition(base, model.ReadingsQuery{})
	_, err := expression.NewBuilder().WithKeyCondition(result).Build()
	if err != nil {
		t.Fatalf("no filters: %v", err)
	}

	// Start time only.
	start := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	result = applyTimestampCondition(base, model.ReadingsQuery{StartTime: &start})
	_, err = expression.NewBuilder().WithKeyCondition(result).Build()
	if err != nil {
		t.Fatalf("start only: %v", err)
	}

	// End time only.
	end := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	result = applyTimestampCondition(base, model.ReadingsQuery{EndTime: &end})
	_, err = expression.NewBuilder().WithKeyCondition(result).Build()
	if err != nil {
		t.Fatalf("end only: %v", err)
	}

	// Both.
	result = applyTimestampCondition(base, model.ReadingsQuery{StartTime: &start, EndTime: &end})
	_, err = expression.NewBuilder().WithKeyCondition(result).Build()
	if err != nil {
		t.Fatalf("both: %v", err)
	}
}

func TestNewReadingsStore(t *testing.T) {
	client := dynamodb.New(dynamodb.Options{Region: "us-east-1"})
	store := NewReadingsStore(client, "test-table", "test-index")
	if store == nil {
		t.Error("expected non-nil store")
	}
}
