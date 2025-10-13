package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	// PipelineRunsTotal tracks total number of PipelineRuns by phase and namespace
	PipelineRunsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_pipelineruns_total",
			Help: "Total number of PipelineRuns created",
		},
		[]string{"phase", "namespace"},
	)

	// PipelineRunDuration tracks duration of PipelineRuns in seconds
	PipelineRunDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "c8s_pipelineruns_duration_seconds",
			Help:    "Duration of PipelineRuns in seconds",
			Buckets: prometheus.ExponentialBuckets(10, 2, 10), // 10s, 20s, 40s, ..., 5120s (~85min)
		},
		[]string{"namespace", "config"},
	)

	// StepResourceUsageCPU tracks CPU usage per step
	StepResourceUsageCPU = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "c8s_pipeline_step_resource_usage_cpu_cores",
			Help: "CPU resource usage in cores per pipeline step",
		},
		[]string{"step", "namespace", "pipeline_run"},
	)

	// StepResourceUsageMemory tracks memory usage per step
	StepResourceUsageMemory = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "c8s_pipeline_step_resource_usage_memory_bytes",
			Help: "Memory resource usage in bytes per pipeline step",
		},
		[]string{"step", "namespace", "pipeline_run"},
	)

	// ActivePipelineRuns tracks currently running PipelineRuns
	ActivePipelineRuns = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "c8s_active_pipelineruns",
			Help: "Number of currently active (Running) PipelineRuns",
		},
		[]string{"namespace"},
	)

	// PendingSteps tracks steps waiting to execute
	PendingSteps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "c8s_pending_steps",
			Help: "Number of steps in Pending phase (waiting for resources)",
		},
		[]string{"namespace"},
	)

	// FailedSteps tracks failed steps
	FailedSteps = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_failed_steps_total",
			Help: "Total number of failed pipeline steps",
		},
		[]string{"step", "namespace"},
	)

	// JobCreationDuration tracks time to create Kubernetes Jobs
	JobCreationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "c8s_job_creation_duration_seconds",
			Help:    "Time taken to create Kubernetes Jobs for steps",
			Buckets: prometheus.LinearBuckets(0.1, 0.1, 10), // 0.1s to 1s
		},
		[]string{"namespace"},
	)

	// ReconcileErrors tracks reconciliation errors
	ReconcileErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "c8s_reconcile_errors_total",
			Help: "Total number of reconciliation errors",
		},
		[]string{"controller", "namespace"},
	)
)

// init registers all metrics with controller-runtime metrics registry
func init() {
	metrics.Registry.MustRegister(
		PipelineRunsTotal,
		PipelineRunDuration,
		StepResourceUsageCPU,
		StepResourceUsageMemory,
		ActivePipelineRuns,
		PendingSteps,
		FailedSteps,
		JobCreationDuration,
		ReconcileErrors,
	)
}

// RecordPipelineRunCreated increments PipelineRunsTotal counter
func RecordPipelineRunCreated(namespace, phase string) {
	PipelineRunsTotal.WithLabelValues(phase, namespace).Inc()
}

// RecordPipelineRunCompleted records completion metrics
func RecordPipelineRunCompleted(namespace, config string, durationSeconds float64) {
	PipelineRunDuration.WithLabelValues(namespace, config).Observe(durationSeconds)
}

// SetActiveRuns updates the active PipelineRuns gauge
func SetActiveRuns(namespace string, count int) {
	ActivePipelineRuns.WithLabelValues(namespace).Set(float64(count))
}

// SetPendingSteps updates the pending steps gauge
func SetPendingSteps(namespace string, count int) {
	PendingSteps.WithLabelValues(namespace).Set(float64(count))
}

// RecordStepFailed increments failed steps counter
func RecordStepFailed(namespace, step string) {
	FailedSteps.WithLabelValues(step, namespace).Inc()
}

// RecordStepResourceUsage records CPU and memory usage for a step
func RecordStepResourceUsage(namespace, pipelineRun, step string, cpuCores, memoryBytes float64) {
	StepResourceUsageCPU.WithLabelValues(step, namespace, pipelineRun).Set(cpuCores)
	StepResourceUsageMemory.WithLabelValues(step, namespace, pipelineRun).Set(memoryBytes)
}

// RecordJobCreationDuration records time taken to create a Job
func RecordJobCreationDuration(namespace string, durationSeconds float64) {
	JobCreationDuration.WithLabelValues(namespace).Observe(durationSeconds)
}

// RecordReconcileError increments reconciliation error counter
func RecordReconcileError(controller, namespace string) {
	ReconcileErrors.WithLabelValues(controller, namespace).Inc()
}
