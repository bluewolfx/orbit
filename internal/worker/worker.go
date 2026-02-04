package worker

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	pb "github.com/bluewolfx/Orbit/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Worker struct {
	id              string
	managerAddr     string
	poolSize        int
	client          pb.JobServiceClient
	conn            *grpc.ClientConn
	wg              sync.WaitGroup
	stopChan        chan struct{}
	heartbeatTicker *time.Ticker
}

func NewWorker(id, managerAddr string, poolSize int) (*Worker, error) {
	if id == "" {
		id = fmt.Sprintf("worker-%d", rand.Intn(10000))
	}

	return &Worker{
		id:          id,
		managerAddr: managerAddr,
		poolSize:    poolSize,
		stopChan:    make(chan struct{}),
	}, nil
}

func (w *Worker) Start(ctx context.Context) error {
	// Connect to manager
	conn, err := grpc.Dial(w.managerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to manager: %w", err)
	}
	w.conn = conn
	w.client = pb.NewJobServiceClient(conn)

	// Register with manager
	if err := w.register(ctx); err != nil {
		return fmt.Errorf("failed to register with manager: %w", err)
	}

	// Start heartbeat
	w.startHeartbeat(ctx)

	// Start worker pool
	log.Printf("Starting worker %s with pool size %d", w.id, w.poolSize)
	for i := 0; i < w.poolSize; i++ {
		w.wg.Add(1)
		go w.workerRoutine(ctx, i)
	}

	return nil
}

func (w *Worker) Stop() {
	log.Printf("Stopping worker %s", w.id)
	close(w.stopChan)
	if w.heartbeatTicker != nil {
		w.heartbeatTicker.Stop()
	}
	w.wg.Wait()
	if w.conn != nil {
		w.conn.Close()
	}
}

func (w *Worker) register(ctx context.Context) error {
	resp, err := w.client.RegisterWorker(ctx, &pb.WorkerInfo{
		WorkerId: w.id,
		Address:  w.id,
		Capacity: int32(w.poolSize),
	})
	if err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf("registration failed: %s", resp.Message)
	}

	log.Printf("Worker %s registered successfully", w.id)
	return nil
}

func (w *Worker) startHeartbeat(ctx context.Context) {
	w.heartbeatTicker = time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-w.stopChan:
				return
			case <-w.heartbeatTicker.C:
				_, err := w.client.Heartbeat(ctx, &pb.WorkerInfo{
					WorkerId: w.id,
				})
				if err != nil {
					log.Printf("Heartbeat failed: %v", err)
				}
			}
		}
	}()
}

func (w *Worker) workerRoutine(ctx context.Context, workerNum int) {
	defer w.wg.Done()
	log.Printf("Worker routine %d started", workerNum)

	for {
		select {
		case <-w.stopChan:
			log.Printf("Worker routine %d stopping", workerNum)
			return
		default:
			// Request job from manager
			job, err := w.client.RequestJob(ctx, &pb.JobRequest{
				WorkerId: w.id,
			})
			if err != nil {
				log.Printf("Failed to request job: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			if job.Id == "" {
				// No jobs available
				time.Sleep(2 * time.Second)
				continue
			}

			// Execute job
			log.Printf("Worker %s routine %d processing job %s (type: %s)", w.id, workerNum, job.Id, job.Type)
			w.executeJob(ctx, job)
		}
	}
}

func (w *Worker) executeJob(ctx context.Context, job *pb.Job) {
	startTime := time.Now()
	
	// Simulate job execution
	result, err := w.processJob(job)
	
	duration := time.Since(startTime)
	log.Printf("Job %s completed in %v", job.Id, duration)

	// Report status back to manager
	status := pb.JobStatus_COMPLETED
	errorMsg := ""
	if err != nil {
		status = pb.JobStatus_FAILED
		errorMsg = err.Error()
	}

	_, err = w.client.ReportJobStatus(ctx, &pb.JobStatusReport{
		JobId:    job.Id,
		WorkerId: w.id,
		Status:   status,
		Result:   result,
		Error:    errorMsg,
	})
	if err != nil {
		log.Printf("Failed to report job status: %v", err)
	}
}

func (w *Worker) processJob(job *pb.Job) (string, error) {
	// Simulate different job types
	switch job.Type {
	case "compute":
		// Simulate computation
		time.Sleep(time.Duration(rand.Intn(3)+1) * time.Second)
		return fmt.Sprintf("Computed result for: %s", job.Payload), nil
		
	case "data-processing":
		// Simulate data processing
		time.Sleep(time.Duration(rand.Intn(5)+2) * time.Second)
		return fmt.Sprintf("Processed data: %s", job.Payload), nil
		
	case "ml-inference":
		// Simulate ML inference
		time.Sleep(time.Duration(rand.Intn(4)+1) * time.Second)
		return fmt.Sprintf("Inference result for: %s", job.Payload), nil
		
	default:
		// Default job processing
		time.Sleep(time.Duration(rand.Intn(2)+1) * time.Second)
		return fmt.Sprintf("Processed: %s", job.Payload), nil
	}
}
