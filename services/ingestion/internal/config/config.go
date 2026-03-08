package config

import (
	"os"
	"strings"
	"time"
)

// OWMLocation represents a named lat/lon pair for OpenWeatherMap polling.
type OWMLocation struct {
	Name string
	Lat  string
	Lon  string
}

type Config struct {
	Port                  string
	AWSRegion             string
	DynamoDBEndpoint      string
	ReadingsTableName     string
	StationsTableName     string
	ReadingTypeIndexName  string
	DataIngestionTopicARN string
	AlertEventsTopicARN   string
	SNSEndpoint           string

	// Poller
	PollEnabled  bool
	PollInterval time.Duration

	// NOAA
	NOAAEnabled    bool
	NOAAStationIDs []string
	NOAAUserAgent  string
	NOAABaseURL    string

	// OpenWeatherMap
	OWMEnabled   bool
	OWMAPIKey    string
	OWMLocations []OWMLocation
	OWMBaseURL   string
}

func Load() Config {
	pollInterval, _ := time.ParseDuration(envOrDefault("POLL_INTERVAL", "5m"))

	return Config{
		Port:                  envOrDefault("PORT", "8080"),
		AWSRegion:             envOrDefault("AWS_REGION", "us-east-1"),
		DynamoDBEndpoint:      os.Getenv("DYNAMODB_ENDPOINT"),
		ReadingsTableName:     envOrDefault("READINGS_TABLE_NAME", "ecos-sensor-readings"),
		StationsTableName:     envOrDefault("STATIONS_TABLE_NAME", "ecos-stations"),
		ReadingTypeIndexName:  envOrDefault("READING_TYPE_INDEX_NAME", "byReadingType"),
		DataIngestionTopicARN: os.Getenv("DATA_INGESTION_TOPIC_ARN"),
		AlertEventsTopicARN:   os.Getenv("ALERT_EVENTS_TOPIC_ARN"),
		SNSEndpoint:           os.Getenv("SNS_ENDPOINT"),

		PollEnabled:  os.Getenv("POLL_ENABLED") == "true",
		PollInterval: pollInterval,

		NOAAEnabled:    os.Getenv("NOAA_ENABLED") == "true",
		NOAAStationIDs: parseCSV(os.Getenv("NOAA_STATION_IDS")),
		NOAAUserAgent:  envOrDefault("NOAA_USER_AGENT", "(ecos-platform, contact@ecos.dev)"),
		NOAABaseURL:    envOrDefault("NOAA_BASE_URL", "https://api.weather.gov"),

		OWMEnabled:   os.Getenv("OWM_ENABLED") == "true",
		OWMAPIKey:    os.Getenv("OWM_API_KEY"),
		OWMLocations: parseOWMLocations(os.Getenv("OWM_LOCATIONS")),
		OWMBaseURL:   envOrDefault("OWM_BASE_URL", "https://api.openweathermap.org"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parseCSV splits a comma-separated string into trimmed, non-empty tokens.
func parseCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// parseOWMLocations parses "Name:lat,lon;Name2:lat2,lon2" format.
func parseOWMLocations(s string) []OWMLocation {
	if s == "" {
		return nil
	}
	var locs []OWMLocation
	for _, entry := range strings.Split(s, ";") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Format: Name:lat,lon
		parts := strings.SplitN(entry, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name := strings.TrimSpace(parts[0])
		coords := strings.SplitN(parts[1], ",", 2)
		if len(coords) != 2 {
			continue
		}
		locs = append(locs, OWMLocation{
			Name: name,
			Lat:  strings.TrimSpace(coords[0]),
			Lon:  strings.TrimSpace(coords[1]),
		})
	}
	return locs
}
