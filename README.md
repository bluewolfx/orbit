# Orbit

High-performance distributed task orchestrator built in Go, utilizing gRPC for node communication and worker-pool patterns for concurrent job execution.

## Features

🚀 **Concurrency**: Worker pool pattern with goroutines to handle multiple jobs simultaneously  
📡 **Communication**: gRPC for efficient Manager-Worker communication  
💾 **Persistence**: PostgreSQL for job state storage and Redis for fast message queue  
📊 **Observability**: Prometheus metrics for tracking job success/failure rates  

## Architecture

Orbit consists of two main components:

1. **Manager**: 
   - REST API for job submission and status queries
   - gRPC server for worker communication
   - Job distribution and tracking
   - Prometheus metrics endpoint

2. **Worker**:
   - Worker pool with configurable goroutines
   - gRPC client for manager communication
   - Job execution with automatic status reporting
   - Heartbeat mechanism for health monitoring

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- PostgreSQL and Redis (if running locally)

### Running with Docker Compose

```bash
# Start all services (Manager, Workers, PostgreSQL, Redis)
docker-compose up --build

# Scale workers
docker-compose up --build --scale worker1=3
```

### Running Locally

1. **Start dependencies**:
```bash
# PostgreSQL
docker run -d --name postgres -e POSTGRES_DB=orbit -e POSTGRES_USER=orbit -e POSTGRES_PASSWORD=orbit -p 5432:5432 postgres:16-alpine

# Redis
docker run -d --name redis -p 6379:6379 redis:7-alpine
```

2. **Build and run Manager**:
```bash
go build -o bin/manager ./cmd/manager
./bin/manager
```

3. **Build and run Workers**:
```bash
go build -o bin/worker ./cmd/worker
./bin/worker -manager localhost:50051 -pool 5
```

## API Usage

### Submit a Job

```bash
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{"type": "compute", "payload": "Calculate PI"}'
```

Response:
```json
{
  "job_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### Get Job Status

```bash
curl http://localhost:8080/api/jobs/550e8400-e29b-41d4-a716-446655440000
```

### List Jobs

```bash
curl http://localhost:8080/api/jobs?limit=10
```

### Health Check

```bash
curl http://localhost:8080/health
```

## Job Types

Orbit supports multiple job types:

- `compute`: CPU-intensive computations (1-3s)
- `data-processing`: Data transformation tasks (2-5s)
- `ml-inference`: Machine learning inference (1-4s)
- `default`: Generic job processing (1-2s)

## Metrics

Prometheus metrics are available at `http://localhost:8080/metrics`:

- `orbit_jobs_total`: Total jobs processed (labeled by status)
- `orbit_jobs_in_progress`: Current jobs being processed
- `orbit_job_duration_seconds`: Job execution duration histogram
- `orbit_queue_length`: Current job queue length
- `orbit_active_workers`: Number of registered workers

## Configuration

Environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://orbit:orbit@localhost:5432/orbit?sslmode=disable` |
| `REDIS_URL` | Redis address | `localhost:6379` |
| `MANAGER_GRPC_PORT` | Manager gRPC port | `50051` |
| `MANAGER_HTTP_PORT` | Manager HTTP port | `8080` |
| `WORKER_POOL_SIZE` | Goroutines per worker | `10` |

## Development

### Generate protobuf code

```bash
protoc --go_out=. --go_opt=paths=source_relative \
  --go-grpc_out=. --go-grpc_opt=paths=source_relative \
  api/proto/job.proto
```

### Build

```bash
go build -o bin/manager ./cmd/manager
go build -o bin/worker ./cmd/worker
```

## Project Structure

```
orbit/
├── api/
│   └── proto/          # gRPC protobuf definitions
├── cmd/
│   ├── manager/        # Manager application
│   └── worker/         # Worker application
├── internal/
│   ├── database/       # PostgreSQL integration
│   ├── redis/          # Redis queue
│   ├── manager/        # Manager logic (gRPC & HTTP)
│   ├── worker/         # Worker pool implementation
│   ├── models/         # Data models
│   └── metrics/        # Prometheus metrics
├── pkg/
│   └── config/         # Configuration management
└── docker-compose.yml  # Docker orchestration
```

## License

MIT
