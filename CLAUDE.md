# Climate Data Analysis Platform

Microservices-based platform that ingests climate data from weather stations, satellites, and ocean sensors, processes and analyzes it, and serves real-time insights through a web dashboard. Designed as a cost-conscious side project on AWS.

## Tech Stack

| Component | Choice |
|---|---|
| Sensor data storage | DynamoDB (pay-per-request) |
| Analysis results storage | RDS PostgreSQL |
| Caching | Redis / ElastiCache |
| Authentication | AWS Cognito |
| API Gateway | AWS API Gateway (managed) |
| Message broker | AWS SQS/SNS |
| Backend - data ingestion | Go |
| Backend - analysis | Python |
| Frontend | Next.js, React, TypeScript, TailwindCSS |
| IaC | TypeScript CDK |
| CI/CD | GitHub Actions |
| Monitoring | CloudWatch |

## Services

| Service | Language | Responsibility |
|---|---|---|
| **Data Ingestion Service** | Go | Polls external APIs, ingests sensor data, stores to DynamoDB, normalizes/transforms data, publishes events to SQS/SNS |
| **Analysis Service** | Python | Runs analysis jobs, manages caching (Redis), stores results in RDS PostgreSQL |
| **Web Dashboard** | Next.js | Data visualization, job management UI, real-time updates via WebSocket |
| **AWS API Gateway** | Managed | Routes requests, integrates with Cognito for auth |

## Key Patterns

- **Event-Driven**: Services communicate through SQS/SNS message queues
- **Cache-Aside**: Cache expensive computations in Redis with appropriate TTLs
- **CQRS**: Separate read/write paths for analysis service
- **Cost Minimization**: Pay-per-request DynamoDB, managed API Gateway, right-sized infrastructure

## Development Guidelines

### Service Communication
- REST for synchronous client-facing APIs
- gRPC for inter-service communication
- SQS/SNS for async operations and events
- Implement retry logic and circuit breakers

### Error Handling
- Return appropriate HTTP status codes with detailed error messages
- Log errors with context
- Use dead letter queues for failed messages

### Caching
- Cache expensive computations with TTLs based on data volatility
- Invalidate cache on data updates
- Monitor cache hit rates

### Testing
- Unit tests for business logic
- Integration tests for API endpoints
- End-to-end tests for critical flows
- Load testing for performance validation

## Reference Docs

- [Architecture & Data Flows](docs/architecture.md)
- [API Specification](docs/api-spec.md)
- [Data Models](docs/data-models.md)
- [Roadmap & User Stories](docs/roadmap.md)
- [Infrastructure & Security](docs/infrastructure.md)
