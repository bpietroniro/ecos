package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ecos/ingestion/internal/config"
	"github.com/ecos/ingestion/internal/model"
	"github.com/google/uuid"
)

// OWMProvider fetches weather data from the OpenWeatherMap API.
type OWMProvider struct {
	client    *http.Client
	baseURL   string
	apiKey    string
	locations []config.OWMLocation
}

// NewOWMProvider creates an OpenWeatherMap provider.
func NewOWMProvider(baseURL, apiKey string, locations []config.OWMLocation) *OWMProvider {
	return &OWMProvider{
		client:    &http.Client{Timeout: 30 * time.Second},
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		locations: locations,
	}
}

func (p *OWMProvider) Name() string { return "openweathermap" }

func (p *OWMProvider) Poll(ctx context.Context) (*ProviderResult, error) {
	result := &ProviderResult{}

	for _, loc := range p.locations {
		weather, err := p.fetchWeather(ctx, loc)
		if err != nil {
			slog.Warn("owm: failed to fetch location", "name", loc.Name, "error", err)
			continue
		}

		station := p.toStation(loc, weather)
		result.Stations = append(result.Stations, station)

		readings := p.toReadings(station.ID, loc, weather)
		result.Readings = append(result.Readings, readings...)
	}

	return result, nil
}

// owmResponse is the relevant subset of the OWM /data/2.5/weather response.
type owmResponse struct {
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Main struct {
		Temp     float64 `json:"temp"`
		Pressure float64 `json:"pressure"`
		Humidity float64 `json:"humidity"`
	} `json:"main"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Rain *struct {
		OneHour float64 `json:"1h"`
	} `json:"rain"`
	Dt int64 `json:"dt"`
}

func (p *OWMProvider) fetchWeather(ctx context.Context, loc config.OWMLocation) (*owmResponse, error) {
	url := fmt.Sprintf("%s/data/2.5/weather?lat=%s&lon=%s&appid=%s&units=metric",
		p.baseURL, loc.Lat, loc.Lon, p.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var weather owmResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &weather, nil
}

func (p *OWMProvider) toStation(loc config.OWMLocation, w *owmResponse) model.Station {
	return model.Station{
		ID:     fmt.Sprintf("owm-%s", strings.ToLower(loc.Name)),
		Name:   fmt.Sprintf("OpenWeatherMap %s", loc.Name),
		Type:   model.StationTypeWeather,
		Status: model.StationStatusActive,
		Location: model.Location{
			Latitude:  w.Coord.Lat,
			Longitude: w.Coord.Lon,
		},
		Instruments: []string{
			"temperature", "pressure", "humidity",
			"wind_speed", "wind_direction", "precipitation",
		},
		Metadata: map[string]string{
			"source": "openweathermap",
			"name":   loc.Name,
		},
	}
}

func (p *OWMProvider) toReadings(stationID string, _ config.OWMLocation, w *owmResponse) []model.SensorReading {
	ts := time.Unix(w.Dt, 0).UTC().Format(time.RFC3339)
	location := model.Location{
		Latitude:  w.Coord.Lat,
		Longitude: w.Coord.Lon,
	}
	meta := map[string]string{"source": "openweathermap"}

	readings := []model.SensorReading{
		{
			ID:          uuid.New().String(),
			StationID:   stationID,
			Timestamp:   ts,
			ReadingType: model.ReadingTypeTemperature,
			Value:       w.Main.Temp,
			Unit:        "celsius",
			Quality:     model.Quality{Status: model.QualityRaw},
			Location:    location,
			Metadata:    meta,
		},
		{
			ID:          uuid.New().String(),
			StationID:   stationID,
			Timestamp:   ts,
			ReadingType: model.ReadingTypePressure,
			Value:       w.Main.Pressure,
			Unit:        "hPa",
			Quality:     model.Quality{Status: model.QualityRaw},
			Location:    location,
			Metadata:    meta,
		},
		{
			ID:          uuid.New().String(),
			StationID:   stationID,
			Timestamp:   ts,
			ReadingType: model.ReadingTypeHumidity,
			Value:       w.Main.Humidity,
			Unit:        "percent",
			Quality:     model.Quality{Status: model.QualityRaw},
			Location:    location,
			Metadata:    meta,
		},
		{
			ID:          uuid.New().String(),
			StationID:   stationID,
			Timestamp:   ts,
			ReadingType: model.ReadingTypeWindSpeed,
			Value:       w.Wind.Speed,
			Unit:        "m/s",
			Quality:     model.Quality{Status: model.QualityRaw},
			Location:    location,
			Metadata:    meta,
		},
		{
			ID:          uuid.New().String(),
			StationID:   stationID,
			Timestamp:   ts,
			ReadingType: model.ReadingTypeWindDir,
			Value:       w.Wind.Deg,
			Unit:        "degrees",
			Quality:     model.Quality{Status: model.QualityRaw},
			Location:    location,
			Metadata:    meta,
		},
	}

	if w.Rain != nil {
		readings = append(readings, model.SensorReading{
			ID:          uuid.New().String(),
			StationID:   stationID,
			Timestamp:   ts,
			ReadingType: model.ReadingTypePrecip,
			Value:       w.Rain.OneHour,
			Unit:        "mm",
			Quality:     model.Quality{Status: model.QualityRaw},
			Location:    location,
			Metadata:    meta,
		})
	}

	return readings
}
