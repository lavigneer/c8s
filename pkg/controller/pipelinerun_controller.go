/*
Copyright 2025 C8S Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
)

// PipelineRunReconciler reconciles a PipelineRun object
type PipelineRunReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=c8s.dev,resources=pipelineruns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=c8s.dev,resources=pipelineruns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=c8s.dev,resources=pipelineruns/finalizers,verbs=update
// +kubebuilder:rbac:groups=c8s.dev,resources=pipelineconfigs,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PipelineRunReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the PipelineRun instance
	pipelineRun := &c8sv1alpha1.PipelineRun{}
	if err := r.Get(ctx, req.NamespacedName, pipelineRun); err != nil {
		// PipelineRun not found, ignore since object must have been deleted
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling PipelineRun",
		"name", pipelineRun.Name,
		"namespace", pipelineRun.Namespace,
		"phase", pipelineRun.Status.Phase,
	)

	// TODO: Implement reconciliation logic in Phase 3
	// 1. Fetch referenced PipelineConfig
	// 2. If phase is Pending, initialize status and set to Running
	// 3. Parse pipeline steps and build dependency graph (DAG)
	// 4. Create Jobs for steps that are ready to run
	// 5. Watch Jobs and update step status based on Job status
	// 6. Update PipelineRun phase based on overall step completion
	// 7. Handle log collection and artifact upload to S3
	// 8. Set completion time and final phase (Succeeded/Failed/Cancelled)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PipelineRunReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&c8sv1alpha1.PipelineRun{}).
		Owns(&batchv1.Job{}).
		Complete(r)
}
