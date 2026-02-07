package model

import (
	"testing"
)

func TestValidateCreateReading_Valid(t *testing.T) {
	req := CreateReadingRequest{
		StationID:   "station-1",
		Timestamp:   "2024-01-15T10:30:00Z",
		ReadingType: ReadingTypeTemperature,
		Value:       22.5,
		Unit:        "celsius",
		Location:    Location{Latitude: 40.7128, Longitude: -74.0060, Elevation: 10},
	}

	errs := ValidateCreateReading(req)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateCreateReading_MissingFields(t *testing.T) {
	req := CreateReadingRequest{}

	errs := ValidateCreateReading(req)
	if len(errs) < 3 {
		t.Errorf("expected at least 3 errors, got %d: %v", len(errs), errs)
	}

	fields := map[string]bool{}
	for _, e := range errs {
		fields[e.Field] = true
	}

	for _, f := range []string{"stationId", "timestamp", "readingType", "unit"} {
		if !fields[f] {
			t.Errorf("expected error for field %s", f)
		}
	}
}

func TestValidateCreateReading_InvalidTimestamp(t *testing.T) {
	req := CreateReadingRequest{
		StationID:   "station-1",
		Timestamp:   "not-a-timestamp",
		ReadingType: ReadingTypeTemperature,
		Value:       22.5,
		Unit:        "celsius",
		Location:    Location{Latitude: 40.7128, Longitude: -74.0060},
	}

	errs := ValidateCreateReading(req)
	if len(errs) != 1 || errs[0].Field != "timestamp" {
		t.Errorf("expected timestamp error, got %v", errs)
	}
}

func TestValidateCreateReading_InvalidReadingType(t *testing.T) {
	req := CreateReadingRequest{
		StationID:   "station-1",
		Timestamp:   "2024-01-15T10:30:00Z",
		ReadingType: "invalid_type",
		Value:       22.5,
		Unit:        "celsius",
		Location:    Location{Latitude: 40.7128, Longitude: -74.0060},
	}

	errs := ValidateCreateReading(req)
	if len(errs) != 1 || errs[0].Field != "readingType" {
		t.Errorf("expected readingType error, got %v", errs)
	}
}

func TestValidateCreateReading_InvalidLatitude(t *testing.T) {
	req := CreateReadingRequest{
		StationID:   "station-1",
		Timestamp:   "2024-01-15T10:30:00Z",
		ReadingType: ReadingTypeTemperature,
		Value:       22.5,
		Unit:        "celsius",
		Location:    Location{Latitude: 91, Longitude: -74.0060},
	}

	errs := ValidateCreateReading(req)
	if len(errs) != 1 || errs[0].Field != "location.latitude" {
		t.Errorf("expected location.latitude error, got %v", errs)
	}
}

func TestValidateCreateStation_Valid(t *testing.T) {
	req := CreateStationRequest{
		Name:        "Test Station",
		Type:        StationTypeWeather,
		Location:    Location{Latitude: 40.7128, Longitude: -74.0060, Elevation: 10},
		Instruments: []string{"thermometer", "barometer"},
	}

	errs := ValidateCreateStation(req)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateCreateStation_MissingFields(t *testing.T) {
	req := CreateStationRequest{}

	errs := ValidateCreateStation(req)
	if len(errs) < 2 {
		t.Errorf("expected at least 2 errors, got %d: %v", len(errs), errs)
	}
}

func TestValidateCreateStation_InvalidType(t *testing.T) {
	req := CreateStationRequest{
		Name:     "Test Station",
		Type:     "invalid",
		Location: Location{Latitude: 40.7128, Longitude: -74.0060},
	}

	errs := ValidateCreateStation(req)
	if len(errs) != 1 || errs[0].Field != "type" {
		t.Errorf("expected type error, got %v", errs)
	}
}

func TestValidateUpdateStation_Valid(t *testing.T) {
	name := "Updated Station"
	status := StationStatusMaintenance
	req := UpdateStationRequest{
		Name:   &name,
		Status: &status,
	}

	errs := ValidateUpdateStation(req)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateUpdateStation_InvalidStatus(t *testing.T) {
	status := StationStatus("bogus")
	req := UpdateStationRequest{
		Status: &status,
	}

	errs := ValidateUpdateStation(req)
	if len(errs) != 1 || errs[0].Field != "status" {
		t.Errorf("expected status error, got %v", errs)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	errs := ValidationErrors{
		{Field: "a", Message: "bad"},
		{Field: "b", Message: "worse"},
	}
	s := errs.Error()
	if s == "" {
		t.Error("expected non-empty string")
	}
}

func TestValidationErrors_Empty(t *testing.T) {
	errs := ValidationErrors{}
	if errs.Error() != "" {
		t.Error("expected empty string for no errors")
	}
}
