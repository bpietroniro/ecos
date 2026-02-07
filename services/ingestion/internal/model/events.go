package model

import "time"

// EventType enumerates SNS event types.
type EventType string

const (
	EventNewDataReceived        EventType = "NewDataReceived"
	EventDataValidated          EventType = "DataValidated"
	EventDataValidationFailed   EventType = "DataValidationFailed"
	EventExtremeCondition        EventType = "ExtremeConditionDetected"
	EventStationOffline         EventType = "StationOffline"
	EventDataQualityIssue       EventType = "DataQualityIssue"
)

// DataEvent is the SNS message payload for data ingestion events.
type DataEvent struct {
	EventType EventType     `json:"eventType"`
	Timestamp string        `json:"timestamp"`
	Reading   SensorReading `json:"reading"`
	Message   string        `json:"message,omitempty"`
}

// AlertEvent is the SNS message payload for alert events.
type AlertEvent struct {
	EventType EventType `json:"eventType"`
	Timestamp string    `json:"timestamp"`
	StationID string    `json:"stationId"`
	Message   string    `json:"message"`
	Details   any       `json:"details,omitempty"`
}

// NewDataEvent creates a DataEvent with the current timestamp.
func NewDataEvent(eventType EventType, reading SensorReading, message string) DataEvent {
	return DataEvent{
		EventType: eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Reading:   reading,
		Message:   message,
	}
}

// NewAlertEvent creates an AlertEvent with the current timestamp.
func NewAlertEvent(eventType EventType, stationID, message string, details any) AlertEvent {
	return AlertEvent{
		EventType: eventType,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		StationID: stationID,
		Message:   message,
		Details:   details,
	}
}
