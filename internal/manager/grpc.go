package manager

import (
	"context"
	"time"

	pb "github.com/bluewolfx/Orbit/api/proto"
	"github.com/bluewolfx/Orbit/internal/models"
)

type GRPCServer struct {
	pb.UnimplementedJobServiceServer
	manager *Manager
}

func NewGRPCServer(manager *Manager) *GRPCServer {
	return &GRPCServer{manager: manager}
}

func (s *GRPCServer) RequestJob(ctx context.Context, req *pb.JobRequest) (*pb.Job, error) {
	job, err := s.manager.GetNextJob(ctx, req.WorkerId)
	if err != nil {
		// Return empty job if no jobs available
		return &pb.Job{}, nil
	}

	return &pb.Job{
		Id:        job.ID,
		Type:      job.Type,
		Payload:   job.Payload,
		CreatedAt: job.CreatedAt.Unix(),
	}, nil
}

func (s *GRPCServer) ReportJobStatus(ctx context.Context, req *pb.JobStatusReport) (*pb.JobStatusResponse, error) {
	var status models.JobStatus
	switch req.Status {
	case pb.JobStatus_COMPLETED:
		status = models.JobStatusCompleted
	case pb.JobStatus_FAILED:
		status = models.JobStatusFailed
	default:
		status = models.JobStatusRunning
	}

	err := s.manager.UpdateJobStatus(ctx, req.JobId, req.WorkerId, status, req.Result, req.Error)
	if err != nil {
		return &pb.JobStatusResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.JobStatusResponse{
		Success: true,
		Message: "Status updated successfully",
	}, nil
}

func (s *GRPCServer) RegisterWorker(ctx context.Context, req *pb.WorkerInfo) (*pb.RegistrationResponse, error) {
	worker := &models.Worker{
		ID:           req.WorkerId,
		Address:      req.Address,
		Capacity:     int(req.Capacity),
		LastSeen:     time.Now(),
		RegisteredAt: time.Now(),
	}

	err := s.manager.RegisterWorker(ctx, worker)
	if err != nil {
		return &pb.RegistrationResponse{
			Success: false,
			Message: err.Error(),
		}, err
	}

	return &pb.RegistrationResponse{
		Success: true,
		Message: "Worker registered successfully",
	}, nil
}

func (s *GRPCServer) Heartbeat(ctx context.Context, req *pb.WorkerInfo) (*pb.HeartbeatResponse, error) {
	err := s.manager.UpdateWorkerHeartbeat(ctx, req.WorkerId)
	if err != nil {
		return &pb.HeartbeatResponse{Success: false}, err
	}

	return &pb.HeartbeatResponse{Success: true}, nil
}
