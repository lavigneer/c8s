package controller

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	c8sv1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/scheduler"
)

// CreateMatrixPipelineRuns creates multiple PipelineRuns for matrix strategy execution
// Returns the created PipelineRuns
func CreateMatrixPipelineRuns(
	ctx context.Context,
	c client.Client,
	pipelineConfig *c8sv1alpha1.PipelineConfig,
	baseRun *c8sv1alpha1.PipelineRun,
) ([]*c8sv1alpha1.PipelineRun, error) {
	logger := log.FromContext(ctx)

	// Expand matrix to get all combinations
	combinations, err := scheduler.ExpandMatrix(pipelineConfig.Spec.Matrix)
	if err != nil {
		return nil, fmt.Errorf("failed to expand matrix: %w", err)
	}

	logger.Info("Creating matrix PipelineRuns",
		"combinations", len(combinations),
		"config", pipelineConfig.Name,
	)

	// Generate a parent ID for grouping matrix runs
	parentID := fmt.Sprintf("%s-%s", baseRun.Name, baseRun.UID[:8])

	var createdRuns []*c8sv1alpha1.PipelineRun

	for i, matrixVars := range combinations {
		// Create a PipelineRun for this matrix combination
		matrixRun := &c8sv1alpha1.PipelineRun{
			ObjectMeta: metav1.ObjectMeta{
				Name:      scheduler.GenerateMatrixRunName(baseRun.Name, i, matrixVars),
				Namespace: baseRun.Namespace,
				Labels: mergeLabels(
					baseRun.Labels,
					map[string]string{
						"c8s.dev/matrix-parent": parentID,
						"c8s.dev/matrix-index":  fmt.Sprintf("%d", i),
					},
					scheduler.MatrixToLabels(matrixVars),
				),
				Annotations: baseRun.Annotations,
			},
			Spec: c8sv1alpha1.PipelineRunSpec{
				PipelineConfigRef: baseRun.Spec.PipelineConfigRef,
				Commit:            baseRun.Spec.Commit,
				Branch:            baseRun.Spec.Branch,
				TriggeredBy:       baseRun.Spec.TriggeredBy,
				TriggeredAt:       baseRun.Spec.TriggeredAt,
				MatrixIndex:       matrixVars,
				CommitMessage:     baseRun.Spec.CommitMessage,
				Author:            baseRun.Spec.Author,
			},
		}

		// Create the PipelineRun
		if err := c.Create(ctx, matrixRun); err != nil {
			logger.Error(err, "Failed to create matrix PipelineRun",
				"index", i,
				"name", matrixRun.Name,
			)
			// Continue creating other runs even if one fails
			continue
		}

		logger.Info("Created matrix PipelineRun",
			"index", i,
			"name", matrixRun.Name,
			"matrixVars", matrixVars,
		)

		createdRuns = append(createdRuns, matrixRun)
	}

	if len(createdRuns) == 0 {
		return nil, fmt.Errorf("failed to create any matrix PipelineRuns")
	}

	return createdRuns, nil
}

// ApplyMatrixToConfig creates a new PipelineConfig with matrix variables substituted
// This is used during reconciliation to get the actual step definitions for a matrix run
func ApplyMatrixToConfig(config *c8sv1alpha1.PipelineConfig, matrixVars map[string]string) *c8sv1alpha1.PipelineConfig {
	newConfig := config.DeepCopy()

	// Apply matrix substitution to each step
	for i := range newConfig.Spec.Steps {
		newConfig.Spec.Steps[i] = scheduler.ApplyMatrixToStep(newConfig.Spec.Steps[i], matrixVars)
	}

	return newConfig
}

// GetMatrixParentRun fetches the parent PipelineRun for a matrix execution
func GetMatrixParentRun(ctx context.Context, c client.Client, matrixRun *c8sv1alpha1.PipelineRun) (*c8sv1alpha1.PipelineRun, error) {
	parentID, ok := matrixRun.Labels["c8s.dev/matrix-parent"]
	if !ok {
		return nil, fmt.Errorf("matrix-parent label not found")
	}

	// List all PipelineRuns with matching UID prefix
	// This is a heuristic since we can't directly look up by UID prefix
	// In a real implementation, we might store the parent name explicitly
	var runs c8sv1alpha1.PipelineRunList
	if err := c.List(ctx, &runs, client.InNamespace(matrixRun.Namespace)); err != nil {
		return nil, err
	}

	for _, run := range runs.Items {
		expectedID := fmt.Sprintf("%s-%s", run.Name, run.UID[:8])
		if expectedID == parentID {
			return &run, nil
		}
	}

	return nil, fmt.Errorf("parent PipelineRun not found for parent ID: %s", parentID)
}

// ListMatrixRuns lists all PipelineRuns that are part of a matrix execution
func ListMatrixRuns(ctx context.Context, c client.Client, namespace, parentID string) ([]c8sv1alpha1.PipelineRun, error) {
	var runs c8sv1alpha1.PipelineRunList
	if err := c.List(ctx, &runs,
		client.InNamespace(namespace),
		client.MatchingLabels{"c8s.dev/matrix-parent": parentID},
	); err != nil {
		return nil, err
	}

	return runs.Items, nil
}

// IsMatrixRun checks if a PipelineRun is part of a matrix execution
func IsMatrixRun(run *c8sv1alpha1.PipelineRun) bool {
	_, hasParent := run.Labels["c8s.dev/matrix-parent"]
	return hasParent || len(run.Spec.MatrixIndex) > 0
}

// ShouldCreateMatrixRuns determines if matrix runs should be created for a PipelineRun
// Matrix runs are created only once when the initial PipelineRun is created
func ShouldCreateMatrixRuns(run *c8sv1alpha1.PipelineRun, config *c8sv1alpha1.PipelineConfig) bool {
	// Don't create matrix runs if this is already a matrix run
	if IsMatrixRun(run) {
		return false
	}

	// Don't create if no matrix strategy defined
	if config.Spec.Matrix == nil || len(config.Spec.Matrix.Dimensions) == 0 {
		return false
	}

	// Only create matrix runs if the run is in Pending phase
	// This ensures we only create them once
	if run.Status.Phase != "" && run.Status.Phase != c8sv1alpha1.PipelineRunPhasePending {
		return false
	}

	return true
}

// MatrixReconciler handles matrix strategy expansion
type MatrixReconciler struct {
	client.Client
}

// Reconcile handles matrix PipelineRun creation
func (r *MatrixReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the PipelineRun
	var run c8sv1alpha1.PipelineRun
	if err := r.Get(ctx, req.NamespacedName, &run); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Skip if already a matrix run or completed
	if IsMatrixRun(&run) {
		return ctrl.Result{}, nil
	}

	// Fetch PipelineConfig
	var config c8sv1alpha1.PipelineConfig
	configKey := types.NamespacedName{
		Name:      run.Spec.PipelineConfigRef,
		Namespace: run.Namespace,
	}
	if err := r.Get(ctx, configKey, &config); err != nil {
		return ctrl.Result{}, err
	}

	// Check if we should create matrix runs
	if !ShouldCreateMatrixRuns(&run, &config) {
		return ctrl.Result{}, nil
	}

	logger.Info("Creating matrix PipelineRuns for matrix strategy",
		"pipelineRun", run.Name,
		"config", config.Name,
	)

	// Create matrix runs
	matrixRuns, err := CreateMatrixPipelineRuns(ctx, r.Client, &config, &run)
	if err != nil {
		logger.Error(err, "Failed to create matrix PipelineRuns")
		return ctrl.Result{}, err
	}

	// Update the original run to mark it as a matrix parent
	run.Status.Phase = c8sv1alpha1.PipelineRunPhaseSucceeded
	run.Status.CompletionTime = &metav1.Time{Time: metav1.Now().Time}
	if err := r.Status().Update(ctx, &run); err != nil {
		logger.Error(err, "Failed to update matrix parent status")
		return ctrl.Result{}, err
	}

	logger.Info("Successfully created matrix PipelineRuns",
		"count", len(matrixRuns),
		"parent", run.Name,
	)

	return ctrl.Result{}, nil
}

// mergeLabels merges multiple label maps
func mergeLabels(labelMaps ...map[string]string) map[string]string {
	result := make(map[string]string)
	for _, labels := range labelMaps {
		for k, v := range labels {
			result[k] = v
		}
	}
	return result
}
