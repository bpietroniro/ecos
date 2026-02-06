# Infrastructure & Security

## Deployment

- **Infrastructure as Code**: TypeScript CDK for all AWS resources
- **Containerization**: Docker for Data Ingestion Service (Go) and Analysis Service (Python)
- **Compute**: ECS Fargate (serverless containers) for backend services
- **Frontend hosting**: S3 + CloudFront for static Next.js export, or ECS for SSR
- **CI/CD**: GitHub Actions pipelines for build, test, and deploy
- **Monitoring**: CloudWatch for logs, metrics, and alerting

## Security

- **Authentication**: AWS Cognito (user pools for frontend, API keys for external integrations)
- **Authorization**: Cognito groups / IAM policies for role-based access
- **Transport encryption**: HTTPS everywhere via API Gateway and CloudFront
- **Data at rest**: DynamoDB and RDS encryption enabled by default
- **Rate limiting**: API Gateway throttling per API key / user
- **Security audits**: Regular dependency scanning and infrastructure review

## External Data Sources

Used for development, testing, and production data ingestion.

| Source | Description |
|---|---|
| [NOAA Climate Data Online API](https://www.ncdc.noaa.gov/cdo-web/webservices/v2) | Historical and current US climate data |
| [OpenWeatherMap API](https://openweathermap.org/api) | Global weather data and forecasts |
| [NASA GISS Surface Temperature Analysis](https://data.giss.nasa.gov/) | Global surface temperature datasets |
| [WorldBank Climate Data API](https://datahelpdesk.worldbank.org/knowledgebase/articles/902061) | Country-level climate indicators |
