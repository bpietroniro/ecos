package poller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ecos/ingestion/internal/config"
	"github.com/ecos/ingestion/internal/model"
)

const owmResponseNYC = `{
  "coord": {"lon": -74.0, "lat": 40.71},
  "main": {"temp": 8.5, "pressure": 1013.0, "humidity": 72.0},
  "wind": {"speed": 5.1, "deg": 210.0},
  "rain": {"1h": 0.5},
  "dt": 1705312800
}`

const owmResponseNoRain = `{
  "coord": {"lon": -118.24, "lat": 34.05},
  "main": {"temp": 22.0, "pressure": 1015.0, "humidity": 45.0},
  "wind": {"speed": 3.2, "deg": 90.0},
  "dt": 1705312800
}`

func TestOWMProvider_Poll(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/data/2.5/weather" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if q.Get("appid") != "test-key" {
			t.Errorf("expected appid=test-key, got %q", q.Get("appid"))
		}
		if q.Get("units") != "metric" {
			t.Errorf("expected units=metric, got %q", q.Get("units"))
		}

		if q.Get("lat") == "40.71" {
			w.Write([]byte(owmResponseNYC))
		} else {
			w.Write([]byte(owmResponseNoRain))
		}
	}))
	defer srv.Close()

	locations := []config.OWMLocation{
		{Name: "NYC", Lat: "40.71", Lon: "-74.00"},
	}

	provider := NewOWMProvider(srv.URL, "test-key", locations)
	result, err := provider.Poll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Station
	if len(result.Stations) != 1 {
		t.Fatalf("expected 1 station, got %d", len(result.Stations))
	}
	s := result.Stations[0]
	if s.ID != "owm-nyc" {
		t.Errorf("station ID = %q, want %q", s.ID, "owm-nyc")
	}
	if s.Type != model.StationTypeWeather {
		t.Errorf("station type = %q, want %q", s.Type, model.StationTypeWeather)
	}
	if s.Location.Latitude != 40.71 {
		t.Errorf("latitude = %f, want 40.71", s.Location.Latitude)
	}

	// Readings — 5 base + 1 rain = 6
	if len(result.Readings) != 6 {
		t.Fatalf("expected 6 readings (with rain), got %d", len(result.Readings))
	}

	readingMap := make(map[model.ReadingType]model.SensorReading)
	for _, r := range result.Readings {
		readingMap[r.ReadingType] = r
	}

	temp := readingMap[model.ReadingTypeTemperature]
	if temp.Value != 8.5 {
		t.Errorf("temperature = %f, want 8.5", temp.Value)
	}
	if temp.Unit != "celsius" {
		t.Errorf("temperature unit = %q, want %q", temp.Unit, "celsius")
	}
	if temp.StationID != "owm-nyc" {
		t.Errorf("stationId = %q, want %q", temp.StationID, "owm-nyc")
	}

	pressure := readingMap[model.ReadingTypePressure]
	if pressure.Value != 1013.0 {
		t.Errorf("pressure = %f, want 1013.0", pressure.Value)
	}
	if pressure.Unit != "hPa" {
		t.Errorf("pressure unit = %q, want %q", pressure.Unit, "hPa")
	}

	precip := readingMap[model.ReadingTypePrecip]
	if precip.Value != 0.5 {
		t.Errorf("precipitation = %f, want 0.5", precip.Value)
	}
}

func TestOWMProvider_PollNoRain(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(owmResponseNoRain))
	}))
	defer srv.Close()

	locations := []config.OWMLocation{
		{Name: "LA", Lat: "34.05", Lon: "-118.24"},
	}

	provider := NewOWMProvider(srv.URL, "test-key", locations)
	result, err := provider.Poll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 5 readings without rain
	if len(result.Readings) != 5 {
		t.Fatalf("expected 5 readings (no rain), got %d", len(result.Readings))
	}

	for _, r := range result.Readings {
		if r.ReadingType == model.ReadingTypePrecip {
			t.Error("precipitation should not be present when rain is absent")
		}
	}
}

func TestOWMProvider_PollSkipsFailedLocation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("lat") == "0.00" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"message":"Invalid API key"}`))
			return
		}
		w.Write([]byte(owmResponseNYC))
	}))
	defer srv.Close()

	locations := []config.OWMLocation{
		{Name: "BAD", Lat: "0.00", Lon: "0.00"},
		{Name: "NYC", Lat: "40.71", Lon: "-74.00"},
	}

	provider := NewOWMProvider(srv.URL, "test-key", locations)
	result, err := provider.Poll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Stations) != 1 {
		t.Errorf("expected 1 station (BAD skipped), got %d", len(result.Stations))
	}
}

func TestOWMProvider_Name(t *testing.T) {
	p := NewOWMProvider("http://localhost", "key", nil)
	if p.Name() != "openweathermap" {
		t.Errorf("Name() = %q, want %q", p.Name(), "openweathermap")
	}
}
