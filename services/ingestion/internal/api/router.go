package api

import (
	"github.com/go-chi/chi/v5"

	"github.com/ecos/ingestion/internal/events"
	"github.com/ecos/ingestion/internal/store"
)

// NewRouter creates and configures the HTTP router.
func NewRouter(
	readingsStore store.ReadingsStore,
	stationsStore store.StationsStore,
	publisher events.Publisher,
) chi.Router {
	r := chi.NewRouter()

	// Middleware stack.
	r.Use(Recovery)
	r.Use(RequestID)
	r.Use(RequestLogger)

	// Handlers.
	readings := NewReadingsHandler(readingsStore, publisher)
	stations := NewStationsHandler(stationsStore)

	// Health check.
	r.Get("/healthz", HealthHandler())

	// API routes.
	r.Route("/api", func(r chi.Router) {
		// Readings.
		r.Get("/readings", readings.List)
		r.Post("/readings", readings.Create)
		r.Post("/readings/batch", readings.BatchCreate)
		r.Get("/readings/station/{stationId}", readings.GetByStation)

		// Stations.
		r.Get("/stations", stations.List)
		r.Post("/stations", stations.Create)
		r.Put("/stations/{id}", stations.Update)
		r.Delete("/stations/{id}", stations.Delete)
	})

	return r
}
