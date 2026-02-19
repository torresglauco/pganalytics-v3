package errors

import (
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
	StatusCode int    `json:"-"` // HTTP status code
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%d] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(statusCode, code int, message, details string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: statusCode,
	}
}

// ============================================================================
// Common Errors
// ============================================================================

// BadRequest represents a 400 Bad Request error
func BadRequest(message, details string) *AppError {
	return NewAppError(http.StatusBadRequest, 400, message, details)
}

// Unauthorized represents a 401 Unauthorized error
func Unauthorized(message, details string) *AppError {
	return NewAppError(http.StatusUnauthorized, 401, message, details)
}

// Forbidden represents a 403 Forbidden error
func Forbidden(message, details string) *AppError {
	return NewAppError(http.StatusForbidden, 403, message, details)
}

// NotFound represents a 404 Not Found error
func NotFound(message, details string) *AppError {
	return NewAppError(http.StatusNotFound, 404, message, details)
}

// Conflict represents a 409 Conflict error
func Conflict(message, details string) *AppError {
	return NewAppError(http.StatusConflict, 409, message, details)
}

// InternalServerError represents a 500 Internal Server Error
func InternalServerError(message, details string) *AppError {
	return NewAppError(http.StatusInternalServerError, 500, message, details)
}

// ServiceUnavailable represents a 503 Service Unavailable error
func ServiceUnavailable(message, details string) *AppError {
	return NewAppError(http.StatusServiceUnavailable, 503, message, details)
}

// ============================================================================
// Specific Errors
// ============================================================================

// InvalidCredentials represents authentication failure
func InvalidCredentials() *AppError {
	return Unauthorized("Invalid credentials", "Username or password is incorrect")
}

// TokenExpired represents an expired JWT token
func TokenExpired() *AppError {
	return Unauthorized("Token expired", "JWT token has expired")
}

// InvalidToken represents an invalid JWT token
func InvalidToken(details string) *AppError {
	return Unauthorized("Invalid token", details)
}

// MissingAuthHeader represents missing Authorization header
func MissingAuthHeader() *AppError {
	return Unauthorized("Missing authorization header", "Authorization header is required")
}

// InvalidCertificate represents certificate validation failure
func InvalidCertificate(details string) *AppError {
	return Unauthorized("Invalid certificate", details)
}

// CollectorNotFound represents a missing collector
func CollectorNotFound(id string) *AppError {
	return NotFound("Collector not found", fmt.Sprintf("Collector with ID %s not found", id))
}

// CollectorAlreadyExists represents duplicate collector registration
func CollectorAlreadyExists(hostname string) *AppError {
	return Conflict("Collector already exists", fmt.Sprintf("A collector with hostname %s is already registered", hostname))
}

// UserNotFound represents a missing user
func UserNotFound(username string) *AppError {
	return NotFound("User not found", fmt.Sprintf("User %s not found", username))
}

// InvalidJSON represents JSON parsing error
func InvalidJSON(details string) *AppError {
	return BadRequest("Invalid JSON", details)
}

// ValidationError represents data validation failure
func ValidationError(field, reason string) *AppError {
	return BadRequest("Validation failed", fmt.Sprintf("Field '%s': %s", field, reason))
}

// DatabaseError represents a database operation error
func DatabaseError(operation, details string) *AppError {
	return InternalServerError(
		"Database error",
		fmt.Sprintf("Failed to %s: %s", operation, details),
	)
}

// ConfigurationError represents a configuration issue
func ConfigurationError(details string) *AppError {
	return InternalServerError("Configuration error", details)
}

// ExternalServiceError represents a dependency/external service error
func ExternalServiceError(service, details string) *AppError {
	return ServiceUnavailable(
		fmt.Sprintf("%s unavailable", service),
		details,
	)
}

// ============================================================================
// Conversion Helpers
// ============================================================================

// ToAppError converts any error to an AppError if it isn't already
func ToAppError(err error) *AppError {
	if appErr, ok := err.(*AppError); ok {
		return appErr
	}

	// Unknown error
	return InternalServerError("Unknown error", err.Error())
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}
