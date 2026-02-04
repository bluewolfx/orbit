package models

import "time"

type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type Job struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Payload   string    `json:"payload"`
	Status    JobStatus `json:"status"`
	WorkerID  string    `json:"worker_id,omitempty"`
	Result    string    `json:"result,omitempty"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Worker struct {
	ID         string    `json:"id"`
	Address    string    `json:"address"`
	Capacity   int       `json:"capacity"`
	LastSeen   time.Time `json:"last_seen"`
	RegisteredAt time.Time `json:"registered_at"`
}
