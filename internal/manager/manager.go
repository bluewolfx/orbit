package manager

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bluewolfx/Orbit/internal/database"
	"github.com/bluewolfx/Orbit/internal/metrics"
	"github.com/bluewolfx/Orbit/internal/models"
	"github.com/bluewolfx/Orbit/internal/redis"
	"github.com/google/uuid"
)

type Manager struct {
	db      *database.Database
	queue   *redis.Queue
	workers map[string]*models.Worker
	mu      sync.RWMutex
}

func NewManager(db *database.Database, queue *redis.Queue) *Manager {
	return &Manager{
		db:      db,
		queue:   queue,
		workers: make(map[string]*models.Worker),
	}
}

func (m *Manager) SubmitJob(ctx context.Context, jobType, payload string) (string, error) {
	jobID := uuid.New().String()
	
	job := &models.Job{
		ID:        jobID,
		Type:      jobType,
		Payload:   payload,
		Status:    models.JobStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.db.CreateJob(ctx, job); err != nil {
		return "", fmt.Errorf("failed to create job: %w", err)
	}

	if err := m.queue.EnqueueJob(ctx, jobID); err != nil {
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	log.Printf("Job submitted: %s (type: %s)", jobID, jobType)
	return jobID, nil
}

func (m *Manager) GetJob(ctx context.Context, id string) (*models.Job, error) {
	return m.db.GetJob(ctx, id)
}

func (m *Manager) ListJobs(ctx context.Context, limit int) ([]*models.Job, error) {
	return m.db.ListJobs(ctx, limit)
}

func (m *Manager) GetNextJob(ctx context.Context, workerID string) (*models.Job, error) {
	jobID, err := m.queue.DequeueJob(ctx)
	if err != nil {
		return nil, err
	}

	job, err := m.db.GetJob(ctx, jobID)
	if err != nil {
		return nil, err
	}

	// Update job status to running
	if err := m.db.UpdateJobStatus(ctx, jobID, models.JobStatusRunning, workerID, "", ""); err != nil {
		return nil, err
	}

	job.Status = models.JobStatusRunning
	job.WorkerID = workerID
	
	metrics.RecordJobStarted()
	log.Printf("Job %s assigned to worker %s", jobID, workerID)
	
	return job, nil
}

func (m *Manager) UpdateJobStatus(ctx context.Context, jobID, workerID string, status models.JobStatus, result, errorMsg string) error {
	if err := m.db.UpdateJobStatus(ctx, jobID, status, workerID, result, errorMsg); err != nil {
		return err
	}

	metrics.RecordJobFinished()
	
	if status == models.JobStatusCompleted {
		metrics.RecordJobCompleted()
		log.Printf("Job %s completed by worker %s", jobID, workerID)
	} else if status == models.JobStatusFailed {
		metrics.RecordJobFailed()
		log.Printf("Job %s failed on worker %s: %s", jobID, workerID, errorMsg)
	}

	return nil
}

func (m *Manager) RegisterWorker(ctx context.Context, worker *models.Worker) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.workers[worker.ID] = worker
	
	if err := m.db.RegisterWorker(ctx, worker); err != nil {
		return err
	}

	metrics.ActiveWorkers.Set(float64(len(m.workers)))
	log.Printf("Worker registered: %s (address: %s, capacity: %d)", worker.ID, worker.Address, worker.Capacity)
	
	return nil
}

func (m *Manager) UpdateWorkerHeartbeat(ctx context.Context, workerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if worker, exists := m.workers[workerID]; exists {
		worker.LastSeen = time.Now()
		return m.db.UpdateWorkerHeartbeat(ctx, workerID)
	}

	return fmt.Errorf("worker not found: %s", workerID)
}

func (m *Manager) GetActiveWorkerCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.workers)
}

func (m *Manager) StartMetricsUpdater(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			queueLen, err := m.queue.QueueLength(ctx)
			if err != nil {
				log.Printf("Failed to get queue length: %v", err)
				continue
			}
			metrics.QueueLength.Set(float64(queueLen))
		}
	}
}
