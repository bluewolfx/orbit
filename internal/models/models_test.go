package models

import (
	"testing"
	"time"
)

func TestJobStatusConstants(t *testing.T) {
	statuses := []JobStatus{
		JobStatusPending,
		JobStatusRunning,
		JobStatusCompleted,
		JobStatusFailed,
	}

	if len(statuses) != 4 {
		t.Error("Expected 4 job statuses")
	}

	if JobStatusPending != "pending" {
		t.Errorf("Expected pending, got %s", JobStatusPending)
	}
}

func TestJobCreation(t *testing.T) {
	job := &Job{
		ID:        "test-id",
		Type:      "compute",
		Payload:   "test payload",
		Status:    JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if job.ID != "test-id" {
		t.Errorf("Expected job ID to be test-id, got %s", job.ID)
	}

	if job.Status != JobStatusPending {
		t.Errorf("Expected status to be pending, got %s", job.Status)
	}
}

func TestWorkerCreation(t *testing.T) {
	worker := &Worker{
		ID:           "worker-1",
		Address:      "localhost:50051",
		Capacity:     5,
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
	}

	if worker.ID != "worker-1" {
		t.Errorf("Expected worker ID to be worker-1, got %s", worker.ID)
	}

	if worker.Capacity != 5 {
		t.Errorf("Expected capacity to be 5, got %d", worker.Capacity)
	}
}
