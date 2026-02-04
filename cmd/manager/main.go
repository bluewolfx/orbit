package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	pb "github.com/bluewolfx/Orbit/api/proto"
	"github.com/bluewolfx/Orbit/internal/database"
	"github.com/bluewolfx/Orbit/internal/manager"
	"github.com/bluewolfx/Orbit/internal/redis"
	"github.com/bluewolfx/Orbit/pkg/config"
)

func main() {
	log.Println("Starting Orbit Manager...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	// Initialize database
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected")

	// Initialize Redis queue
	queue, err := redis.NewQueue(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to initialize Redis queue: %v", err)
	}
	defer queue.Close()
	log.Println("Redis queue connected")

	// Create manager
	mgr := manager.NewManager(db, queue)

	// Start metrics updater
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go mgr.StartMetricsUpdater(ctx)

	// Start gRPC server
	grpcServer := grpc.NewServer()
	pb.RegisterJobServiceServer(grpcServer, manager.NewGRPCServer(mgr))

	grpcListener, err := net.Listen("tcp", ":"+cfg.ManagerGRPCPort)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port: %v", err)
	}

	go func() {
		log.Printf("gRPC server listening on :%s", cfg.ManagerGRPCPort)
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server
	router := mux.NewRouter()
	httpServer := manager.NewHTTPServer(mgr)
	httpServer.SetupRoutes(router)
	
	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler())

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.ManagerHTTPPort)
		if err := http.ListenAndServe(":"+cfg.ManagerHTTPPort, router); err != nil {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down manager...")
	grpcServer.GracefulStop()
}
