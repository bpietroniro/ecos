package store

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/ecos/ingestion/internal/model"
)

func TestStationsStore_MarshalRoundTrip(t *testing.T) {
	original := model.Station{
		ID:             "s-1",
		Name:           "Weather Station Alpha",
		Type:           model.StationTypeWeather,
		Location:       model.Location{Latitude: 40.7128, Longitude: -74.0060, Elevation: 10},
		Status:         model.StationStatusActive,
		Instruments:    []string{"thermometer", "barometer", "anemometer"},
		LastReportedAt: "2024-01-15T10:30:00Z",
		Metadata:       map[string]string{"region": "northeast"},
	}

	item, err := attributevalue.MarshalMap(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	// Verify key attribute exists.
	if item["id"] == nil {
		t.Error("expected id attribute")
	}

	var restored model.Station
	if err := attributevalue.UnmarshalMap(item, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if restored.ID != original.ID {
		t.Errorf("ID: got %s, want %s", restored.ID, original.ID)
	}
	if restored.Name != original.Name {
		t.Errorf("Name: got %s, want %s", restored.Name, original.Name)
	}
	if string(restored.Type) != string(original.Type) {
		t.Errorf("Type: got %s, want %s", restored.Type, original.Type)
	}
	if string(restored.Status) != string(original.Status) {
		t.Errorf("Status: got %s, want %s", restored.Status, original.Status)
	}
	if len(restored.Instruments) != len(original.Instruments) {
		t.Errorf("Instruments length: got %d, want %d", len(restored.Instruments), len(original.Instruments))
	}
	if restored.LastReportedAt != original.LastReportedAt {
		t.Errorf("LastReportedAt: got %s, want %s", restored.LastReportedAt, original.LastReportedAt)
	}
}

func TestNewStationsStore(t *testing.T) {
	client := dynamodb.New(dynamodb.Options{Region: "us-east-1"})
	store := NewStationsStore(client, "test-table")
	if store == nil {
		t.Error("expected non-nil store")
	}
}
