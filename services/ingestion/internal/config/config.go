package config

import (
	"os"
)

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
}

func Load() Config {
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
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
