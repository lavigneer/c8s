# Data Model: Kubernetes-Native CI System

**Feature**: 001-build-a-continuous
**Date**: 2025-10-12
**Status**: Complete

## Overview

This document defines all entities (Custom Resource Definitions), their schemas, relationships, validation rules, and state transitions for the Kubernetes-native CI system. All entities are implemented as Kubernetes CRDs backed by etcd, following K8s API conventions.

---

## Entity Summary

| Entity | Kind | Purpose | Namespaced |
|--------|------|---------|------------|
| PipelineConfig | CRD | Defines pipeline structure and steps | Yes |
| PipelineRun | CRD | Represents single execution instance | Yes |
| StepExecution | Sub-resource | Individual step status within PipelineRun | N/A (embedded) |
| RepositoryConnection | CRD | Links repository to webhook configuration | Yes |
| ResourceQuota | Native K8s | Limits resources per team/namespace | Yes |
| Secret | Native K8s | Stores credentials and sensitive data | Yes |

---

## Entity 1: PipelineConfig

**Purpose**: Defines the structure of a CI pipeline including steps, dependencies, resource requirements, and execution conditions. Stored in cluster and referenced by repository webhook configuration.

**CRD Definition**:
```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: pipelineconfigs.c8s.dev
spec:
  group: c8s.dev
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          required: [spec]
          properties:
            spec:
              type: object
              required: [repository, steps]
              properties:
                repository:
                  type: string
                  description: Git repository URL (https or ssh)
                  pattern: '^(https?|git|ssh)://.*'
                branches:
                  type: array
                  description: Branch filters (glob patterns, default ["*"])
                  items:
                    type: string
                  default: ["*"]
                steps:
                  type: array
                  description: Pipeline steps in execution order
                  minItems: 1
                  items:
                    type: object
                    required: [name, image, commands]
                    properties:
                      name:
                        type: string
                        description: Step identifier (must be unique)
                        pattern: '^[a-z0-9]([-a-z0-9]*[a-z0-9])?$'
                      image:
                        type: string
                        description: Container image for step execution
                      commands:
                        type: array
                        description: Shell commands to execute
                        items:
                          type: string
                      dependsOn:
                        type: array
                        description: Step names that must complete before this step
                        items:
                          type: string
                      resources:
                        type: object
                        description: CPU/memory requests and limits
                        properties:
                          cpu:
                            type: string
                            pattern: '^[0-9]+m?$'
                            default: "500m"
                          memory:
                            type: string
                            pattern: '^[0-9]+(Mi|Gi)$'
                            default: "1Gi"
                      timeout:
                        type: string
                        description: Step timeout (e.g., "30m", "2h")
                        pattern: '^[0-9]+(s|m|h)$'
                        default: "30m"
                      artifacts:
                        type: array
                        description: File patterns to upload to artifact storage
                        items:
                          type: string
                      secrets:
                        type: array
                        description: Secret references to inject as env vars
                        items:
                          type: object
                          required: [secretRef, key]
                          properties:
                            secretRef:
                              type: string
                              description: Kubernetes Secret name
                            key:
                              type: string
                              description: Key within Secret
                            envVar:
                              type: string
                              description: Environment variable name (defaults to key)
                      conditional:
                        type: object
                        description: Conditions for step execution
                        properties:
                          branch:
                            type: string
                            description: Execute only on matching branch pattern
                          onSuccess:
                            type: boolean
                            description: Execute only if previous steps succeeded
                            default: true
                timeout:
                  type: string
                  description: Pipeline-level timeout
                  pattern: '^[0-9]+(s|m|h)$'
                  default: "1h"
                matrix:
                  type: object
                  description: Matrix strategy for parallel execution
                  properties:
                    dimensions:
                      type: object
                      additionalProperties:
                        type: array
                        items:
                          type: string
                    exclude:
                      type: array
                      items:
                        type: object
            status:
              type: object
              description: Runtime status (managed by controller)
              properties:
                lastRun:
                  type: string
                  format: date-time
                totalRuns:
                  type: integer
                successRate:
                  type: number
  scope: Namespaced
  names:
    plural: pipelineconfigs
    singular: pipelineconfig
    kind: PipelineConfig
    shortNames:
      - pc
```

**Validation Rules**:
- `spec.repository` MUST be valid Git URL
- `spec.steps[].name` MUST be unique within pipeline
- `spec.steps[].dependsOn` MUST reference existing step names (no cycles)
- `spec.steps[].resources.cpu` MUST not exceed namespace ResourceQuota
- `spec.steps[].resources.memory` MUST not exceed namespace ResourceQuota
- `spec.matrix.dimensions` MUST contain at least one dimension if specified
- Total timeout MUST be >= sum of all step timeouts

**Relationships**:
- Referenced by: `PipelineRun.spec.pipelineConfigRef`
- References: Kubernetes Secrets (via `spec.steps[].secrets[].secretRef`)

**Example**:
```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineConfig
metadata:
  name: my-app-ci
  namespace: team-a
spec:
  repository: https://github.com/org/my-app
  branches: ["main", "develop", "feature/*"]
  steps:
    - name: test
      image: golang:1.21
      commands:
        - go test ./...
      resources:
        cpu: 1000m
        memory: 2Gi
      timeout: 10m
    - name: build
      image: golang:1.21
      commands:
        - go build -o app
      dependsOn: [test]
      artifacts:
        - app
        - "dist/**/*"
    - name: push-image
      image: gcr.io/kaniko-project/executor:latest
      commands:
        - /kaniko/executor --dockerfile=Dockerfile --context=. --destination=myregistry/app:$COMMIT_SHA
      dependsOn: [build]
      secrets:
        - secretRef: registry-credentials
          key: config.json
          envVar: DOCKER_CONFIG
  timeout: 1h
```

---

## Entity 2: PipelineRun

**Purpose**: Represents a single execution instance of a PipelineConfig, triggered by webhook, manual trigger, or scheduled event. Tracks execution state, logs, artifacts, and outcomes.

**CRD Definition**:
```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: pipelineruns.c8s.dev
spec:
  group: c8s.dev
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          required: [spec]
          properties:
            spec:
              type: object
              required: [pipelineConfigRef, commit]
              properties:
                pipelineConfigRef:
                  type: string
                  description: Reference to PipelineConfig name
                commit:
                  type: string
                  description: Git commit SHA
                  pattern: '^[a-f0-9]{7,40}$'
                branch:
                  type: string
                  description: Branch name
                triggeredBy:
                  type: string
                  description: User or system that triggered run
                  default: "system"
                triggeredAt:
                  type: string
                  format: date-time
                matrixIndex:
                  type: object
                  description: Matrix combination for this run (if matrix strategy)
                  additionalProperties:
                    type: string
            status:
              type: object
              properties:
                phase:
                  type: string
                  enum: [Pending, Running, Succeeded, Failed, Cancelled]
                  description: Overall pipeline run phase
                startTime:
                  type: string
                  format: date-time
                completionTime:
                  type: string
                  format: date-time
                steps:
                  type: array
                  description: Status of each step
                  items:
                    type: object
                    required: [name, phase]
                    properties:
                      name:
                        type: string
                      phase:
                        type: string
                        enum: [Pending, Running, Succeeded, Failed, Skipped]
                      jobName:
                        type: string
                        description: Kubernetes Job name for this step
                      startTime:
                        type: string
                        format: date-time
                      completionTime:
                        type: string
                        format: date-time
                      exitCode:
                        type: integer
                      logURL:
                        type: string
                        description: Object storage URL for logs
                      artifactURLs:
                        type: array
                        items:
                          type: string
                conditions:
                  type: array
                  description: Conditions for status tracking
                  items:
                    type: object
                    required: [type, status]
                    properties:
                      type:
                        type: string
                      status:
                        type: string
                        enum: [True, False, Unknown]
                      reason:
                        type: string
                      message:
                        type: string
                      lastTransitionTime:
                        type: string
                        format: date-time
      additionalPrinterColumns:
        - name: Status
          type: string
          jsonPath: .status.phase
        - name: Config
          type: string
          jsonPath: .spec.pipelineConfigRef
        - name: Commit
          type: string
          jsonPath: .spec.commit
        - name: Age
          type: date
          jsonPath: .metadata.creationTimestamp
  scope: Namespaced
  names:
    plural: pipelineruns
    singular: pipelinerun
    kind: PipelineRun
    shortNames:
      - pr
```

**State Transitions**:
```
Pending → Running → Succeeded
                 → Failed
                 → Cancelled
```

- **Pending**: PipelineRun created, waiting for controller to schedule Jobs
- **Running**: At least one step Job is running
- **Succeeded**: All steps completed with exitCode 0
- **Failed**: At least one step failed (exitCode != 0) or timed out
- **Cancelled**: User or system cancelled execution

**Validation Rules**:
- `spec.pipelineConfigRef` MUST reference existing PipelineConfig in same namespace
- `spec.commit` MUST be valid Git commit SHA (7-40 hex characters)
- `status.phase` transitions MUST follow state machine (no Succeeded → Running)
- `status.steps[].name` MUST match step names in referenced PipelineConfig
- Once `status.phase` is terminal (Succeeded/Failed/Cancelled), no further updates allowed

**Relationships**:
- References: `PipelineConfig` (via `spec.pipelineConfigRef`)
- Owns: Kubernetes Jobs (one per step, created by controller)
- References: Object storage for logs and artifacts (via `status.steps[].logURL`)

**Example**:
```yaml
apiVersion: c8s.dev/v1alpha1
kind: PipelineRun
metadata:
  name: my-app-ci-abc123
  namespace: team-a
spec:
  pipelineConfigRef: my-app-ci
  commit: abc123def456
  branch: main
  triggeredBy: user@example.com
  triggeredAt: "2025-10-12T10:30:00Z"
status:
  phase: Running
  startTime: "2025-10-12T10:30:05Z"
  steps:
    - name: test
      phase: Succeeded
      jobName: my-app-ci-abc123-test
      startTime: "2025-10-12T10:30:05Z"
      completionTime: "2025-10-12T10:35:00Z"
      exitCode: 0
      logURL: s3://c8s-logs/team-a/my-app-ci-abc123/test.log
    - name: build
      phase: Running
      jobName: my-app-ci-abc123-build
      startTime: "2025-10-12T10:35:05Z"
  conditions:
    - type: JobsCreated
      status: "True"
      reason: AllJobsScheduled
      lastTransitionTime: "2025-10-12T10:30:05Z"
```

---

## Entity 3: RepositoryConnection

**Purpose**: Links a Git repository to the CI system with webhook configuration, authentication, and branch filters. Created by admin to onboard new repositories.

**CRD Definition**:
```yaml
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: repositoryconnections.c8s.dev
spec:
  group: c8s.dev
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          required: [spec]
          properties:
            spec:
              type: object
              required: [repository, provider]
              properties:
                repository:
                  type: string
                  description: Git repository URL
                  pattern: '^(https?|git|ssh)://.*'
                provider:
                  type: string
                  enum: [github, gitlab, bitbucket]
                  description: Version control provider
                webhookSecretRef:
                  type: string
                  description: Kubernetes Secret containing webhook signature secret
                authSecretRef:
                  type: string
                  description: Kubernetes Secret containing Git credentials
                pipelineConfigRef:
                  type: string
                  description: Default PipelineConfig to use for this repository
            status:
              type: object
              properties:
                webhookURL:
                  type: string
                  description: Webhook endpoint URL to configure in provider
                webhookRegistered:
                  type: boolean
                  description: Whether webhook is successfully registered
                lastEvent:
                  type: object
                  properties:
                    type:
                      type: string
                    commit:
                      type: string
                    timestamp:
                      type: string
                      format: date-time
  scope: Namespaced
  names:
    plural: repositoryconnections
    singular: repositoryconnection
    kind: RepositoryConnection
    shortNames:
      - rc
```

**Validation Rules**:
- `spec.repository` MUST be valid Git URL
- `spec.webhookSecretRef` MUST reference existing Secret with `webhook-secret` key
- `spec.authSecretRef` (if specified) MUST reference existing Secret with `username`/`password` or `ssh-key`
- `spec.pipelineConfigRef` MUST reference existing PipelineConfig in same namespace

**Relationships**:
- References: PipelineConfig (via `spec.pipelineConfigRef`)
- References: Kubernetes Secrets (via `spec.webhookSecretRef`, `spec.authSecretRef`)
- Used by: Webhook receiver service to validate incoming webhooks

**Example**:
```yaml
apiVersion: c8s.dev/v1alpha1
kind: RepositoryConnection
metadata:
  name: my-app-repo
  namespace: team-a
spec:
  repository: https://github.com/org/my-app
  provider: github
  webhookSecretRef: my-app-webhook-secret
  authSecretRef: github-access-token
  pipelineConfigRef: my-app-ci
status:
  webhookURL: https://c8s.example.com/webhooks/github
  webhookRegistered: true
  lastEvent:
    type: push
    commit: abc123def456
    timestamp: "2025-10-12T10:30:00Z"
```

---

## Entity 4: Secret (Kubernetes Native)

**Purpose**: Stores sensitive data (credentials, API keys, certificates) that must be injected into pipeline steps securely without logging.

**Schema**: Standard Kubernetes Secret (opaque type)

**Validation Rules**:
- Secrets MUST be in same namespace as PipelineConfig that references them
- Secret keys MUST be valid environment variable names if used as env vars
- Controller MUST mask secret values in logs before persisting to object storage

**Example**:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: database-credentials
  namespace: team-a
type: Opaque
stringData:
  DB_HOST: postgres.example.com
  DB_USER: app_user
  DB_PASSWORD: super-secret-password
```

**Usage in PipelineConfig**:
```yaml
steps:
  - name: test
    image: golang:1.21
    commands:
      - go test ./...
    secrets:
      - secretRef: database-credentials
        key: DB_PASSWORD
        envVar: DATABASE_PASSWORD
```

---

## Entity 5: ResourceQuota (Kubernetes Native)

**Purpose**: Limits compute resources (CPU, memory, storage, concurrent executions) per team/namespace to prevent resource exhaustion.

**Schema**: Standard Kubernetes ResourceQuota

**Validation Rules**:
- Admission webhook MUST reject PipelineRun creation if it would exceed quota
- Controller MUST not create Jobs that would exceed quota limits
- Sum of all step resources in a PipelineRun MUST fit within quota

**Example**:
```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: team-a-quota
  namespace: team-a
spec:
  hard:
    requests.cpu: "100"
    requests.memory: 200Gi
    limits.cpu: "200"
    limits.memory: 400Gi
    pods: "50"
    persistentvolumeclaims: "10"
```

---

## Entity Relationships Diagram

```
┌─────────────────────────┐
│  RepositoryConnection   │
│  ─────────────────────  │
│  + repository           │
│  + provider             │
│  + webhookSecretRef ────┼───────┐
│  + pipelineConfigRef ───┼───┐   │
└─────────────────────────┘   │   │
                              │   │
                              ▼   ▼
┌─────────────────────────┐   ┌─────────────┐
│    PipelineConfig       │   │   Secret    │
│  ───────────────────    │   │  ─────────  │
│  + repository           │   │  + data     │
│  + steps[]              │   └─────────────┘
│    - name               │         ▲
│    - image              │         │
│    - commands           │         │ references
│    - secrets[] ─────────┼─────────┘
│    - dependsOn[]        │
│    - resources          │
└─────────────────────────┘
            │
            │ referenced by
            ▼
┌─────────────────────────┐
│      PipelineRun        │
│  ───────────────────    │
│  + pipelineConfigRef    │
│  + commit               │
│  + branch               │
│  + status               │
│    - phase              │
│    - steps[]            │
│      - jobName ─────────┼───┐
│      - logURL           │   │ owns
│      - artifactURLs     │   │
└─────────────────────────┘   ▼
                          ┌─────────────────┐
                          │ Kubernetes Job  │
                          │  ─────────────  │
                          │  + podTemplate  │
                          │  + completions  │
                          └─────────────────┘
```

---

## Data Flow: Webhook to PipelineRun Execution

1. **Webhook arrives**: GitHub sends push event to `https://c8s.example.com/webhooks/github`
2. **Signature verified**: Webhook service validates HMAC signature using `RepositoryConnection.spec.webhookSecretRef`
3. **RepositoryConnection lookup**: Webhook service queries RepositoryConnection by repository URL
4. **PipelineRun created**: Webhook service creates PipelineRun CRD with commit SHA, branch, triggeredBy
5. **Controller watches**: Controller's informer receives PipelineRun create event
6. **Reconciliation**: Controller reads PipelineConfig, validates quotas, creates Jobs for each step
7. **Jobs execute**: Kubernetes scheduler places Pods, containers run user commands
8. **Status updates**: Controller watches Job status, updates PipelineRun.status.steps[]
9. **Logs persisted**: Controller streams logs from Pods to object storage, updates logURL
10. **Artifacts uploaded**: Sidecar container uploads artifacts to object storage, updates artifactURLs
11. **Completion**: Controller sets PipelineRun.status.phase to Succeeded/Failed

---

## Indexing and Performance

**Recommended Indexes** (via controller-runtime field indexers):

1. **PipelineRun by phase**:
   - Field: `status.phase`
   - Use case: Query all Running pipelines for metrics dashboard

2. **PipelineRun by pipelineConfigRef**:
   - Field: `spec.pipelineConfigRef`
   - Use case: List all runs for a specific pipeline configuration

3. **PipelineRun by commit**:
   - Field: `spec.commit`
   - Use case: Deduplicate identical commits (optional feature)

4. **RepositoryConnection by repository**:
   - Field: `spec.repository`
   - Use case: Webhook receiver lookup by repository URL

**Performance Considerations**:
- **Watch caching**: controller-runtime informers cache all CRDs locally, reducing API server load
- **Object storage**: Logs and artifacts stored in S3, not etcd, avoiding large object storage in CRDs
- **TTL cleanup**: Jobs have `ttlSecondsAfterFinished: 3600` to auto-delete completed workloads
- **Log retention**: Lifecycle policies on S3 bucket delete logs after 30 days

---

## Validation and Admission Control

**Admission Webhooks** (ValidatingWebhookConfiguration):

1. **PipelineConfig validation**:
   - Validate step dependency graph is acyclic (no circular dependencies)
   - Validate all referenced Secrets exist
   - Validate resource requests don't exceed namespace quota

2. **PipelineRun validation**:
   - Validate pipelineConfigRef exists
   - Validate commit SHA format
   - Reject if namespace quota would be exceeded

3. **RepositoryConnection validation**:
   - Validate provider is supported
   - Validate webhook secret exists
   - Validate repository URL format

**Implementation**: Admission webhook runs as separate service, receives admission review requests from K8s API server.

---

## Summary

All entities leverage Kubernetes-native CRDs for state storage, enabling:
- Declarative management via `kubectl` and GitOps
- Built-in RBAC for access control
- Watch API for efficient change detection
- Strong consistency from etcd
- No external database dependencies

Next steps: Generate OpenAPI contracts for REST API endpoints that expose these entities.
