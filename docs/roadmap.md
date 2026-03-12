# Roadmap

## Development Phases

### Phase 1: Core Infrastructure
1. Set up AWS infrastructure with CDK (VPC, subnets, security groups)
2. Configure AWS API Gateway with Cognito authentication
3. Create Data Ingestion Service with REST API for readings and stations
4. Set up DynamoDB tables for sensor data and station metadata
5. Implement external API polling (NOAA, OpenWeatherMap)
6. Seed DynamoDB stations table with initial NOAA station set; poller reads station list from DB rather than config, so stations can be managed via the API without redeploying
7. Store OWM API key in AWS Secrets Manager; inject into ECS task via secrets (not plaintext env)

### Phase 2: Data Processing & Analysis
1. Set up SQS/SNS message queues and event topics
2. Implement data transformation and validation in Data Ingestion Service
3. Create Analysis Service with job management
4. Set up RDS PostgreSQL for analysis results
5. Add Redis/ElastiCache caching for the Analysis Service

> **Design decision (revisit in Phase 2):** The current poller accumulates a local copy of NOAA/OWM data primarily to build historical depth for trend analysis and to enable station offline detection. Once the Analysis Service exists, reconsider whether ingestion should remain continuous or shift to query-driven — i.e., the Analysis Service fetches the relevant historical window from upstream APIs at analysis time and caches the normalized result. Query-driven removes the polling overhead and avoids storing data that upstream sources already keep, but requires the Analysis Service to own the fetch-and-normalize logic.

### Phase 3: Frontend & User Experience
1. Create Next.js web dashboard
2. Implement data visualization components (charts, maps)
3. Add WebSocket support for real-time updates
4. Build analysis job management UI
5. Integrate Cognito authentication in the frontend
6. Station proximity filtering: use coordinates from `GET /api/stations` to sort/filter by distance to user location client-side — no new backend endpoint needed

### Phase 4: Advanced Features
1. Implement automated quality control in Data Ingestion Service
2. Add scheduled report generation
3. Create alert system (extreme conditions, station offline)
4. Build data export functionality
5. Add API rate limiting and CloudWatch monitoring

## User Stories & Use Cases

### Climate Scientists
- View current readings from all stations
- Start analyses of temperature trends over time periods
- Export analyzed data for research
- Set up alerts for extreme conditions

### Station Technicians
- View health status of maintained stations
- Receive alerts when stations stop reporting
- Update station metadata after maintenance

### Public Data Consumers
- Access historical climate data through API
- Subscribe to real-time weather updates

### Automated Systems
- Ingest data from weather stations regularly
- Perform automated quality control
- Generate scheduled reports

## Next Steps
1. Set up AWS account and initial CDK infrastructure
2. Create Data Ingestion Service scaffolding (Go)
3. Create Analysis Service scaffolding (Python)
4. Scaffold Next.js Web Dashboard
5. Implement first API endpoints (readings CRUD)
