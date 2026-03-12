# ecos

A microservices-based climate data analysis platform that ingests sensor readings from weather stations, satellites, and ocean sensors, processes them through an analysis pipeline, and surfaces insights via a web dashboard.

## Services

| Service | Language | Description |
|---|---|---|
| `services/ingestion` | Go | Polls external APIs, stores readings to DynamoDB, publishes events to SNS |
| Analysis Service (TODO) | Python | Runs analysis jobs, caches results, stores to PostgreSQL |
| Web Dashboard (TODO) | Next.js | Data visualization and job management UI |

## Infrastructure

Managed with AWS CDK (TypeScript). Key resources:

- **DynamoDB** — sensor readings and station metadata (pay-per-request)
- **RDS PostgreSQL** — analysis results
- **ElastiCache Redis** — computation caching
- **SQS/SNS** — async event bus between services
- **API Gateway + Cognito** — request routing and auth
- **ECS Fargate** — service hosting

IaC lives in [`infra/`](infra/).

## Docs

- [Architecture & Data Flows](docs/architecture.md)
- [API Specification](docs/api-spec.md)
- [Data Models](docs/data-models.md)
- [Infrastructure & Security](docs/infrastructure.md)
- [Roadmap](docs/roadmap.md)
