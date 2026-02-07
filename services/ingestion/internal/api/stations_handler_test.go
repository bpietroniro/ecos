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

// mockStationsStore implements store.StationsStore for testing.
type mockStationsStore struct {
	stations map[string]model.Station
}

func newMockStationsStore() *mockStationsStore {
	return &mockStationsStore{stations: make(map[string]model.Station)}
}

func (m *mockStationsStore) Get(_ context.Context, id string) (*model.Station, error) {
	s, ok := m.stations[id]
	if !ok {
		return nil, nil
	}
	return &s, nil
}

func (m *mockStationsStore) Put(_ context.Context, station model.Station) error {
	m.stations[station.ID] = station
	return nil
}

func (m *mockStationsStore) Delete(_ context.Context, id string) error {
	delete(m.stations, id)
	return nil
}

func (m *mockStationsStore) List(_ context.Context) ([]model.Station, error) {
	result := make([]model.Station, 0, len(m.stations))
	for _, s := range m.stations {
		result = append(result, s)
	}
	return result, nil
}

func TestStationsHandler_Create(t *testing.T) {
	store := newMockStationsStore()
	handler := NewStationsHandler(store)

	body, _ := json.Marshal(model.CreateStationRequest{
		Name:        "Test Station",
		Type:        model.StationTypeWeather,
		Location:    model.Location{Latitude: 40.7128, Longitude: -74.0060, Elevation: 10},
		Instruments: []string{"thermometer"},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/stations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", rec.Code, rec.Body.String())
	}

	if len(store.stations) != 1 {
		t.Errorf("expected 1 station in store, got %d", len(store.stations))
	}
}

func TestStationsHandler_Create_ValidationError(t *testing.T) {
	handler := NewStationsHandler(newMockStationsStore())

	body, _ := json.Marshal(model.CreateStationRequest{})
	req := httptest.NewRequest(http.MethodPost, "/api/stations", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

func TestStationsHandler_List(t *testing.T) {
	store := newMockStationsStore()
	store.stations["s-1"] = model.Station{ID: "s-1", Name: "Station A"}
	handler := NewStationsHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/api/stations", nil)
	rec := httptest.NewRecorder()

	handler.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestStationsHandler_Update(t *testing.T) {
	store := newMockStationsStore()
	store.stations["s-1"] = model.Station{
		ID:     "s-1",
		Name:   "Old Name",
		Type:   model.StationTypeWeather,
		Status: model.StationStatusActive,
	}
	handler := NewStationsHandler(store)

	newName := "New Name"
	body, _ := json.Marshal(model.UpdateStationRequest{Name: &newName})

	req := httptest.NewRequest(http.MethodPut, "/api/stations/s-1", bytes.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "s-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	if store.stations["s-1"].Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", store.stations["s-1"].Name)
	}
}

func TestStationsHandler_Update_NotFound(t *testing.T) {
	handler := NewStationsHandler(newMockStationsStore())

	newName := "Name"
	body, _ := json.Marshal(model.UpdateStationRequest{Name: &newName})

	req := httptest.NewRequest(http.MethodPut, "/api/stations/nonexistent", bytes.NewReader(body))
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	handler.Update(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}

func TestStationsHandler_Delete(t *testing.T) {
	store := newMockStationsStore()
	store.stations["s-1"] = model.Station{ID: "s-1", Name: "Station A"}
	handler := NewStationsHandler(store)

	req := httptest.NewRequest(http.MethodDelete, "/api/stations/s-1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "s-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", rec.Code)
	}

	if len(store.stations) != 0 {
		t.Errorf("expected 0 stations, got %d", len(store.stations))
	}
}

func TestStationsHandler_Delete_NotFound(t *testing.T) {
	handler := NewStationsHandler(newMockStationsStore())

	req := httptest.NewRequest(http.MethodDelete, "/api/stations/nonexistent", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "nonexistent")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	handler.Delete(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rec.Code)
	}
}
