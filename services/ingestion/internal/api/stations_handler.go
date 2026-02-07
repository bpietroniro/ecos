package api

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ecos/ingestion/internal/model"
	"github.com/ecos/ingestion/internal/store"
)

// StationsHandler handles station-related HTTP requests.
type StationsHandler struct {
	store store.StationsStore
}

// NewStationsHandler creates a new StationsHandler.
func NewStationsHandler(s store.StationsStore) *StationsHandler {
	return &StationsHandler{store: s}
}

// List handles GET /api/stations
func (h *StationsHandler) List(w http.ResponseWriter, r *http.Request) {
	stations, err := h.store.List(r.Context())
	if err != nil {
		slog.Error("list stations", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list stations", nil)
		return
	}

	respondJSON(w, http.StatusOK, stations)
}

// Create handles POST /api/stations
func (h *StationsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateStationRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if errs := model.ValidateCreateStation(req); len(errs) > 0 {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	station := model.Station{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Type:        req.Type,
		Location:    req.Location,
		Status:      model.StationStatusActive,
		Instruments: req.Instruments,
		Metadata:    req.Metadata,
	}

	if err := h.store.Put(r.Context(), station); err != nil {
		slog.Error("put station", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to store station", nil)
		return
	}

	respondJSON(w, http.StatusCreated, station)
}

// Update handles PUT /api/stations/{id}
func (h *StationsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required", nil)
		return
	}

	var req model.UpdateStationRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if errs := model.ValidateUpdateStation(req); len(errs) > 0 {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	existing, err := h.store.Get(r.Context(), id)
	if err != nil {
		slog.Error("get station for update", "error", err, "id", id)
		respondError(w, http.StatusInternalServerError, "failed to get station", nil)
		return
	}
	if existing == nil {
		respondError(w, http.StatusNotFound, "station not found", nil)
		return
	}

	// Apply partial updates.
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Type != nil {
		existing.Type = *req.Type
	}
	if req.Location != nil {
		existing.Location = *req.Location
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	if req.Instruments != nil {
		existing.Instruments = req.Instruments
	}
	if req.Metadata != nil {
		existing.Metadata = req.Metadata
	}

	if err := h.store.Put(r.Context(), *existing); err != nil {
		slog.Error("update station", "error", err, "id", id)
		respondError(w, http.StatusInternalServerError, "failed to update station", nil)
		return
	}

	respondJSON(w, http.StatusOK, existing)
}

// Delete handles DELETE /api/stations/{id}
func (h *StationsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		respondError(w, http.StatusBadRequest, "id is required", nil)
		return
	}

	existing, err := h.store.Get(r.Context(), id)
	if err != nil {
		slog.Error("get station for delete", "error", err, "id", id)
		respondError(w, http.StatusInternalServerError, "failed to get station", nil)
		return
	}
	if existing == nil {
		respondError(w, http.StatusNotFound, "station not found", nil)
		return
	}

	if err := h.store.Delete(r.Context(), id); err != nil {
		slog.Error("delete station", "error", err, "id", id)
		respondError(w, http.StatusInternalServerError, "failed to delete station", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
