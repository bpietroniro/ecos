package poller

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/ecos/ingestion/internal/events"
	"github.com/ecos/ingestion/internal/model"
	"github.com/ecos/ingestion/internal/store"
)

// Poller orchestrates polling providers on a configurable interval.
type Poller struct {
	providers     []Provider
	readingsStore store.ReadingsStore
	stationsStore store.StationsStore
	publisher     events.Publisher
	interval      time.Duration

	mu             sync.Mutex
	knownStations  map[string]bool
}

// New creates a Poller.
func New(
	providers []Provider,
	readingsStore store.ReadingsStore,
	stationsStore store.StationsStore,
	publisher events.Publisher,
	interval time.Duration,
) *Poller {
	return &Poller{
		providers:     providers,
		readingsStore: readingsStore,
		stationsStore: stationsStore,
		publisher:     publisher,
		interval:      interval,
		knownStations: make(map[string]bool),
	}
}

// Run polls immediately, then on the configured interval until ctx is cancelled.
func (p *Poller) Run(ctx context.Context) {
	slog.Info("poller started", "interval", p.interval, "providers", len(p.providers))

	p.tick(ctx)

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("poller stopped")
			return
		case <-ticker.C:
			p.tick(ctx)
		}
	}
}

func (p *Poller) tick(ctx context.Context) {
	for _, provider := range p.providers {
		if ctx.Err() != nil {
			return
		}

		slog.Info("polling provider", "provider", provider.Name())

		result, err := provider.Poll(ctx)
		if err != nil {
			slog.Error("provider poll failed", "provider", provider.Name(), "error", err)
			continue
		}

		p.processResult(ctx, provider.Name(), result)
	}
}

func (p *Poller) processResult(ctx context.Context, providerName string, result *ProviderResult) {
	// Upsert stations (skip if already known).
	for _, station := range result.Stations {
		if p.isKnownStation(station.ID) {
			continue
		}

		if err := p.stationsStore.Put(ctx, station); err != nil {
			slog.Error("failed to put station",
				"provider", providerName, "stationId", station.ID, "error", err)
			continue
		}

		p.markKnownStation(station.ID)
		slog.Info("registered station", "provider", providerName, "stationId", station.ID)
	}

	// Store readings and publish events.
	for _, reading := range result.Readings {
		if err := p.readingsStore.Put(ctx, reading); err != nil {
			slog.Error("failed to put reading",
				"provider", providerName, "readingId", reading.ID, "error", err)
			continue
		}

		event := model.NewDataEvent(model.EventNewDataReceived, reading, "polled from "+providerName)
		if err := p.publisher.PublishDataEvent(ctx, event); err != nil {
			slog.Error("failed to publish event",
				"provider", providerName, "readingId", reading.ID, "error", err)
		}
	}

	slog.Info("poll complete",
		"provider", providerName,
		"stations", len(result.Stations),
		"readings", len(result.Readings))
}

func (p *Poller) isKnownStation(id string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.knownStations[id]
}

func (p *Poller) markKnownStation(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.knownStations[id] = true
}
