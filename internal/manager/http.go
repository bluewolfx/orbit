package manager

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	manager *Manager
}

func NewHTTPServer(manager *Manager) *HTTPServer {
	return &HTTPServer{manager: manager}
}

type SubmitJobRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type SubmitJobResponse struct {
	JobID string `json:"job_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (s *HTTPServer) SetupRoutes(router *mux.Router) {
	router.HandleFunc("/api/jobs", s.submitJob).Methods("POST")
	router.HandleFunc("/api/jobs/{id}", s.getJob).Methods("GET")
	router.HandleFunc("/api/jobs", s.listJobs).Methods("GET")
	router.HandleFunc("/health", s.health).Methods("GET")
}

func (s *HTTPServer) submitJob(w http.ResponseWriter, r *http.Request) {
	var req SubmitJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Type == "" || req.Payload == "" {
		respondError(w, "Type and payload are required", http.StatusBadRequest)
		return
	}

	jobID, err := s.manager.SubmitJob(r.Context(), req.Type, req.Payload)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, SubmitJobResponse{JobID: jobID}, http.StatusCreated)
}

func (s *HTTPServer) getJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["id"]

	job, err := s.manager.GetJob(r.Context(), jobID)
	if err != nil {
		respondError(w, "Job not found", http.StatusNotFound)
		return
	}

	respondJSON(w, job, http.StatusOK)
}

func (s *HTTPServer) listJobs(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	jobs, err := s.manager.ListJobs(r.Context(), limit)
	if err != nil {
		respondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	respondJSON(w, jobs, http.StatusOK)
}

func (s *HTTPServer) health(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, map[string]interface{}{
		"status":         "ok",
		"active_workers": s.manager.GetActiveWorkerCount(),
	}, http.StatusOK)
}

func respondJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, message string, status int) {
	respondJSON(w, ErrorResponse{Error: message}, status)
}
