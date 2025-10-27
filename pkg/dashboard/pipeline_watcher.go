package dashboard

import (
	"context"
	"fmt"
	"sync"

	"github.com/org/c8s/pkg/apis/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PipelineWatcher watches Kubernetes PipelineRun resources and broadcasts updates
type PipelineWatcher struct {
	k8sClient    *K8sClient
	watchers     map[string]watch.Interface
	watcherMutex sync.RWMutex
	stopCh       chan struct{}
	done         chan struct{}
}

// NewPipelineWatcher creates a new pipeline watcher
func NewPipelineWatcher(k8sClient *K8sClient) *PipelineWatcher {
	return &PipelineWatcher{
		k8sClient: k8sClient,
		watchers:  make(map[string]watch.Interface),
		stopCh:    make(chan struct{}),
		done:      make(chan struct{}),
	}
}

// Start begins watching PipelineRuns in a namespace
func (w *PipelineWatcher) Start(ctx context.Context, namespace string) error {
	// Create a watcher for PipelineRun resources
	listOpts := []client.ListOption{
		client.InNamespace(namespace),
	}

	// Watch for changes
	watchInterface, err := w.k8sClient.Client.Watch(ctx, &v1alpha1.PipelineRunList{}, listOpts...)
	if err != nil {
		return fmt.Errorf("failed to watch PipelineRuns: %w", err)
	}

	w.watcherMutex.Lock()
	w.watchers[namespace] = watchInterface
	w.watcherMutex.Unlock()

	// Process watch events in a goroutine
	go w.processWatchEvents(ctx, namespace, watchInterface)

	return nil
}

// processWatchEvents processes Kubernetes watch events
func (w *PipelineWatcher) processWatchEvents(ctx context.Context, namespace string, watchInterface watch.Interface) {
	defer func() {
		w.watcherMutex.Lock()
		delete(w.watchers, namespace)
		w.watcherMutex.Unlock()
	}()

	eventChan := watchInterface.ResultChan()

	for {
		select {
		case <-ctx.Done():
			watchInterface.Stop()
			return

		case <-w.stopCh:
			watchInterface.Stop()
			return

		case event, ok := <-eventChan:
			if !ok {
				// Channel closed
				return
			}

			w.handleWatchEvent(event)
		}
	}
}

// handleWatchEvent processes a Kubernetes watch event
func (w *PipelineWatcher) handleWatchEvent(event watch.Event) {
	run, ok := event.Object.(*v1alpha1.PipelineRun)
	if !ok {
		return
	}

	// Convert to DTO
	dto := MapPipelineRunToDTO(run)
	if dto == nil {
		return
	}

	// Get project ID from labels
	projectID := run.Labels["project"]
	if projectID == "" {
		projectID = "default-project"
	}

	// Broadcast the update based on event type
	switch event.Type {
	case watch.Added:
		// New pipeline run created
		// This will be handled by the SSE broadcaster
		BroadcastPipelineUpdate(projectID, dto)

	case watch.Modified:
		// Pipeline run status changed
		BroadcastPipelineUpdate(projectID, dto)

	case watch.Deleted:
		// Pipeline run deleted (less common, but handle it)
		// Create a deleted event
		BroadcastPipelineDelete(projectID, dto)
	}
}

// BroadcastPipelineDelete sends a deletion event
func BroadcastPipelineDelete(projectID string, run *PipelineRunDTO) {
	// This is a placeholder - would need similar broadcaster pattern
	// For now, we'll just broadcast a deletion event
}

// Stop stops the watcher
func (w *PipelineWatcher) Stop() {
	w.watcherMutex.Lock()
	defer w.watcherMutex.Unlock()

	close(w.stopCh)

	for _, watchInterface := range w.watchers {
		watchInterface.Stop()
	}
	w.watchers = make(map[string]watch.Interface)

	close(w.done)
}

// WaitForStop waits for the watcher to stop
func (w *PipelineWatcher) WaitForStop() {
	<-w.done
}

// GetWatchingNamespaces returns list of namespaces being watched
func (w *PipelineWatcher) GetWatchingNamespaces() []string {
	w.watcherMutex.RLock()
	defer w.watcherMutex.RUnlock()

	namespaces := make([]string, 0, len(w.watchers))
	for ns := range w.watchers {
		namespaces = append(namespaces, ns)
	}

	return namespaces
}

// HealthCheck returns true if watcher is healthy
func (w *PipelineWatcher) HealthCheck() bool {
	w.watcherMutex.RLock()
	defer w.watcherMutex.RUnlock()

	// Consider healthy if we're watching at least one namespace
	return len(w.watchers) > 0
}
