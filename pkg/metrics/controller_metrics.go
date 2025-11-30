package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// PipelineRunsTotal tracks total number of pipeline runs by status
	PipelineRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_pipelineruns_total",
			Help: "Total number of pipeline runs by status",
		},
		[]string{"namespace", "status"},
	)

	// PipelineRunDuration tracks pipeline run duration in seconds
	PipelineRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "c8s_pipelinerun_duration_seconds",
			Help:    "Duration of pipeline runs in seconds",
			Buckets: []float64{10, 30, 60, 120, 300, 600, 1800, 3600}, // 10s to 1h
		},
		[]string{"namespace", "status"},
	)

	// ReconcileCount tracks number of reconciliation loops
	ReconcileCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_reconcile_total",
			Help: "Total number of reconciliation loops",
		},
		[]string{"controller", "result"},
	)

	// ReconcileDuration tracks reconciliation duration
	ReconcileDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "c8s_reconcile_duration_seconds",
			Help:    "Duration of reconciliation loops in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"controller"},
	)

	// JobsCreated tracks number of Jobs created
	JobsCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_jobs_created_total",
			Help: "Total number of Jobs created",
		},
		[]string{"namespace", "step"},
	)

	// JobsFailed tracks number of failed Jobs
	JobsFailed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_jobs_failed_total",
			Help: "Total number of failed Jobs",
		},
		[]string{"namespace", "step", "reason"},
	)

	// ActivePipelineRuns tracks currently active pipeline runs
	ActivePipelineRuns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "c8s_active_pipelineruns",
			Help: "Number of currently active pipeline runs",
		},
		[]string{"namespace", "phase"},
	)
)

func init() {
	// Register custom metrics with controller-runtime's registry
	metrics.Registry.MustRegister(
		PipelineRunsTotal,
		PipelineRunDuration,
		ReconcileCount,
		ReconcileDuration,
		JobsCreated,
		JobsFailed,
		ActivePipelineRuns,
	)
}
