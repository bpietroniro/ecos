package poller

import (
	"context"

	"github.com/ecos/ingestion/internal/model"
)

// ProviderResult holds the stations and readings fetched by a provider.
type ProviderResult struct {
	Readings []model.SensorReading
	Stations []model.Station
}

// Provider fetches weather data from an external API.
type Provider interface {
	Name() string
	Poll(ctx context.Context) (*ProviderResult, error)
}
