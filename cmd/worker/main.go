package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bluewolfx/Orbit/internal/worker"
)

func main() {
	managerAddr := flag.String("manager", "localhost:50051", "Manager gRPC address")
	workerID := flag.String("id", "", "Worker ID (auto-generated if not provided)")
	poolSize := flag.Int("pool", 5, "Worker pool size")
	flag.Parse()

	log.Printf("Starting Orbit Worker (ID: %s, Pool Size: %d)...", *workerID, *poolSize)

	w, err := worker.NewWorker(*workerID, *managerAddr, *poolSize)
	if err != nil {
		log.Fatalf("Failed to create worker: %v", err)
	}

	ctx := context.Background()
	if err := w.Start(ctx); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down worker...")
	w.Stop()
}
