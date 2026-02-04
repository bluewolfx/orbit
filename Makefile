.PHONY: help proto build run-manager run-worker docker-up docker-down clean test

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

proto: ## Generate protobuf code
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/job.proto

build: proto ## Build binaries
	go build -o bin/manager ./cmd/manager
	go build -o bin/worker ./cmd/worker

run-manager: ## Run manager locally
	./bin/manager

run-worker: ## Run worker locally
	./bin/worker -manager localhost:50051 -pool 5

docker-up: ## Start all services with docker-compose
	docker-compose up --build

docker-down: ## Stop all services
	docker-compose down

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f api/proto/*.pb.go

test: ## Run tests
	go test -v ./...

deps: ## Install dependencies
	go mod download
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
