package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	v1alpha1 "github.com/org/c8s/pkg/apis/v1alpha1"
	"github.com/org/c8s/pkg/controller"
)

func TestBasicPipelineExecution(t *testing.T) {
	// Setup scheme
	s := runtime.NewScheme()
	require.NoError(t, scheme.AddToScheme(s))
	require.NoError(t, v1alpha1.AddToScheme(s))

	// Create fake client
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()

	// Create controller
	r := &controller.PipelineRunReconciler{
		Client: fakeClient,
		Scheme: s,
	}

	ctx := context.Background()

	// Create PipelineConfig with 2-step pipeline (test -> build)
	pipelineConfig := &v1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pipeline",
			Namespace: "default",
		},
		Spec: v1alpha1.PipelineConfigSpec{
			Repository: "https://github.com/example/repo",
			Branches:   []string{"main"},
			Steps: []v1alpha1.PipelineStep{
				{
					Name:     "test",
					Image:    "golang:1.21",
					Commands: []string{"go test ./..."},
				},
				{
					Name:      "build",
					Image:     "golang:1.21",
					Commands:  []string{"go build ./..."},
					DependsOn: []string{"test"}, // Depends on test step
				},
			},
		},
	}
	require.NoError(t, fakeClient.Create(ctx, pipelineConfig))

	// Create PipelineRun
	pipelineRun := &v1alpha1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-run-1",
			Namespace: "default",
		},
		Spec: v1alpha1.PipelineRunSpec{
			PipelineConfigRef: "test-pipeline",
			Commit:            "abc123",
			Branch:            "main",
			TriggeredBy:       "github-webhook",
		},
	}
	require.NoError(t, fakeClient.Create(ctx, pipelineRun))

	// First reconcile: should set status to Pending and create job for test step
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-run-1",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(ctx, req)
	require.NoError(t, err)
	assert.False(t, result.Requeue)

	// Check PipelineRun status was updated
	updatedRun := &v1alpha1.PipelineRun{}
	require.NoError(t, fakeClient.Get(ctx, req.NamespacedName, updatedRun))

	// Should be in Running phase (or Pending depending on implementation)
	assert.Contains(t, []string{"Pending", "Running"}, updatedRun.Status.Phase)
	assert.NotNil(t, updatedRun.Status.StartTime)

	// Check that test job was created
	testJob := &batchv1.Job{}
	testJobKey := types.NamespacedName{
		Name:      "test-run-1-test",
		Namespace: "default",
	}
	require.NoError(t, fakeClient.Get(ctx, testJobKey, testJob))
	assert.Equal(t, "test", testJob.Labels["c8s.dev/step-name"])
	assert.Equal(t, "test-run-1", testJob.Labels["c8s.dev/pipeline-run"])

	// Verify owner reference
	require.Len(t, testJob.OwnerReferences, 1)
	assert.Equal(t, "test-run-1", testJob.OwnerReferences[0].Name)
	assert.Equal(t, "PipelineRun", testJob.OwnerReferences[0].Kind)

	// Build job should NOT exist yet (dependency not satisfied)
	buildJob := &batchv1.Job{}
	buildJobKey := types.NamespacedName{
		Name:      "test-run-1-build",
		Namespace: "default",
	}
	err = fakeClient.Get(ctx, buildJobKey, buildJob)
	assert.Error(t, err) // Should not exist yet

	// Simulate test job succeeding
	testJob.Status.Succeeded = 1
	testJob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	require.NoError(t, fakeClient.Status().Update(ctx, testJob))

	// Second reconcile: should create build job now that test succeeded
	result, err = r.Reconcile(ctx, req)
	require.NoError(t, err)

	// Check that build job was created
	require.NoError(t, fakeClient.Get(ctx, buildJobKey, buildJob))
	assert.Equal(t, "build", buildJob.Labels["c8s.dev/step-name"])

	// Verify PipelineRun status has both steps
	require.NoError(t, fakeClient.Get(ctx, req.NamespacedName, updatedRun))
	assert.Len(t, updatedRun.Status.Steps, 2)

	// Find test step status
	var testStepStatus *v1alpha1.StepStatus
	for i := range updatedRun.Status.Steps {
		if updatedRun.Status.Steps[i].Name == "test" {
			testStepStatus = &updatedRun.Status.Steps[i]
			break
		}
	}
	require.NotNil(t, testStepStatus)
	assert.Equal(t, "Succeeded", testStepStatus.Phase)
	assert.Equal(t, "test-run-1-test", testStepStatus.JobName)

	// Simulate build job succeeding
	buildJob.Status.Succeeded = 1
	buildJob.Status.CompletionTime = &metav1.Time{Time: time.Now()}
	require.NoError(t, fakeClient.Status().Update(ctx, buildJob))

	// Third reconcile: should mark PipelineRun as Succeeded
	result, err = r.Reconcile(ctx, req)
	require.NoError(t, err)

	// Check final status
	require.NoError(t, fakeClient.Get(ctx, req.NamespacedName, updatedRun))
	assert.Equal(t, "Succeeded", updatedRun.Status.Phase)
	assert.NotNil(t, updatedRun.Status.CompletionTime)

	// Verify both steps succeeded
	for _, stepStatus := range updatedRun.Status.Steps {
		assert.Equal(t, "Succeeded", stepStatus.Phase)
		assert.NotNil(t, stepStatus.StartTime)
		assert.NotNil(t, stepStatus.CompletionTime)
	}
}

func TestPipelineExecutionWithFailure(t *testing.T) {
	// Setup scheme
	s := runtime.NewScheme()
	require.NoError(t, scheme.AddToScheme(s))
	require.NoError(t, v1alpha1.AddToScheme(s))

	// Create fake client
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()

	// Create controller
	r := &controller.PipelineRunReconciler{
		Client: fakeClient,
		Scheme: s,
	}

	ctx := context.Background()

	// Create PipelineConfig with single step
	pipelineConfig := &v1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "failing-pipeline",
			Namespace: "default",
		},
		Spec: v1alpha1.PipelineConfigSpec{
			Repository: "https://github.com/example/repo",
			Branches:   []string{"main"},
			Steps: []v1alpha1.PipelineStep{
				{
					Name:     "test",
					Image:    "golang:1.21",
					Commands: []string{"go test ./..."},
				},
			},
		},
	}
	require.NoError(t, fakeClient.Create(ctx, pipelineConfig))

	// Create PipelineRun
	pipelineRun := &v1alpha1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "failing-run-1",
			Namespace: "default",
		},
		Spec: v1alpha1.PipelineRunSpec{
			PipelineConfigRef: "failing-pipeline",
			Commit:            "def456",
			Branch:            "main",
			TriggeredBy:       "manual",
		},
	}
	require.NoError(t, fakeClient.Create(ctx, pipelineRun))

	// First reconcile
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "failing-run-1",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(ctx, req)
	require.NoError(t, err)
	assert.False(t, result.Requeue)

	// Get the created job
	testJob := &batchv1.Job{}
	testJobKey := types.NamespacedName{
		Name:      "failing-run-1-test",
		Namespace: "default",
	}
	require.NoError(t, fakeClient.Get(ctx, testJobKey, testJob))

	// Simulate job failing
	testJob.Status.Failed = 1
	testJob.Status.Conditions = []batchv1.JobCondition{
		{
			Type:    batchv1.JobFailed,
			Status:  corev1.ConditionTrue,
			Reason:  "BackoffLimitExceeded",
			Message: "Job has reached the specified backoff limit",
		},
	}
	require.NoError(t, fakeClient.Status().Update(ctx, testJob))

	// Second reconcile: should mark PipelineRun as Failed
	result, err = r.Reconcile(ctx, req)
	require.NoError(t, err)

	// Check final status
	updatedRun := &v1alpha1.PipelineRun{}
	require.NoError(t, fakeClient.Get(ctx, req.NamespacedName, updatedRun))
	assert.Equal(t, "Failed", updatedRun.Status.Phase)
	assert.NotNil(t, updatedRun.Status.CompletionTime)

	// Verify step marked as failed
	require.Len(t, updatedRun.Status.Steps, 1)
	assert.Equal(t, "Failed", updatedRun.Status.Steps[0].Phase)
}

func TestPipelineExecutionParallelSteps(t *testing.T) {
	// Setup scheme
	s := runtime.NewScheme()
	require.NoError(t, scheme.AddToScheme(s))
	require.NoError(t, v1alpha1.AddToScheme(s))

	// Create fake client
	fakeClient := fake.NewClientBuilder().WithScheme(s).Build()

	// Create controller
	r := &controller.PipelineRunReconciler{
		Client: fakeClient,
		Scheme: s,
	}

	ctx := context.Background()

	// Create PipelineConfig with 2 parallel steps (no dependencies)
	pipelineConfig := &v1alpha1.PipelineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parallel-pipeline",
			Namespace: "default",
		},
		Spec: v1alpha1.PipelineConfigSpec{
			Repository: "https://github.com/example/repo",
			Branches:   []string{"main"},
			Steps: []v1alpha1.PipelineStep{
				{
					Name:     "lint",
					Image:    "golangci/golangci-lint:latest",
					Commands: []string{"golangci-lint run"},
				},
				{
					Name:     "test",
					Image:    "golang:1.21",
					Commands: []string{"go test ./..."},
				},
			},
		},
	}
	require.NoError(t, fakeClient.Create(ctx, pipelineConfig))

	// Create PipelineRun
	pipelineRun := &v1alpha1.PipelineRun{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "parallel-run-1",
			Namespace: "default",
		},
		Spec: v1alpha1.PipelineRunSpec{
			PipelineConfigRef: "parallel-pipeline",
			Commit:            "ghi789",
			Branch:            "main",
			TriggeredBy:       "manual",
		},
	}
	require.NoError(t, fakeClient.Create(ctx, pipelineRun))

	// First reconcile: should create BOTH jobs (parallel execution)
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "parallel-run-1",
			Namespace: "default",
		},
	}

	result, err := r.Reconcile(ctx, req)
	require.NoError(t, err)
	assert.False(t, result.Requeue)

	// Check that BOTH jobs were created
	lintJob := &batchv1.Job{}
	lintJobKey := types.NamespacedName{
		Name:      "parallel-run-1-lint",
		Namespace: "default",
	}
	require.NoError(t, fakeClient.Get(ctx, lintJobKey, lintJob))

	testJob := &batchv1.Job{}
	testJobKey := types.NamespacedName{
		Name:      "parallel-run-1-test",
		Namespace: "default",
	}
	require.NoError(t, fakeClient.Get(ctx, testJobKey, testJob))

	// Both jobs should exist simultaneously (parallel execution)
	assert.Equal(t, "lint", lintJob.Labels["c8s.dev/step-name"])
	assert.Equal(t, "test", testJob.Labels["c8s.dev/step-name"])
}
