package responses

import (
	"encoding/json"
	"net/http"
)

// APIResponse wraps all API responses in a standardized format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Metadata   `json:"meta,omitempty"`
}

// APIError contains error details
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Metadata contains pagination and other metadata
type Metadata struct {
	Total      int `json:"total,omitempty"`
	Page       int `json:"page,omitempty"`
	PerPage    int `json:"per_page,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

// RespondSuccess sends a successful response with data
func RespondSuccess(w http.ResponseWriter, statusCode int, data interface{}) error {
	return respondJSON(w, statusCode, APIResponse{
		Success: true,
		Data:    data,
	})
}

// RespondSuccessWithMeta sends a successful response with metadata
func RespondSuccessWithMeta(w http.ResponseWriter, statusCode int, data interface{}, meta *Metadata) error {
	return respondJSON(w, statusCode, APIResponse{
		Success: true,
		Data:    data,
		Meta:    meta,
	})
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, statusCode int, code, message string) error {
	return respondJSON(w, statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// RespondErrorWithDetails sends an error response with additional details
func RespondErrorWithDetails(w http.ResponseWriter, statusCode int, code, message string, details interface{}) error {
	return respondJSON(w, statusCode, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// RespondNotFound sends a 404 Not Found response
func RespondNotFound(w http.ResponseWriter, resource string) error {
	return RespondError(w, http.StatusNotFound, "NOT_FOUND", "Resource not found: "+resource)
}

// RespondBadRequest sends a 400 Bad Request response
func RespondBadRequest(w http.ResponseWriter, message string) error {
	return RespondError(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// RespondUnauthorized sends a 401 Unauthorized response
func RespondUnauthorized(w http.ResponseWriter) error {
	return RespondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
}

// RespondForbidden sends a 403 Forbidden response
func RespondForbidden(w http.ResponseWriter) error {
	return RespondError(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
}

// RespondInternalError sends a 500 Internal Server Error response
func RespondInternalError(w http.ResponseWriter, err error) error {
	message := "Internal server error"
	if err != nil {
		message = err.Error()
	}
	return RespondError(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

// respondJSON is the internal helper for sending JSON responses
func respondJSON(w http.ResponseWriter, statusCode int, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(response)
}
