package model

import "time"

// ReadingType enumerates supported sensor reading types.
type ReadingType string

const (
	ReadingTypeTemperature ReadingType = "temperature"
	ReadingTypeHumidity    ReadingType = "humidity"
	ReadingTypePressure    ReadingType = "pressure"
	ReadingTypeWindSpeed   ReadingType = "wind_speed"
	ReadingTypeWindDir     ReadingType = "wind_direction"
	ReadingTypePrecip      ReadingType = "precipitation"
	ReadingTypeSolarRad    ReadingType = "solar_radiation"
	ReadingTypeSeaTemp     ReadingType = "sea_temperature"
	ReadingTypeSalinity    ReadingType = "salinity"
	ReadingTypeWaveHeight  ReadingType = "wave_height"
)

// QualityStatus represents the data quality state.
type QualityStatus string

const (
	QualityRaw       QualityStatus = "raw"
	QualityValidated QualityStatus = "validated"
	QualityFlagged   QualityStatus = "flagged"
	QualityRejected  QualityStatus = "rejected"
)

// Location represents a geographic position.
type Location struct {
	Latitude  float64 `json:"latitude" dynamodbav:"latitude"`
	Longitude float64 `json:"longitude" dynamodbav:"longitude"`
	Elevation float64 `json:"elevation" dynamodbav:"elevation"`
}

// Quality represents data quality information.
type Quality struct {
	Status QualityStatus `json:"status" dynamodbav:"status"`
	Flags  []string      `json:"flags" dynamodbav:"flags"`
}

// SensorReading represents a single reading from a sensor.
type SensorReading struct {
	ID          string            `json:"id" dynamodbav:"id"`
	StationID   string            `json:"stationId" dynamodbav:"stationId"`
	Timestamp   string            `json:"timestamp" dynamodbav:"timestamp"`
	SortKey     string            `json:"-" dynamodbav:"sk"`
	ReadingType ReadingType       `json:"readingType" dynamodbav:"readingType"`
	Value       float64           `json:"value" dynamodbav:"value"`
	Unit        string            `json:"unit" dynamodbav:"unit"`
	Quality     Quality           `json:"quality" dynamodbav:"quality"`
	Location    Location          `json:"location" dynamodbav:"location"`
	Metadata    map[string]string `json:"metadata,omitempty" dynamodbav:"metadata,omitempty"`
}

// CreateReadingRequest is the request body for creating a reading.
type CreateReadingRequest struct {
	StationID   string            `json:"stationId"`
	Timestamp   string            `json:"timestamp"`
	ReadingType ReadingType       `json:"readingType"`
	Value       float64           `json:"value"`
	Unit        string            `json:"unit"`
	Location    Location          `json:"location"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// BatchCreateReadingsRequest is the request body for batch creating readings.
type BatchCreateReadingsRequest struct {
	Readings []CreateReadingRequest `json:"readings"`
}

// BatchCreateReadingsResponse is the response for batch creating readings.
type BatchCreateReadingsResponse struct {
	Succeeded int             `json:"succeeded"`
	Failed    int             `json:"failed"`
	Readings  []SensorReading `json:"readings"`
	Errors    []BatchError    `json:"errors,omitempty"`
}

// BatchError represents an error for a specific item in a batch operation.
type BatchError struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

// ReadingsQuery holds query parameters for listing readings.
type ReadingsQuery struct {
	StationID   string
	ReadingType ReadingType
	StartTime   *time.Time
	EndTime     *time.Time
	Limit       int32
}
