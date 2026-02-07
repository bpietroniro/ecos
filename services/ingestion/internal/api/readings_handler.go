package api

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ecos/ingestion/internal/events"
	"github.com/ecos/ingestion/internal/model"
	"github.com/ecos/ingestion/internal/store"
)

// ReadingsHandler handles reading-related HTTP requests.
type ReadingsHandler struct {
	store     store.ReadingsStore
	publisher events.Publisher
}

// NewReadingsHandler creates a new ReadingsHandler.
func NewReadingsHandler(s store.ReadingsStore, p events.Publisher) *ReadingsHandler {
	return &ReadingsHandler{store: s, publisher: p}
}

// List handles GET /api/readings
func (h *ReadingsHandler) List(w http.ResponseWriter, r *http.Request) {
	q, err := parseReadingsQuery(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid query parameters", err.Error())
		return
	}

	readings, err := h.store.List(r.Context(), q)
	if err != nil {
		slog.Error("list readings", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list readings", nil)
		return
	}

	respondJSON(w, http.StatusOK, readings)
}

// GetByStation handles GET /api/readings/station/{stationId}
func (h *ReadingsHandler) GetByStation(w http.ResponseWriter, r *http.Request) {
	stationID := chi.URLParam(r, "stationId")
	if stationID == "" {
		respondError(w, http.StatusBadRequest, "stationId is required", nil)
		return
	}

	q, err := parseReadingsQuery(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid query parameters", err.Error())
		return
	}

	readings, err := h.store.GetByStation(r.Context(), stationID, q)
	if err != nil {
		slog.Error("get readings by station", "error", err, "stationId", stationID)
		respondError(w, http.StatusInternalServerError, "failed to get readings", nil)
		return
	}

	respondJSON(w, http.StatusOK, readings)
}

// Create handles POST /api/readings
func (h *ReadingsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateReadingRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if errs := model.ValidateCreateReading(req); len(errs) > 0 {
		respondError(w, http.StatusBadRequest, "validation failed", errs)
		return
	}

	reading := model.SensorReading{
		ID:          uuid.New().String(),
		StationID:   req.StationID,
		Timestamp:   req.Timestamp,
		ReadingType: req.ReadingType,
		Value:       req.Value,
		Unit:        req.Unit,
		Quality:     model.Quality{Status: model.QualityRaw, Flags: []string{}},
		Location:    req.Location,
		Metadata:    req.Metadata,
	}

	if err := h.store.Put(r.Context(), reading); err != nil {
		slog.Error("put reading", "error", err)
		respondError(w, http.StatusInternalServerError, "failed to store reading", nil)
		return
	}

	// Publish NewDataReceived event.
	event := model.NewDataEvent(model.EventNewDataReceived, reading, "new sensor reading ingested")
	if err := h.publisher.PublishDataEvent(r.Context(), event); err != nil {
		slog.Error("publish data event", "error", err, "readingId", reading.ID)
	}

	respondJSON(w, http.StatusCreated, reading)
}

// BatchCreate handles POST /api/readings/batch
func (h *ReadingsHandler) BatchCreate(w http.ResponseWriter, r *http.Request) {
	var req model.BatchCreateReadingsRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	if len(req.Readings) == 0 {
		respondError(w, http.StatusBadRequest, "readings array is required and must not be empty", nil)
		return
	}

	if len(req.Readings) > 100 {
		respondError(w, http.StatusBadRequest, "batch size must not exceed 100", nil)
		return
	}

	resp := model.BatchCreateReadingsResponse{}

	for i, reqReading := range req.Readings {
		if errs := model.ValidateCreateReading(reqReading); len(errs) > 0 {
			resp.Failed++
			resp.Errors = append(resp.Errors, model.BatchError{
				Index:   i,
				Message: errs.Error(),
			})
			continue
		}

		reading := model.SensorReading{
			ID:          uuid.New().String(),
			StationID:   reqReading.StationID,
			Timestamp:   reqReading.Timestamp,
			ReadingType: reqReading.ReadingType,
			Value:       reqReading.Value,
			Unit:        reqReading.Unit,
			Quality:     model.Quality{Status: model.QualityRaw, Flags: []string{}},
			Location:    reqReading.Location,
			Metadata:    reqReading.Metadata,
		}

		if err := h.store.Put(r.Context(), reading); err != nil {
			slog.Error("batch put reading", "error", err, "index", i)
			resp.Failed++
			resp.Errors = append(resp.Errors, model.BatchError{
				Index:   i,
				Message: "failed to store reading",
			})
			continue
		}

		// Publish event per reading.
		event := model.NewDataEvent(model.EventNewDataReceived, reading, "new sensor reading ingested (batch)")
		if err := h.publisher.PublishDataEvent(r.Context(), event); err != nil {
			slog.Error("publish data event", "error", err, "readingId", reading.ID)
		}

		resp.Succeeded++
		resp.Readings = append(resp.Readings, reading)
	}

	status := http.StatusCreated
	if resp.Failed > 0 && resp.Succeeded == 0 {
		status = http.StatusBadRequest
	} else if resp.Failed > 0 {
		status = http.StatusMultiStatus
	}

	respondJSON(w, status, resp)
}

func parseReadingsQuery(r *http.Request) (model.ReadingsQuery, error) {
	q := model.ReadingsQuery{
		StationID:   r.URL.Query().Get("stationId"),
		ReadingType: model.ReadingType(r.URL.Query().Get("readingType")),
	}

	if v := r.URL.Query().Get("startTime"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return q, err
		}
		q.StartTime = &t
	}

	if v := r.URL.Query().Get("endTime"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return q, err
		}
		q.EndTime = &t
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		n, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return q, err
		}
		q.Limit = int32(n)
	}

	return q, nil
}
