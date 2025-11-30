// Package apierrors provides centralized error codes and messages for API responses.
package apierrors

// Error codes used in API responses
const (
	// Client errors (4xx)
	CodeInvalidRequest    = "INVALID_REQUEST"
	CodeUnauthorized      = "UNAUTHORIZED"
	CodeForbidden         = "FORBIDDEN"
	CodeNotFound          = "NOT_FOUND"
	CodeValidationFailed  = "VALIDATION_FAILED"
	CodeArtifactNotFound  = "ARTIFACT_NOT_FOUND"
	CodeProjectNotFound   = "PROJECT_NOT_FOUND"
	CodeLogsNotFound      = "LOGS_NOT_FOUND"
	CodeInvalidParam      = "INVALID_PARAM"

	// Server errors (5xx)
	CodeInternalError = "INTERNAL_ERROR"
	CodeFetchFailed   = "FETCH_FAILED"
	CodeCreateFailed  = "CREATE_FAILED"
	CodeDeleteFailed  = "DELETE_FAILED"
)

// Common error messages
const (
	// Authentication/Authorization
	MsgUnauthorized        = "User not authenticated"
	MsgUnauthorizedGeneric = "Unauthorized"
	MsgForbidden           = "Forbidden"

	// Validation
	MsgInvalidRequest     = "Invalid request body"
	MsgValidationFailed   = "Validation failed"
	MsgRunIDRequired      = "runId required"
	MsgProjectIDRequired  = "projectId required"
	MsgArtifactIDRequired = "artifactId required"
	MsgUsernamePassword   = "Username and password required"

	// Not Found
	MsgNotFound        = "The requested resource was not found"
	MsgProjectNotFound = "Project not found"
	MsgArtifactNotFound = "Artifact not found"
	MsgPipelineNotFound = "Pipeline run not found"
	MsgLogsNotFound    = "Logs not found for step"

	// Server Errors
	MsgInternalError      = "An unexpected error occurred"
	MsgInternalGeneric    = "Internal server error"
	MsgFetchFailed        = "Failed to fetch projects"
	MsgCreateFailed       = "Failed to create project"
	MsgDeleteFailed       = "Failed to delete project"
	MsgRenderFailed       = "Failed to render dashboard"
	MsgRenderPipeline     = "Failed to render pipeline details"
	MsgRenderLogin        = "Failed to render login page"
	MsgStreamingUnsupported = "Streaming not supported"
	MsgStreamEstablishFailed = "Failed to establish stream"
)
