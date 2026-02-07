package model

// StationType enumerates station types.
type StationType string

const (
	StationTypeWeather   StationType = "weather"
	StationTypeOcean     StationType = "ocean"
	StationTypeSatellite StationType = "satellite"
)

// StationStatus enumerates station statuses.
type StationStatus string

const (
	StationStatusActive      StationStatus = "active"
	StationStatusInactive    StationStatus = "inactive"
	StationStatusMaintenance StationStatus = "maintenance"
)

// Station represents a sensor station.
type Station struct {
	ID             string            `json:"id" dynamodbav:"id"`
	Name           string            `json:"name" dynamodbav:"name"`
	Type           StationType       `json:"type" dynamodbav:"type"`
	Location       Location          `json:"location" dynamodbav:"location"`
	Status         StationStatus     `json:"status" dynamodbav:"status"`
	Instruments    []string          `json:"instruments" dynamodbav:"instruments"`
	LastReportedAt string            `json:"lastReportedAt,omitempty" dynamodbav:"lastReportedAt,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
}

// CreateStationRequest is the request body for registering a station.
type CreateStationRequest struct {
	Name        string            `json:"name"`
	Type        StationType       `json:"type"`
	Location    Location          `json:"location"`
	Instruments []string          `json:"instruments"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// UpdateStationRequest is the request body for updating a station.
type UpdateStationRequest struct {
	Name        *string           `json:"name,omitempty"`
	Type        *StationType      `json:"type,omitempty"`
	Location    *Location         `json:"location,omitempty"`
	Status      *StationStatus    `json:"status,omitempty"`
	Instruments []string          `json:"instruments,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}
