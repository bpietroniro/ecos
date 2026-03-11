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

	"github.com/ecos/ingestion/internal/model"
	"github.com/google/uuid"
)

// NOAAProvider fetches observations from the NOAA Weather API (api.weather.gov).
type NOAAProvider struct {
	client     *http.Client
	baseURL    string
	stationIDs []string
	userAgent  string
}

// NewNOAAProvider creates a NOAA provider.
func NewNOAAProvider(baseURL, userAgent string, stationIDs []string) *NOAAProvider {
	return &NOAAProvider{
		client: &http.Client{Timeout: 30 * time.Second},
		baseURL:    strings.TrimRight(baseURL, "/"),
		stationIDs: stationIDs,
		userAgent:  userAgent,
	}
}

func (p *NOAAProvider) Name() string { return "noaa" }

func (p *NOAAProvider) Poll(ctx context.Context) (*ProviderResult, error) {
	result := &ProviderResult{}

	for _, sid := range p.stationIDs {
		obs, err := p.fetchObservation(ctx, sid)
		if err != nil {
			slog.Warn("noaa: failed to fetch station", "stationId", sid, "error", err)
			continue
		}

		station := p.toStation(sid, obs)
		result.Stations = append(result.Stations, station)

		readings := p.toReadings(station.ID, obs)
		result.Readings = append(result.Readings, readings...)
	}

	return result, nil
}

// noaaObservation is the relevant subset of a NOAA observations/latest response.
type noaaObservation struct {
	Geometry struct {
		Coordinates []float64 `json:"coordinates"` // [lon, lat, elevation]
	} `json:"geometry"`
	Properties struct {
		Station              string         `json:"station"`
		Timestamp            string         `json:"timestamp"`
		Temperature          noaaMeasurement `json:"temperature"`
		WindSpeed            noaaMeasurement `json:"windSpeed"`
		WindDirection        noaaMeasurement `json:"windDirection"`
		BarometricPressure   noaaMeasurement `json:"barometricPressure"`
		RelativeHumidity     noaaMeasurement `json:"relativeHumidity"`
		PrecipitationLastHour noaaMeasurement `json:"precipitationLastHour"`
	} `json:"properties"`
}

type noaaMeasurement struct {
	UnitCode string   `json:"unitCode"`
	Value    *float64 `json:"value"`
}

func (p *NOAAProvider) fetchObservation(ctx context.Context, stationID string) (*noaaObservation, error) {
	url := fmt.Sprintf("%s/stations/%s/observations/latest", p.baseURL, stationID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", p.userAgent)
	req.Header.Set("Accept", "application/geo+json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var obs noaaObservation
	if err := json.NewDecoder(resp.Body).Decode(&obs); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &obs, nil
}

func (p *NOAAProvider) toStation(stationID string, obs *noaaObservation) model.Station {
	station := model.Station{
		ID:     fmt.Sprintf("noaa-%s", stationID),
		Name:   fmt.Sprintf("NOAA %s", stationID),
		Type:   model.StationTypeWeather,
		Status: model.StationStatusActive,
		Instruments: []string{
			"temperature", "wind_speed", "wind_direction",
			"pressure", "humidity", "precipitation",
		},
		Metadata: map[string]string{
			"source":    "noaa",
			"stationId": stationID,
		},
	}

	if coords := obs.Geometry.Coordinates; len(coords) >= 2 {
		station.Location = model.Location{
			Longitude: coords[0],
			Latitude:  coords[1],
		}
		if len(coords) >= 3 {
			station.Location.Elevation = coords[2]
		}
	}

	return station
}

func (p *NOAAProvider) toReadings(stationID string, obs *noaaObservation) []model.SensorReading {
	ts := obs.Properties.Timestamp
	if ts == "" {
		ts = time.Now().UTC().Format(time.RFC3339)
	}

	var loc model.Location
	if coords := obs.Geometry.Coordinates; len(coords) >= 2 {
		loc = model.Location{Longitude: coords[0], Latitude: coords[1]}
		if len(coords) >= 3 {
			loc.Elevation = coords[2]
		}
	}

	type field struct {
		m           noaaMeasurement
		readingType model.ReadingType
	}

	fields := []field{
		{obs.Properties.Temperature, model.ReadingTypeTemperature},
		{obs.Properties.WindSpeed, model.ReadingTypeWindSpeed},
		{obs.Properties.WindDirection, model.ReadingTypeWindDir},
		{obs.Properties.BarometricPressure, model.ReadingTypePressure},
		{obs.Properties.RelativeHumidity, model.ReadingTypeHumidity},
		{obs.Properties.PrecipitationLastHour, model.ReadingTypePrecip},
	}

	var readings []model.SensorReading
	for _, f := range fields {
		if f.m.Value == nil {
			continue
		}
		readings = append(readings, model.SensorReading{
			ID:          uuid.New().String(),
			StationID:   stationID,
			Timestamp:   ts,
			ReadingType: f.readingType,
			Value:       *f.m.Value,
			Unit:        noaaUnit(f.m.UnitCode),
			Quality:     model.Quality{Status: model.QualityRaw},
			Location:    loc,
			Metadata: map[string]string{
				"source":       "noaa",
				"unitCodeRaw":  f.m.UnitCode,
			},
		})
	}

	return readings
}

// noaaUnit maps NOAA wmoUnit codes to human-readable units.
func noaaUnit(unitCode string) string {
	switch unitCode {
	case "wmoUnit:degC":
		return "celsius"
	case "wmoUnit:m_s-1":
		return "m/s"
	case "wmoUnit:km_h-1":
		return "km/h"
	case "wmoUnit:degree_(angle)":
		return "degrees"
	case "wmoUnit:Pa":
		return "pascals"
	case "wmoUnit:percent":
		return "percent"
	case "wmoUnit:mm":
		return "mm"
	case "wmoUnit:m":
		return "meters"
	default:
		return unitCode
	}
}
