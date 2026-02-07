# Data Ingestion Service

Go service that accepts sensor readings and station metadata via REST, stores them in DynamoDB, and publishes events to SNS.

## Prerequisites

- **Go 1.23+**
- **Docker** — for DynamoDB Local and container builds
- **AWS CLI** — for local table setup (no real AWS credentials needed)

## Local Development

### 1. Start DynamoDB Local

```bash
make dynamo-local    # starts DynamoDB Local in Docker on port 8000
make dynamo-setup    # creates tables with matching CDK schema
```

If you don't have real AWS credentials configured, set dummy ones:

```bash
export AWS_ACCESS_KEY_ID=fake
export AWS_SECRET_ACCESS_KEY=fake
export AWS_DEFAULT_REGION=us-east-1
```

### 2. Run the service

```bash
make run
```

Starts on port 8080 with DynamoDB Local. SNS runs in no-op mode (logs events) when topic ARNs are not set.

See `.env.example` for all configuration options.

### 3. Smoke test

```bash
# Health check
curl http://localhost:8080/healthz

# Register a station
curl -X POST http://localhost:8080/api/stations \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Test Station",
    "type": "weather",
    "location": {"latitude": 40.71, "longitude": -74.00, "elevation": 10},
    "instruments": ["thermometer"]
  }'

# Submit a reading (replace <stationId> with the id from above)
curl -X POST http://localhost:8080/api/readings \
  -H 'Content-Type: application/json' \
  -d '{
    "stationId": "<stationId>",
    "timestamp": "2024-01-15T10:30:00Z",
    "readingType": "temperature",
    "value": 22.5,
    "unit": "celsius",
    "location": {"latitude": 40.71, "longitude": -74.00, "elevation": 10}
  }'

# Query readings by station
curl http://localhost:8080/api/readings/station/<stationId>

# List all readings
curl http://localhost:8080/api/readings

# List all stations
curl http://localhost:8080/api/stations
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/healthz` | Health check |
| GET | `/api/readings` | List readings (query: `stationId`, `readingType`, `startTime`, `endTime`, `limit`) |
| GET | `/api/readings/station/{stationId}` | Readings for a station |
| POST | `/api/readings` | Submit a reading |
| POST | `/api/readings/batch` | Submit readings in batch (max 100) |
| GET | `/api/stations` | List all stations |
| POST | `/api/stations` | Register a station |
| PUT | `/api/stations/{id}` | Update station metadata |
| DELETE | `/api/stations/{id}` | Deregister a station |

## Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Build the binary |
| `make run` | Build and run locally |
| `make test` | Run all tests |
| `make vet` | Run go vet |
| `make dynamo-local` | Start DynamoDB Local in Docker |
| `make dynamo-setup` | Create local DynamoDB tables |
| `make docker-build` | Build the Docker image |
| `make clean` | Remove build artifacts |
