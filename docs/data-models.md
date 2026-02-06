# Data Models

## SensorReading

Stored in DynamoDB by the Data Ingestion Service.

```typescript
{
  id: string
  stationId: string
  timestamp: DateTime
  readingType: 'temperature' | 'humidity' | 'pressure' | 'wind_speed' | ...
  value: number
  unit: string
  quality: {
    status: 'raw' | 'validated' | 'flagged' | 'rejected'
    flags: string[]
  }
  location: {
    latitude: number
    longitude: number
    elevation: number
  }
  metadata: Record<string, any>
}
```

## Station

Stored in DynamoDB by the Data Ingestion Service.

```typescript
{
  id: string
  name: string
  type: 'weather' | 'ocean' | 'satellite'
  location: {
    latitude: number
    longitude: number
    elevation: number
  }
  status: 'active' | 'inactive' | 'maintenance'
  instruments: string[]
  lastReportedAt: DateTime
  metadata: Record<string, any>
}
```

## AnalysisJob

Stored in RDS PostgreSQL by the Analysis Service.

```typescript
{
  id: string
  type: 'temperature_trend' | 'anomaly_detection' | 'correlation' | ...
  status: 'pending' | 'in_progress' | 'completed' | 'failed'
  parameters: {
    timeRange: { start: DateTime, end: DateTime }
    stations: string[]
    metrics: string[]
    methodology: string
  }
  progress: number
  currentStep: string
  results?: any
  createdAt: DateTime
  completedAt?: DateTime
  error?: string
}
```
