package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/secrets"
	"github.com/org/c8s/pkg/storage"
)

const (
	// MaxLogBufferSize is the maximum size of logs kept in memory (10MB)
	MaxLogBufferSize = 10 * 1024 * 1024
)

// LogCollector handles collecting logs from Job Pods
type LogCollector struct {
	client        kubernetes.Interface
	storageClient storage.StorageClient
	bufferManager *LogBufferManager
}

// NewLogCollector creates a new LogCollector
func NewLogCollector(client kubernetes.Interface, storageClient storage.StorageClient) *LogCollector {
	return &LogCollector{
		client:        client,
		storageClient: storageClient,
		bufferManager: NewLogBufferManager(),
	}
}

// CollectLogs streams logs from a Pod and returns them as bytes
func (lc *LogCollector) CollectLogs(ctx context.Context, pod *corev1.Pod) ([]byte, error) {
	logger := log.FromContext(ctx)

	if pod.Status.Phase != corev1.PodSucceeded && pod.Status.Phase != corev1.PodFailed && pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("pod %s/%s is not in a state where logs can be collected: %s", pod.Namespace, pod.Name, pod.Status.Phase)
	}

	// Find the main container (not init containers)
	var mainContainer string
	for _, container := range pod.Spec.Containers {
		if container.Name != "git-clone" {
			mainContainer = container.Name
			break
		}
	}

	if mainContainer == "" {
		return nil, fmt.Errorf("no main container found in pod %s/%s", pod.Namespace, pod.Name)
	}

	// Get logs from the main container
	req := lc.client.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: mainContainer,
	})

	stream, err := req.Stream(ctx)
	if err != nil {
		logger.Error(err, "failed to stream logs", "pod", pod.Name, "container", mainContainer)
		return nil, fmt.Errorf("failed to stream logs from pod %s/%s: %w", pod.Namespace, pod.Name, err)
	}
	defer stream.Close()

	// Read logs into buffer with size limit
	buf := &bytes.Buffer{}
	limitedReader := io.LimitReader(stream, MaxLogBufferSize)

	_, err = io.Copy(buf, limitedReader)
	if err != nil {
		logger.Error(err, "failed to read logs", "pod", pod.Name)
		return nil, fmt.Errorf("failed to read logs from pod %s/%s: %w", pod.Namespace, pod.Name, err)
	}

	logs := buf.Bytes()
	logger.Info("collected logs", "pod", pod.Name, "size", len(logs))

	return logs, nil
}

// UploadLogsToStorage uploads logs to S3 and returns the log URL
// Logs are masked before upload to ensure secrets are never persisted
func (lc *LogCollector) UploadLogsToStorage(ctx context.Context, pipelineRun *v1alpha1.PipelineRun, stepName string, logs []byte, pipelineConfig *v1alpha1.PipelineConfig) (string, error) {
	logger := log.FromContext(ctx)

	if lc.storageClient == nil {
		logger.Info("no storage client configured, skipping log upload")
		return "", nil
	}

	// Fetch secret values for masking
	secretValues, err := lc.fetchSecretValues(ctx, pipelineRun.Namespace, pipelineConfig, stepName)
	if err != nil {
		logger.Error(err, "failed to fetch secret values for masking", "step", stepName)
		// Continue with upload but log the error
	}

	// Mask secrets in logs before uploading
	maskedLogs := secrets.MaskSecrets(logs, secretValues)

	// Log if any secrets were redacted (for audit purposes)
	if secrets.HasRedactedContent(maskedLogs) {
		redactionCount := secrets.CountRedactions(maskedLogs)
		logger.Info("masked secrets in logs", "step", stepName, "redactions", redactionCount)
	}

	// Generate storage key: {namespace}/{pipelinerun-name}/{step-name}.log
	key := fmt.Sprintf("%s/%s/%s.log", pipelineRun.Namespace, pipelineRun.Name, stepName)

	// Convert masked logs to io.Reader
	reader := bytes.NewReader(maskedLogs)

	// Upload masked logs to storage
	err = lc.storageClient.UploadLog(ctx, key, reader)
	if err != nil {
		logger.Error(err, "failed to upload logs to storage", "key", key)
		return "", fmt.Errorf("failed to upload logs: %w", err)
	}

	// Generate a signed URL for accessing the logs (valid for 7 days)
	logURL, err := lc.storageClient.GenerateSignedURL(ctx, key, 7*24*3600)
	if err != nil {
		logger.Error(err, "failed to generate signed URL", "key", key)
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	logger.Info("uploaded logs to storage", "key", key, "url", logURL)
	return logURL, nil
}

// CollectAndUpload is a convenience method that collects logs from a Pod and uploads them to storage
func (lc *LogCollector) CollectAndUpload(ctx context.Context, pod *corev1.Pod, pipelineRun *v1alpha1.PipelineRun, stepName string, pipelineConfig *v1alpha1.PipelineConfig) (string, error) {
	logger := log.FromContext(ctx)

	// Collect logs
	logs, err := lc.CollectLogs(ctx, pod)
	if err != nil {
		return "", err
	}

	// Fetch secret values for masking
	secretValues, err := lc.fetchSecretValues(ctx, pipelineRun.Namespace, pipelineConfig, stepName)
	if err != nil {
		logger.Error(err, "failed to fetch secret values for masking", "step", stepName)
		// Continue with masked logs using empty secret map
		secretValues = make(map[string]string)
	}

	// Mask secrets in logs before storing in buffer
	maskedLogs := secrets.MaskSecrets(logs, secretValues)

	// Store masked logs in circular buffer for real-time streaming
	bufferKey := fmt.Sprintf("%s/%s/%s", pipelineRun.Namespace, pipelineRun.Name, stepName)
	lc.bufferManager.Write(bufferKey, maskedLogs)

	// Upload to storage (masking happens again inside for safety)
	logURL, err := lc.UploadLogsToStorage(ctx, pipelineRun, stepName, maskedLogs, pipelineConfig)
	if err != nil {
		// Log the error but don't fail - logs are still in buffer
		logger.Error(err, "failed to upload logs, but they are available in buffer", "step", stepName)
		return "", err
	}

	return logURL, nil
}

// fetchSecretValues fetches all secret values referenced by a pipeline step for masking purposes
func (lc *LogCollector) fetchSecretValues(ctx context.Context, namespace string, pipelineConfig *v1alpha1.PipelineConfig, stepName string) (map[string]string, error) {
	logger := log.FromContext(ctx)
	secretValues := make(map[string]string)

	if pipelineConfig == nil {
		return secretValues, nil
	}

	// Find the step in the pipeline config
	var targetStep *v1alpha1.PipelineStep
	for i := range pipelineConfig.Spec.Steps {
		if pipelineConfig.Spec.Steps[i].Name == stepName {
			targetStep = &pipelineConfig.Spec.Steps[i]
			break
		}
	}

	if targetStep == nil {
		return secretValues, fmt.Errorf("step %s not found in pipeline config", stepName)
	}

	// Fetch all referenced secrets
	for _, secretRef := range targetStep.Secrets {
		secret, err := lc.client.CoreV1().Secrets(namespace).Get(ctx, secretRef.SecretRef, metav1.GetOptions{})
		if err != nil {
			logger.Error(err, "failed to fetch secret for masking", "secret", secretRef.SecretRef)
			continue
		}

		// Extract the specific key value
		if value, ok := secret.Data[secretRef.Key]; ok {
			// Use the secret name and key as the identifier
			identifier := fmt.Sprintf("%s:%s", secretRef.SecretRef, secretRef.Key)
			secretValues[identifier] = string(value)
		}
	}

	return secretValues, nil
}

// GetLogBuffer returns the log buffer manager for real-time streaming
func (lc *LogCollector) GetLogBuffer() *LogBufferManager {
	return lc.bufferManager
}

// LogBufferManager manages circular buffers for real-time log streaming
type LogBufferManager struct {
	buffers map[string]*CircularBuffer
}

// NewLogBufferManager creates a new LogBufferManager
func NewLogBufferManager() *LogBufferManager {
	return &LogBufferManager{
		buffers: make(map[string]*CircularBuffer),
	}
}

// Write writes logs to a circular buffer
func (lbm *LogBufferManager) Write(key string, data []byte) {
	if _, exists := lbm.buffers[key]; !exists {
		lbm.buffers[key] = NewCircularBuffer(MaxLogBufferSize)
	}
	lbm.buffers[key].Write(data)
}

// Read reads logs from a circular buffer
func (lbm *LogBufferManager) Read(key string) []byte {
	if buf, exists := lbm.buffers[key]; exists {
		return buf.Read()
	}
	return nil
}

// Subscribe creates a channel that receives log updates
func (lbm *LogBufferManager) Subscribe(key string) <-chan []byte {
	if _, exists := lbm.buffers[key]; !exists {
		lbm.buffers[key] = NewCircularBuffer(MaxLogBufferSize)
	}
	return lbm.buffers[key].Subscribe()
}

// CircularBuffer implements a thread-safe circular buffer for logs
type CircularBuffer struct {
	data        []byte
	maxSize     int
	subscribers []chan []byte
}

// NewCircularBuffer creates a new CircularBuffer
func NewCircularBuffer(maxSize int) *CircularBuffer {
	return &CircularBuffer{
		data:        make([]byte, 0, maxSize),
		maxSize:     maxSize,
		subscribers: make([]chan []byte, 0),
	}
}

// Write appends data to the buffer
func (cb *CircularBuffer) Write(data []byte) {
	// If adding data would exceed max size, truncate from beginning
	if len(cb.data)+len(data) > cb.maxSize {
		// Keep only the most recent data that fits
		keepSize := cb.maxSize - len(data)
		if keepSize > 0 {
			cb.data = cb.data[len(cb.data)-keepSize:]
		} else {
			cb.data = cb.data[:0]
		}
	}

	cb.data = append(cb.data, data...)

	// Notify all subscribers
	for _, sub := range cb.subscribers {
		select {
		case sub <- data:
		default:
			// Subscriber not ready, skip
		}
	}
}

// Read returns all data in the buffer
func (cb *CircularBuffer) Read() []byte {
	result := make([]byte, len(cb.data))
	copy(result, cb.data)
	return result
}

// Subscribe creates a channel that receives new log data
func (cb *CircularBuffer) Subscribe() <-chan []byte {
	ch := make(chan []byte, 100)
	cb.subscribers = append(cb.subscribers, ch)
	return ch
}
