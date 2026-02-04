# Orbit System Demo

## Quick Start Guide

### 1. Start the System
```bash
# Using Docker Compose (recommended)
docker-compose up --build

# Or manually
docker run -d --name postgres -e POSTGRES_DB=orbit -e POSTGRES_USER=orbit -e POSTGRES_PASSWORD=orbit -p 5432:5432 postgres:16-alpine
docker run -d --name redis -p 6379:6379 redis:7-alpine
go build -o bin/manager ./cmd/manager && ./bin/manager &
go build -o bin/worker ./cmd/worker && ./bin/worker -manager localhost:50051 -pool 5
```

### 2. Submit Jobs
```bash
# Submit a compute job
curl -X POST http://localhost:8080/api/jobs \
  -H "Content-Type: application/json" \
  -d '{"type": "compute", "payload": "Calculate PI to 1000 digits"}'

# Response: {"job_id":"550e8400-e29b-41d4-a716-446655440000"}
```

### 3. Check Status
```bash
# Get job status
curl http://localhost:8080/api/jobs/550e8400-e29b-41d4-a716-446655440000 | jq

# List all jobs
curl http://localhost:8080/api/jobs | jq

# Check health
curl http://localhost:8080/health | jq
```

### 4. View Metrics
```bash
# Prometheus metrics
curl http://localhost:8080/metrics | grep orbit_
```

## System Flow

```
1. Client submits job via REST API
   └─> POST /api/jobs {"type": "compute", "payload": "data"}

2. Manager receives job
   ├─> Creates job record in PostgreSQL
   ├─> Enqueues job ID in Redis
   └─> Returns job_id to client

3. Worker polls for jobs via gRPC
   ├─> Sends RequestJob to manager
   └─> Manager dequeues from Redis and returns job

4. Worker processes job
   ├─> Executes in goroutine from pool
   ├─> Simulates work based on job type
   └─> Reports status via ReportJobStatus

5. Manager updates job
   ├─> Updates PostgreSQL with result/error
   ├─> Updates Prometheus metrics
   └─> Logs completion

6. Client can query job status
   └─> GET /api/jobs/{id}
```

## Example Session

```bash
# Terminal 1: Start Manager
./bin/manager
# 2026/02/04 02:18:39 Starting Orbit Manager...
# 2026/02/04 02:18:39 Database connected
# 2026/02/04 02:18:39 Redis queue connected
# 2026/02/04 02:18:39 gRPC server listening on :50051
# 2026/02/04 02:18:39 HTTP server listening on :8080

# Terminal 2: Start Worker
./bin/worker -manager localhost:50051 -pool 3 -id worker-1
# 2026/02/04 02:18:49 Worker worker-1 registered successfully
# 2026/02/04 02:18:49 Starting worker worker-1 with pool size 3

# Terminal 3: Submit Jobs
for i in {1..5}; do
  curl -X POST http://localhost:8080/api/jobs \
    -H "Content-Type: application/json" \
    -d "{\"type\": \"compute\", \"payload\": \"Job $i\"}"
done

# Terminal 4: Monitor Metrics
watch -n 1 'curl -s http://localhost:8080/metrics | grep orbit_'
```

## Job Types

| Type | Description | Duration |
|------|-------------|----------|
| `compute` | CPU-intensive computations | 1-3s |
| `data-processing` | Data transformation tasks | 2-5s |
| `ml-inference` | Machine learning inference | 1-4s |
| `default` | Generic job processing | 1-2s |

## Scaling

```bash
# Scale workers with docker-compose
docker-compose up --scale worker1=5

# Or start multiple workers manually
./bin/worker -manager localhost:50051 -pool 5 -id worker-1 &
./bin/worker -manager localhost:50051 -pool 5 -id worker-2 &
./bin/worker -manager localhost:50051 -pool 5 -id worker-3 &

# Check active workers
curl http://localhost:8080/health
# {"status":"ok","active_workers":3}
```

## Monitoring

Access Prometheus metrics at `http://localhost:8080/metrics`

Key metrics:
- `orbit_jobs_total{status="completed"}` - Successful jobs
- `orbit_jobs_total{status="failed"}` - Failed jobs  
- `orbit_jobs_in_progress` - Currently running
- `orbit_queue_length` - Jobs waiting
- `orbit_active_workers` - Registered workers

## Troubleshooting

### PostgreSQL connection failed
```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Test connection
psql postgres://orbit:orbit@localhost:5432/orbit -c "SELECT 1"
```

### Redis connection failed
```bash
# Check Redis is running
docker ps | grep redis

# Test connection
redis-cli -h localhost -p 6379 PING
```

### Worker not receiving jobs
```bash
# Check worker logs for registration
# Should see: "Worker worker-1 registered successfully"

# Check manager logs for job assignment
# Should see: "Job {id} assigned to worker {worker-id}"

# Verify gRPC connectivity
curl http://localhost:8080/health
```

## Clean Up

```bash
# Stop all services
docker-compose down

# Or stop manually
pkill -f bin/manager
pkill -f bin/worker
docker rm -f orbit-postgres orbit-redis
```
