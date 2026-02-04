#!/bin/bash

# Test script for Orbit distributed job processing system

echo "=== Orbit Test Script ==="
echo ""

# Submit jobs
echo "1. Submitting jobs..."
for i in {1..5}; do
  RESULT=$(curl -s -X POST http://localhost:8080/api/jobs \
    -H "Content-Type: application/json" \
    -d "{\"type\": \"compute\", \"payload\": \"Test Job $i\"}")
  JOB_ID=$(echo $RESULT | grep -o '"job_id":"[^"]*"' | cut -d'"' -f4)
  echo "   Submitted: Job $i (ID: $JOB_ID)"
done

echo ""
echo "2. Waiting for jobs to complete..."
sleep 10

echo ""
echo "3. Checking health..."
curl -s http://localhost:8080/health | jq

echo ""
echo "4. Viewing recent jobs..."
curl -s http://localhost:8080/api/jobs?limit=5 | jq '.[] | {id: .id, type: .type, status: .status, worker_id: .worker_id, result: .result}'

echo ""
echo "5. Prometheus metrics:"
curl -s http://localhost:8080/metrics 2>/dev/null | grep -E "^orbit_"

echo ""
echo "=== Test Complete ==="
