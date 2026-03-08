package poller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecos/ingestion/internal/model"
)

const noaaResponseKNYC = `{
  "geometry": {
    "type": "Point",
    "coordinates": [-73.96, 40.78, 48]
  },
  "properties": {
    "station": "https://api.weather.gov/stations/KNYC",
    "timestamp": "2025-01-15T12:00:00+00:00",
    "temperature": {"unitCode": "wmoUnit:degC", "value": 5.5},
    "windSpeed": {"unitCode": "wmoUnit:km_h-1", "value": 12.0},
    "windDirection": {"unitCode": "wmoUnit:degree_(angle)", "value": 180.0},
    "barometricPressure": {"unitCode": "wmoUnit:Pa", "value": 101300.0},
    "relativeHumidity": {"unitCode": "wmoUnit:percent", "value": 65.0},
    "precipitationLastHour": {"unitCode": "wmoUnit:mm", "value": null}
  }
}`

func TestNOAAProvider_Poll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/stations/KNYC/observations/latest" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("User-Agent") != "test-agent" {
			t.Errorf("expected User-Agent test-agent, got %q", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/geo+json")
		w.Write([]byte(noaaResponseKNYC))
	}))
	defer srv.Close()

	provider := NewNOAAProvider(srv.URL, "test-agent", []string{"KNYC"})
	result, err := provider.Poll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Station
	if len(result.Stations) != 1 {
		t.Fatalf("expected 1 station, got %d", len(result.Stations))
	}
	s := result.Stations[0]
	if s.ID != "noaa-KNYC" {
		t.Errorf("station ID = %q, want %q", s.ID, "noaa-KNYC")
	}
	if s.Type != model.StationTypeWeather {
		t.Errorf("station type = %q, want %q", s.Type, model.StationTypeWeather)
	}
	if s.Location.Latitude != 40.78 {
		t.Errorf("latitude = %f, want 40.78", s.Location.Latitude)
	}
	if s.Location.Longitude != -73.96 {
		t.Errorf("longitude = %f, want -73.96", s.Location.Longitude)
	}
	if s.Location.Elevation != 48 {
		t.Errorf("elevation = %f, want 48", s.Location.Elevation)
	}

	// Readings — precipitationLastHour is null, so expect 5 readings
	if len(result.Readings) != 5 {
		t.Fatalf("expected 5 readings, got %d", len(result.Readings))
	}

	readingMap := make(map[model.ReadingType]model.SensorReading)
	for _, r := range result.Readings {
		readingMap[r.ReadingType] = r
	}

	temp, ok := readingMap[model.ReadingTypeTemperature]
	if !ok {
		t.Fatal("missing temperature reading")
	}
	if temp.Value != 5.5 {
		t.Errorf("temperature value = %f, want 5.5", temp.Value)
	}
	if temp.Unit != "celsius" {
		t.Errorf("temperature unit = %q, want %q", temp.Unit, "celsius")
	}
	if temp.StationID != "noaa-KNYC" {
		t.Errorf("stationId = %q, want %q", temp.StationID, "noaa-KNYC")
	}
	if temp.Timestamp != "2025-01-15T12:00:00+00:00" {
		t.Errorf("timestamp = %q, want %q", temp.Timestamp, "2025-01-15T12:00:00+00:00")
	}
	if temp.Quality.Status != model.QualityRaw {
		t.Errorf("quality status = %q, want %q", temp.Quality.Status, model.QualityRaw)
	}

	wind, ok := readingMap[model.ReadingTypeWindSpeed]
	if !ok {
		t.Fatal("missing wind speed reading")
	}
	if wind.Value != 12.0 {
		t.Errorf("wind speed value = %f, want 12.0", wind.Value)
	}
	if wind.Unit != "km/h" {
		t.Errorf("wind speed unit = %q, want %q", wind.Unit, "km/h")
	}

	pressure, ok := readingMap[model.ReadingTypePressure]
	if !ok {
		t.Fatal("missing pressure reading")
	}
	if pressure.Value != 101300.0 {
		t.Errorf("pressure value = %f, want 101300.0", pressure.Value)
	}
	if pressure.Unit != "pascals" {
		t.Errorf("pressure unit = %q, want %q", pressure.Unit, "pascals")
	}

	if _, ok := readingMap[model.ReadingTypePrecip]; ok {
		t.Error("precipitation should be skipped when value is null")
	}
}

func TestNOAAProvider_PollSkipsFailedStation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/stations/BAD/observations/latest" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"detail":"station not found"}`))
			return
		}
		w.Write([]byte(noaaResponseKNYC))
	}))
	defer srv.Close()

	provider := NewNOAAProvider(srv.URL, "test-agent", []string{"BAD", "KNYC"})
	result, err := provider.Poll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Stations) != 1 {
		t.Errorf("expected 1 station (BAD skipped), got %d", len(result.Stations))
	}
}

func TestNOAAProvider_Name(t *testing.T) {
	p := NewNOAAProvider("http://localhost", "agent", nil)
	if p.Name() != "noaa" {
		t.Errorf("Name() = %q, want %q", p.Name(), "noaa")
	}
}

func TestNoaaUnit(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"wmoUnit:degC", "celsius"},
		{"wmoUnit:km_h-1", "km/h"},
		{"wmoUnit:Pa", "pascals"},
		{"wmoUnit:percent", "percent"},
		{"wmoUnit:mm", "mm"},
		{"wmoUnit:degree_(angle)", "degrees"},
		{"wmoUnit:m", "meters"},
		{"unknown:unit", "unknown:unit"},
	}
	for _, tt := range tests {
		if got := noaaUnit(tt.input); got != tt.want {
			t.Errorf("noaaUnit(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
