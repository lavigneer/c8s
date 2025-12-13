package logstorage

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// KubernetesLogStorage fetches logs from actual Kubernetes Job Pods
type KubernetesLogStorage struct {
	k8sClient client.Client
	namespace string
	runID     string
}

// NewKubernetesLogStorage creates a new Kubernetes log storage for a specific pipeline run
func NewKubernetesLogStorage(k8sClient client.Client, namespace, runID string) *KubernetesLogStorage {
	return &KubernetesLogStorage{
		k8sClient: k8sClient,
		namespace: namespace,
		runID:     runID,
	}
}

// GetStepLogs returns logs for a step by reading from the Kubernetes Pod
func (k *KubernetesLogStorage) GetStepLogs(ctx context.Context, runID, stepID string) (io.ReadCloser, error) {
	// Fetch the PipelineRun to find the Job
	pipelineRun := &c8sv1alpha1.PipelineRun{}
	if err := k.k8sClient.Get(ctx, client.ObjectKey{Namespace: k.namespace, Name: runID}, pipelineRun); err != nil {
		return nil, fmt.Errorf("failed to fetch PipelineRun: %w", err)
	}

	// Find the step status
	var stepStatus *c8sv1alpha1.StepStatus
	for i := range pipelineRun.Status.Steps {
		if pipelineRun.Status.Steps[i].Name == stepID {
			stepStatus = &pipelineRun.Status.Steps[i]
			break
		}
	}
	if stepStatus == nil {
		return nil, fmt.Errorf("step %s not found in pipeline run", stepID)
	}

	// Get the Job name from step status
	if stepStatus.JobName == "" {
		return nil, fmt.Errorf("job not yet created for step %s", stepID)
	}

	// Find the Pod created by this Job
	podList := &corev1.PodList{}
	if err := k.k8sClient.List(ctx, podList,
		client.InNamespace(k.namespace),
		client.MatchingLabels{
			"job-name": stepStatus.JobName,
		},
	); err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no pods found for job %s", stepStatus.JobName)
	}

	pod := &podList.Items[0]

	// Get logs from the pod
	logStream, err := k.getPodLogs(ctx, pod)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}

	return logStream, nil
}

// StreamStepLogs streams logs line-by-line from a Kubernetes Pod
func (k *KubernetesLogStorage) StreamStepLogs(ctx context.Context, runID, stepID string, linesChan chan<- string) error {
	logsReader, err := k.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return err
	}
	defer logsReader.Close()

	// Stream logs line by line
	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case linesChan <- scanner.Text():
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading logs: %w", err)
	}

	return nil
}

// GetLogSnapshot returns the last N lines of logs
func (k *KubernetesLogStorage) GetLogSnapshot(ctx context.Context, runID, stepID string, lines int) ([]string, error) {
	logsReader, err := k.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return nil, err
	}
	defer logsReader.Close()

	var result []string
	scanner := bufio.NewScanner(logsReader)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}

	// Return last N lines
	if len(result) > lines {
		result = result[len(result)-lines:]
	}

	return result, nil
}

// GetLogSize returns the size of logs in bytes
func (k *KubernetesLogStorage) GetLogSize(ctx context.Context, runID, stepID string) (int64, error) {
	logsReader, err := k.GetStepLogs(ctx, runID, stepID)
	if err != nil {
		return 0, err
	}
	defer logsReader.Close()

	// Read all logs and count bytes
	data, err := io.ReadAll(logsReader)
	if err != nil {
		return 0, fmt.Errorf("failed to read logs: %w", err)
	}

	return int64(len(data)), nil
}

// getPodLogs fetches logs from a Kubernetes Pod
func (k *KubernetesLogStorage) getPodLogs(ctx context.Context, pod *corev1.Pod) (io.ReadCloser, error) {
	// Wait for pod to start (it might not have logs immediately)
	// For now, try to get logs from the first container
	if len(pod.Spec.Containers) == 0 {
		return nil, fmt.Errorf("pod has no containers")
	}

	// Get logs from the first non-init container
	containerName := ""
	for _, container := range pod.Spec.Containers {
		if container.Name != "git-clone" { // Skip init container
			containerName = container.Name
			break
		}
	}
	if containerName == "" && len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	// Use the Kubernetes API to get logs
	// This requires the REST client to be available
	// For now, return a simple implementation that reads from the pod's status
	return k.fetchPodLogsFromAPI(ctx, pod, containerName)
}

// fetchPodLogsFromAPI fetches logs using the Kubernetes API
func (k *KubernetesLogStorage) fetchPodLogsFromAPI(ctx context.Context, pod *corev1.Pod, containerName string) (io.ReadCloser, error) {
	// Get the pod logs using the RESTClient
	req := k.k8sClient.RESTClient().
		Get().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("log").
		Param("container", containerName)

	stream, err := req.Stream(ctx)
	if err != nil {
		// If we can't get logs yet, return a placeholder
		return io.NopCloser(strings.NewReader(fmt.Sprintf("Pod logs not available yet (pod phase: %s)\n", pod.Status.Phase))), nil
	}

	return stream, nil
}
