package dashboard

import (
	"context"
	"sync"

	"k8s.io/apimachinery/pkg/watch"
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
// TODO: Implement using a polling mechanism since controller-runtime client.Watch() is not available
func (w *PipelineWatcher) Start(ctx context.Context, namespace string) error {
	// For now, this is a placeholder
	// The watcher functionality will be implemented using polling in a future iteration
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
// TODO: Implement event broadcasting when watch is available
func (w *PipelineWatcher) handleWatchEvent(event watch.Event) {
	// Placeholder for event handling
	// Will be implemented in a future iteration
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
