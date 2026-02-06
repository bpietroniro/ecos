# Architecture

## Service Overview

The platform uses 3 custom services plus AWS-managed API Gateway, simplified from the original 7-service design for cost and operational efficiency.

### Data Ingestion Service (Go)

Consolidates the responsibilities of the original Sensor Data Service, Real-time Gateway, and Data Transform Service.

- **External polling**: Periodically fetches data from NOAA, OpenWeatherMap, NASA GISS, and WorldBank APIs
- **Data ingestion**: Accepts sensor readings via REST API (direct submissions and batch uploads)
- **Data transformation**: Normalizes and validates incoming data (unit conversion, quality checks)
- **Storage**: Writes validated sensor readings and station metadata to DynamoDB
- **Event publishing**: Publishes events to SQS/SNS on data ingestion, validation, and quality issues
- **Station management**: CRUD operations for station metadata

### Analysis Service (Python)

Consolidates the responsibilities of the original Analysis Processing Service and Analysis Cache Service.

- **Job management**: Creates, tracks, and executes long-running analysis jobs (temperature trends, anomaly detection, correlation analysis)
- **Caching**: Manages Redis/ElastiCache for frequently accessed metrics and analysis results
- **Results storage**: Persists completed analysis results to RDS PostgreSQL
- **Event consumption**: Listens for data ingestion events to trigger automated analyses
- **Cache invalidation**: Invalidates cached results when underlying data changes

### Web Dashboard (Next.js)

- **Data visualization**: Charts and maps for sensor readings and analysis results
- **Job management**: UI for creating, monitoring, and viewing analysis jobs
- **Real-time updates**: WebSocket connections for live data feeds
- **Station monitoring**: Health status and alerts for weather stations
- **Authentication**: Integrates with AWS Cognito for user login

### AWS API Gateway (Managed)

- **Request routing**: Routes client requests to Data Ingestion Service and Analysis Service
- **Authentication**: Integrates with AWS Cognito for JWT validation
- **Rate limiting**: Throttles requests per API key / user
- **No custom code**: Fully managed, configured via CDK

## Data Flows

### Data Ingestion
```
External Source → Data Ingestion Service → (validate & transform) → DynamoDB
                                        → SQS/SNS (NewDataReceived event)
                                        → Analysis Service (triggered by event)
```

### Analysis Request
```
Client → API Gateway → Analysis Service → (check Redis cache)
                                        → Data Ingestion Service (fetch readings)
                                        → Process → Redis cache + RDS PostgreSQL
                                        → Client
```

### Real-time Updates
```
Data Ingestion Service → SQS/SNS → Web Dashboard (WebSocket) → Client
```

## Service Communication

| From | To | Method | Use Case |
|---|---|---|---|
| Client | API Gateway | REST (HTTPS) | All client requests |
| API Gateway | Data Ingestion Service | REST | Reading/station CRUD |
| API Gateway | Analysis Service | REST | Job management |
| Data Ingestion Service | Analysis Service | gRPC | Direct data fetches |
| Data Ingestion Service | SQS/SNS | Async messaging | Event publishing |
| Analysis Service | SQS/SNS | Async messaging | Event consumption |
| Analysis Service | Redis | TCP | Cache reads/writes |
| Analysis Service | RDS PostgreSQL | TCP | Results persistence |
| Data Ingestion Service | DynamoDB | AWS SDK | Sensor data storage |

## AWS Infrastructure Mapping

| Service | AWS Resource |
|---|---|
| Data Ingestion Service | ECS Fargate |
| Analysis Service | ECS Fargate |
| Web Dashboard | S3 + CloudFront (static export) |
| API Gateway | AWS API Gateway |
| Sensor data | DynamoDB (pay-per-request) |
| Analysis results | RDS PostgreSQL |
| Cache | ElastiCache (Redis) |
| Message broker | SQS/SNS |
| Auth | Cognito |
| Monitoring | CloudWatch |
