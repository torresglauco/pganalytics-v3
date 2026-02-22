package api

import (
	"net/http"

	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	apperrors "github.com/torresglauco/pganalytics-v3/backend/pkg/errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthMiddleware validates JWT tokens
func (s *Server) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")

		// Extract token using helper
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			errResp := apperrors.ToAppError(err)
			c.JSON(errResp.StatusCode, errResp)
			c.Abort()
			return
		}

		// Validate JWT token
		claims, err := s.jwtManager.ValidateUserToken(token)
		if err != nil {
			errResp := apperrors.ToAppError(err)
			c.JSON(errResp.StatusCode, errResp)
			c.Abort()
			return
		}

		// Store user info in context for handlers to use
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("claims", claims)

		s.logger.Debug("User authenticated",
			zap.Int("user_id", claims.UserID),
			zap.String("username", claims.Username),
			zap.String("role", claims.Role),
		)

		c.Next()
	}
}

// CollectorAuthMiddleware validates collector JWT tokens
func (s *Server) CollectorAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")

		// Extract token using helper
		token, err := auth.ExtractTokenFromHeader(authHeader)
		if err != nil {
			errResp := apperrors.ToAppError(err)
			c.JSON(errResp.StatusCode, errResp)
			c.Abort()
			return
		}

		// Validate collector JWT token
		claims, err := s.jwtManager.ValidateCollectorToken(token)
		if err != nil {
			errResp := apperrors.ToAppError(err)
			c.JSON(errResp.StatusCode, errResp)
			c.Abort()
			return
		}

		// Store collector info in context
		c.Set("collector_id", claims.CollectorID)
		c.Set("hostname", claims.Hostname)
		c.Set("collector_claims", claims)

		s.logger.Debug("Collector authenticated",
			zap.String("collector_id", claims.CollectorID),
			zap.String("hostname", claims.Hostname),
		)

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
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_ip", c.ClientIP()),
		)

		c.Next()

		// Log response status
		s.logger.Debug("Response",
			zap.Int("status_code", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
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
