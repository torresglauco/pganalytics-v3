package api

import (
	"net/http"
	"strings"

	apperrors "github.com/dextra/pganalytics-v3/backend/pkg/errors"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT tokens
// TODO: Implement JWT validation in Phase 2
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errResp := apperrors.MissingAuthHeader()
			c.JSON(errResp.StatusCode, errResp)
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errResp := apperrors.InvalidToken("Invalid authorization header format")
			c.JSON(errResp.StatusCode, errResp)
			c.Abort()
			return
		}

		token := parts[1]

		// TODO: Validate JWT token
		// For now, accept any non-empty token (not for production)
		if token == "" {
			errResp := apperrors.InvalidToken("Token is empty")
			c.JSON(errResp.StatusCode, errResp)
			c.Abort()
			return
		}

		// TODO: Extract user info from token and store in context
		// c.Set("user_id", userID)
		// c.Set("username", username)
		// c.Set("role", role)

		c.Next()
	}
}

// MTLSMiddleware validates mutual TLS authentication
// TODO: Implement mTLS verification in Phase 2
func (s *Server) MTLSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// In production, verify client certificate from TLS connection
		// For now, this is a placeholder

		if c.Request.TLS == nil {
			// No TLS connection found
			if s.config.IsProduction() {
				errResp := apperrors.InvalidCertificate("TLS connection required")
				c.JSON(errResp.StatusCode, errResp)
				c.Abort()
				return
			}
			// Allow non-TLS in development
		} else if c.Request.TLS.PeerCertificates == nil || len(c.Request.TLS.PeerCertificates) == 0 {
			// No client certificate provided
			if s.config.IsProduction() {
				errResp := apperrors.InvalidCertificate("Client certificate required")
				c.JSON(errResp.StatusCode, errResp)
				c.Abort()
				return
			}
		}

		// TODO: Verify certificate authenticity
		// TODO: Extract certificate thumbprint and verify it's registered
		// TODO: Store collector_id in context for metrics processing

		c.Next()
	}
}

// RoleMiddleware checks if user has required role
func (s *Server) RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Get role from context (set by AuthMiddleware)
		// userRole, exists := c.Get("role")
		// if !exists {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "No role found"})
		// 	c.Abort()
		// 	return
		// }

		// TODO: Check if user role matches required role
		// if userRole != requiredRole {
		// 	c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		// 	c.Abort()
		// 	return
		// }

		c.Next()
	}
}

// ErrorResponseMiddleware converts errors to standard response format
func (s *Server) ErrorResponseMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there were any errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			if appErr, ok := err.Err.(*apperrors.AppError); ok {
				c.JSON(appErr.StatusCode, gin.H{
					"error": appErr.Message,
					"code":  appErr.Code,
				})
				return
			}
		}
	}
}

// LoggingMiddleware logs HTTP requests
func (s *Server) LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Log request details
		s.logger.Debug("Request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"remote_ip", c.ClientIP(),
		)

		c.Next()

		// Log response status
		s.logger.Debug("Response",
			"status_code", c.Writer.Status(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
	}
}

// CORSMiddleware adds CORS headers
func (s *Server) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware limits request rate (placeholder for future implementation)
func (s *Server) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement rate limiting
		// Could use libraries like:
		// - github.com/juju/ratelimit
		// - github.com/throttled/throttled
		c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID for tracing
func (s *Server) RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Generate or extract request ID
		// c.Set("request_id", requestID)
		// c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}
