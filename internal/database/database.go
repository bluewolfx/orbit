package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
	"github.com/bluewolfx/Orbit/internal/models"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(databaseURL string) (*Database, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &Database{db: db}
	if err := d.initSchema(); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return d, nil
}

func (d *Database) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS jobs (
		id VARCHAR(255) PRIMARY KEY,
		type VARCHAR(255) NOT NULL,
		payload TEXT NOT NULL,
		status VARCHAR(50) NOT NULL,
		worker_id VARCHAR(255),
		result TEXT,
		error TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
	CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs(created_at);

	CREATE TABLE IF NOT EXISTS workers (
		id VARCHAR(255) PRIMARY KEY,
		address VARCHAR(255) NOT NULL,
		capacity INTEGER NOT NULL,
		last_seen TIMESTAMP NOT NULL,
		registered_at TIMESTAMP NOT NULL
	);
	`

	_, err := d.db.Exec(schema)
	return err
}

func (d *Database) CreateJob(ctx context.Context, job *models.Job) error {
	query := `
		INSERT INTO jobs (id, type, payload, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := d.db.ExecContext(ctx, query, job.ID, job.Type, job.Payload, job.Status, job.CreatedAt, job.UpdatedAt)
	return err
}

func (d *Database) GetJob(ctx context.Context, id string) (*models.Job, error) {
	query := `
		SELECT id, type, payload, status, worker_id, result, error, created_at, updated_at
		FROM jobs WHERE id = $1
	`
	
	job := &models.Job{}
	var workerID, result, errorMsg sql.NullString
	
	err := d.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.Type, &job.Payload, &job.Status,
		&workerID, &result, &errorMsg,
		&job.CreatedAt, &job.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}

	if workerID.Valid {
		job.WorkerID = workerID.String
	}
	if result.Valid {
		job.Result = result.String
	}
	if errorMsg.Valid {
		job.Error = errorMsg.String
	}

	return job, nil
}

func (d *Database) UpdateJobStatus(ctx context.Context, jobID string, status models.JobStatus, workerID, result, errorMsg string) error {
	query := `
		UPDATE jobs 
		SET status = $1, worker_id = $2, result = $3, error = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := d.db.ExecContext(ctx, query, status, workerID, result, errorMsg, time.Now(), jobID)
	return err
}

func (d *Database) ListJobs(ctx context.Context, limit int) ([]*models.Job, error) {
	query := `
		SELECT id, type, payload, status, worker_id, result, error, created_at, updated_at
		FROM jobs ORDER BY created_at DESC LIMIT $1
	`
	
	rows, err := d.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*models.Job
	for rows.Next() {
		job := &models.Job{}
		var workerID, result, errorMsg sql.NullString
		
		if err := rows.Scan(
			&job.ID, &job.Type, &job.Payload, &job.Status,
			&workerID, &result, &errorMsg,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if workerID.Valid {
			job.WorkerID = workerID.String
		}
		if result.Valid {
			job.Result = result.String
		}
		if errorMsg.Valid {
			job.Error = errorMsg.String
		}

		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

func (d *Database) RegisterWorker(ctx context.Context, worker *models.Worker) error {
	query := `
		INSERT INTO workers (id, address, capacity, last_seen, registered_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			address = EXCLUDED.address,
			capacity = EXCLUDED.capacity,
			last_seen = EXCLUDED.last_seen
	`
	_, err := d.db.ExecContext(ctx, query, worker.ID, worker.Address, worker.Capacity, worker.LastSeen, worker.RegisteredAt)
	return err
}

func (d *Database) UpdateWorkerHeartbeat(ctx context.Context, workerID string) error {
	query := `UPDATE workers SET last_seen = $1 WHERE id = $2`
	_, err := d.db.ExecContext(ctx, query, time.Now(), workerID)
	return err
}

func (d *Database) Close() error {
	return d.db.Close()
}
