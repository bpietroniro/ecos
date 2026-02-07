package model

import (
	"fmt"
	"time"
)

// ValidationError holds validation error details.
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	msg := "validation failed:"
	for _, err := range e {
		msg += " " + err.Error() + ";"
	}
	return msg
}

var validReadingTypes = map[ReadingType]bool{
	ReadingTypeTemperature: true,
	ReadingTypeHumidity:    true,
	ReadingTypePressure:    true,
	ReadingTypeWindSpeed:   true,
	ReadingTypeWindDir:     true,
	ReadingTypePrecip:      true,
	ReadingTypeSolarRad:    true,
	ReadingTypeSeaTemp:     true,
	ReadingTypeSalinity:    true,
	ReadingTypeWaveHeight:  true,
}

var validStationTypes = map[StationType]bool{
	StationTypeWeather:   true,
	StationTypeOcean:     true,
	StationTypeSatellite: true,
}

var validStationStatuses = map[StationStatus]bool{
	StationStatusActive:      true,
	StationStatusInactive:    true,
	StationStatusMaintenance: true,
}

// ValidateCreateReading validates a CreateReadingRequest.
func ValidateCreateReading(r CreateReadingRequest) ValidationErrors {
	var errs ValidationErrors

	if r.StationID == "" {
		errs = append(errs, ValidationError{Field: "stationId", Message: "required"})
	}
	if r.Timestamp == "" {
		errs = append(errs, ValidationError{Field: "timestamp", Message: "required"})
	} else if _, err := time.Parse(time.RFC3339, r.Timestamp); err != nil {
		errs = append(errs, ValidationError{Field: "timestamp", Message: "must be RFC3339 format"})
	}
	if r.ReadingType == "" {
		errs = append(errs, ValidationError{Field: "readingType", Message: "required"})
	} else if !validReadingTypes[r.ReadingType] {
		errs = append(errs, ValidationError{Field: "readingType", Message: "invalid reading type"})
	}
	if r.Unit == "" {
		errs = append(errs, ValidationError{Field: "unit", Message: "required"})
	}
	if err := validateLocation(r.Location); err != nil {
		errs = append(errs, *err)
	}

	return errs
}

// ValidateCreateStation validates a CreateStationRequest.
func ValidateCreateStation(s CreateStationRequest) ValidationErrors {
	var errs ValidationErrors

	if s.Name == "" {
		errs = append(errs, ValidationError{Field: "name", Message: "required"})
	}
	if s.Type == "" {
		errs = append(errs, ValidationError{Field: "type", Message: "required"})
	} else if !validStationTypes[s.Type] {
		errs = append(errs, ValidationError{Field: "type", Message: "invalid station type"})
	}
	if err := validateLocation(s.Location); err != nil {
		errs = append(errs, *err)
	}

	return errs
}

// ValidateUpdateStation validates an UpdateStationRequest.
func ValidateUpdateStation(s UpdateStationRequest) ValidationErrors {
	var errs ValidationErrors

	if s.Type != nil && !validStationTypes[*s.Type] {
		errs = append(errs, ValidationError{Field: "type", Message: "invalid station type"})
	}
	if s.Status != nil && !validStationStatuses[*s.Status] {
		errs = append(errs, ValidationError{Field: "status", Message: "invalid station status"})
	}
	if s.Location != nil {
		if err := validateLocation(*s.Location); err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

func validateLocation(loc Location) *ValidationError {
	if loc.Latitude < -90 || loc.Latitude > 90 {
		return &ValidationError{Field: "location.latitude", Message: "must be between -90 and 90"}
	}
	if loc.Longitude < -180 || loc.Longitude > 180 {
		return &ValidationError{Field: "location.longitude", Message: "must be between -180 and 180"}
	}
	return nil
}
