package dashboard

import "time"

// PipelineRunDTO represents a pipeline run for dashboard display
type PipelineRunDTO struct {
	ID              string     `json:"id"`
	ProjectID       string     `json:"project_id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"` // Pending, Running, Succeeded, Failed, Canceled
	CommitSHA       string     `json:"commit_sha"`
	CommitMessage   string     `json:"commit_message"`
	Branch          string     `json:"branch"`
	Author          string     `json:"author"`
	AuthorEmail     string     `json:"author_email"`
	TriggerSource   string     `json:"trigger_source"`
	TriggeredAt     time.Time  `json:"triggered_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	DurationSeconds *int64     `json:"duration_seconds,omitempty"`
	StepCount       int        `json:"step_count"`
	SuccessCount    int        `json:"success_count"`
	FailureCount    int        `json:"failure_count"`
	ArtifactCount   int        `json:"artifact_count"`
}

// StepDTO represents a pipeline step for dashboard display
type StepDTO struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Status          string     `json:"status"` // Pending, Running, Succeeded, Failed, Skipped
	Image           string     `json:"image"`
	Commands        []string   `json:"commands"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	DurationSeconds *int64     `json:"duration_seconds,omitempty"`
	ExitCode        *int32     `json:"exit_code,omitempty"`
	DependsOn       []string   `json:"depends_on,omitempty"`
	LogURL          string     `json:"log_url,omitempty"`
	CPURequest      string     `json:"cpu_request,omitempty"`
	MemoryRequest   string     `json:"memory_request,omitempty"`
}

// ProjectDTO represents a project for dashboard display
type ProjectDTO struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	RepoURL     string     `json:"repository_url"`
	WebhookURL  string     `json:"webhook_url"`
	Namespace   string     `json:"namespace"`
	CreatedAt   time.Time  `json:"created_at"`
	LastRunAt   *time.Time `json:"last_run_at,omitempty"`
	RunCount    int        `json:"run_count"`
}

// ArtifactDTO represents an artifact for dashboard display
type ArtifactDTO struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // binary, report, documentation, log
	MimeType  string    `json:"mime_type"`
	SizeBytes int64     `json:"size_bytes"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}

// LogStreamDTO represents a log stream chunk
type LogStreamDTO struct {
	StepID    string    `json:"step_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	IsError   bool      `json:"is_error"`
}

// PipelineConfigDTO represents a pipeline configuration
type PipelineConfigDTO struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	RepoURL     string    `json:"repository_url"`
	Branches    []string  `json:"branches"`
	Timeout     int       `json:"timeout"`
	CreatedAt   time.Time `json:"created_at"`
}

// ActivityDTO represents an activity feed entry for dashboard display
type ActivityDTO struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`       // deployment, build, error, commit
	Message   string    `json:"message"`
	User      string    `json:"user"`
	Timestamp time.Time `json:"timestamp"`
	ProjectID string    `json:"project_id,omitempty"`
	RunID     string    `json:"run_id,omitempty"`
}
