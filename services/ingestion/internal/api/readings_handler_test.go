package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/ecos/ingestion/internal/model"
)

// mockReadingsStore implements store.ReadingsStore for testing.
type mockReadingsStore struct {
	readings []model.SensorReading
	putErr   error
}

func (m *mockReadingsStore) Put(_ context.Context, reading model.SensorReading) error {
	if m.putErr != nil {
		return m.putErr
	}
	m.readings = append(m.readings, reading)
	return nil
}

func (m *mockReadingsStore) GetByStation(_ context.Context, stationID string, _ model.ReadingsQuery) ([]model.SensorReading, error) {
	var result []model.SensorReading
	for _, r := range m.readings {
		if r.StationID == stationID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockReadingsStore) GetByReadingType(_ context.Context, rt model.ReadingType, _ model.ReadingsQuery) ([]model.SensorReading, error) {
	var result []model.SensorReading
	for _, r := range m.readings {
		if r.ReadingType == rt {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *mockReadingsStore) List(_ context.Context, _ model.ReadingsQuery) ([]model.SensorReading, error) {
	return m.readings, nil
}

// mockPublisher implements events.Publisher for testing.
type mockPublisher struct {
	dataEvents  []model.DataEvent
	alertEvents []model.AlertEvent
}

func (m *mockPublisher) PublishDataEvent(_ context.Context, event model.DataEvent) error {
	m.dataEvents = append(m.dataEvents, event)
	return nil
}

func (m *mockPublisher) PublishAlertEvent(_ context.Context, event model.AlertEvent) error {
	m.alertEvents = append(m.alertEvents, event)
	return nil
}

func TestReadingsHandler_Create(t *testing.T) {
	store := &mockReadingsStore{}
	pub := &mockPublisher{}
	handler := NewReadingsHandler(store, pub)

	body, _ := json.Marshal(model.CreateReadingRequest{
		StationID:   "station-1",
		Timestamp:   "2024-01-15T10:30:00Z",
		ReadingType: model.ReadingTypeTemperature,
		Value:       22.5,
		Unit:        "celsius",
		Location:    model.Location{Latitude: 40.7128, Longitude: -74.0060},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/readings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(store.readings) != 1 {
		t.Errorf("expected 1 reading in store, got %d", len(store.readings))
	}

	if len(pub.dataEvents) != 1 {
		t.Errorf("expected 1 data event published, got %d", len(pub.dataEvents))
	}
}

func TestReadingsHandler_Create_ValidationError(t *testing.T) {
	store := &mockReadingsStore{}
	pub := &mockPublisher{}
	handler := NewReadingsHandler(store, pub)

	body, _ := json.Marshal(model.CreateReadingRequest{})

	req := httptest.NewRequest(http.MethodPost, "/api/readings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReadingsHandler_Create_InvalidBody(t *testing.T) {
	handler := NewReadingsHandler(&mockReadingsStore{}, &mockPublisher{})

	req := httptest.NewRequest(http.MethodPost, "/api/readings", bytes.NewReader([]byte("invalid")))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReadingsHandler_List(t *testing.T) {
	store := &mockReadingsStore{
		readings: []model.SensorReading{
			{ID: "r-1", StationID: "station-1", ReadingType: model.ReadingTypeTemperature},
		},
	}
	handler := NewReadingsHandler(store, &mockPublisher{})

	req := httptest.NewRequest(http.MethodGet, "/api/readings", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestReadingsHandler_GetByStation(t *testing.T) {
	store := &mockReadingsStore{
		readings: []model.SensorReading{
			{ID: "r-1", StationID: "station-1"},
			{ID: "r-2", StationID: "station-2"},
		},
	}
	handler := NewReadingsHandler(store, &mockPublisher{})

	req := httptest.NewRequest(http.MethodGet, "/api/readings/station/station-1", nil)
	// Set chi URL param.
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("stationId", "station-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	handler.GetByStation(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	var readings []model.SensorReading
	if err := json.NewDecoder(rec.Body).Decode(&readings); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(readings) != 1 {
		t.Errorf("expected 1 reading, got %d", len(readings))
	}
}

func TestReadingsHandler_BatchCreate(t *testing.T) {
	store := &mockReadingsStore{}
	pub := &mockPublisher{}
	handler := NewReadingsHandler(store, pub)

	body, _ := json.Marshal(model.BatchCreateReadingsRequest{
		Readings: []model.CreateReadingRequest{
			{
				StationID:   "station-1",
				Timestamp:   "2024-01-15T10:30:00Z",
				ReadingType: model.ReadingTypeTemperature,
				Value:       22.5,
				Unit:        "celsius",
				Location:    model.Location{Latitude: 40.7128, Longitude: -74.0060},
			},
			{
				StationID:   "station-1",
				Timestamp:   "2024-01-15T10:31:00Z",
				ReadingType: model.ReadingTypeHumidity,
				Value:       65.0,
				Unit:        "percent",
				Location:    model.Location{Latitude: 40.7128, Longitude: -74.0060},
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/readings/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.BatchCreate(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp model.BatchCreateReadingsResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if resp.Succeeded != 2 {
		t.Errorf("expected 2 succeeded, got %d", resp.Succeeded)
	}
}

func TestReadingsHandler_BatchCreate_Empty(t *testing.T) {
	handler := NewReadingsHandler(&mockReadingsStore{}, &mockPublisher{})

	body, _ := json.Marshal(model.BatchCreateReadingsRequest{Readings: []model.CreateReadingRequest{}})

	req := httptest.NewRequest(http.MethodPost, "/api/readings/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.BatchCreate(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestReadingsHandler_BatchCreate_PartialFailure(t *testing.T) {
	store := &mockReadingsStore{}
	handler := NewReadingsHandler(store, &mockPublisher{})

	body, _ := json.Marshal(model.BatchCreateReadingsRequest{
		Readings: []model.CreateReadingRequest{
			{
				StationID:   "station-1",
				Timestamp:   "2024-01-15T10:30:00Z",
				ReadingType: model.ReadingTypeTemperature,
				Value:       22.5,
				Unit:        "celsius",
				Location:    model.Location{Latitude: 40.7128, Longitude: -74.0060},
			},
			{
				// Missing required fields — will fail validation.
			},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/readings/batch", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.BatchCreate(rec, req)

	if rec.Code != http.StatusMultiStatus {
		t.Errorf("expected status 207, got %d: %s", rec.Code, rec.Body.String())
	}
}
