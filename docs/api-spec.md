# API Specification

## Data Ingestion Service

### Readings

```
GET    /api/readings                          List readings (with filters)
GET    /api/readings/{id}                     Get a specific reading
GET    /api/readings/station/{stationId}      Get readings for a station
POST   /api/readings                          Submit a single reading
POST   /api/readings/batch                    Submit readings in batch
```

### Stations

```
GET    /api/stations                          List all stations
POST   /api/stations                          Register a new station
PUT    /api/stations/{id}                     Update station metadata
DELETE /api/stations/{id}                     Deregister a station
```

## Analysis Service

### Jobs

```
POST   /api/analysis/jobs                     Create an analysis job
GET    /api/analysis/jobs/{jobId}             Get job details
GET    /api/analysis/jobs/{jobId}/status      Get job status/progress
GET    /api/analysis/jobs/{jobId}/results     Get job results
DELETE /api/analysis/jobs/{jobId}             Cancel/delete a job
```

### Cache (internal)

These endpoints are used internally by the Analysis Service for cache management. They are not exposed through the API Gateway.

```
GET    /cache/readings/{hash}                 Get cached reading aggregation
POST   /cache/readings/{hash}                 Store cached reading aggregation
DELETE /cache/readings/{hash}                 Invalidate cached reading
GET    /cache/analysis/{jobId}                Get cached analysis result
POST   /cache/analysis/{jobId}                Store cached analysis result
```

## Message Queue Events

All events are published to SQS/SNS topics. Consumers subscribe to relevant topics.

### Data Ingestion Events

| Event | Publisher | Description |
|---|---|---|
| `NewDataReceived` | Data Ingestion Service | New sensor data ingested |
| `DataValidated` | Data Ingestion Service | Data passed quality checks |
| `DataValidationFailed` | Data Ingestion Service | Data failed validation |

### Analysis Events

| Event | Publisher | Description |
|---|---|---|
| `AnalysisJobStarted` | Analysis Service | Analysis job initiated |
| `AnalysisJobProgress` | Analysis Service | Progress update |
| `AnalysisJobCompleted` | Analysis Service | Analysis finished successfully |
| `AnalysisJobFailed` | Analysis Service | Analysis encountered error |

### Alert Events

| Event | Publisher | Description |
|---|---|---|
| `ExtremeConditionDetected` | Data Ingestion Service | Threshold exceeded in reading |
| `StationOffline` | Data Ingestion Service | Station stopped reporting |
| `DataQualityIssue` | Data Ingestion Service | Quality control flagged data |
