package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	JobsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "orbit_jobs_total",
			Help: "Total number of jobs processed",
		},
		[]string{"status"},
	)

	JobsInProgress = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "orbit_jobs_in_progress",
			Help: "Number of jobs currently in progress",
		},
	)

	JobDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "orbit_job_duration_seconds",
			Help:    "Duration of job execution in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type"},
	)

	QueueLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "orbit_queue_length",
			Help: "Current length of the job queue",
		},
	)

	ActiveWorkers = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "orbit_active_workers",
			Help: "Number of active workers",
		},
	)
)

func RecordJobCompleted() {
	JobsTotal.WithLabelValues("completed").Inc()
}

func RecordJobFailed() {
	JobsTotal.WithLabelValues("failed").Inc()
}

func RecordJobStarted() {
	JobsInProgress.Inc()
}

func RecordJobFinished() {
	JobsInProgress.Dec()
}
