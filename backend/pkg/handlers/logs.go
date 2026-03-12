package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/torresglauco/pganalytics-v3/backend/internal/storage"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/models"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/services"
)

// IngestLogsRequest is the request body for POST /api/v1/logs/ingest
type IngestLogsRequest struct {
	CollectorID string                 `json:"collector_id"`
	InstanceID  int                    `json:"instance_id"`
	Logs        []map[string]interface{} `json:"logs"`
}

// IngestLogsResponse is the response body
type IngestLogsResponse struct {
	Success   bool     `json:"success"`
	Ingested  int      `json:"ingested"`
	Errors    []string `json:"errors,omitempty"`
	Message   string   `json:"message,omitempty"`
}

// IngestLogs handles POST /api/v1/logs/ingest
func IngestLogs(db storage.Storage, wsManager *services.ConnectionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Validate method
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract API token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(IngestLogsResponse{
				Success: false,
				Message: "Missing authorization header",
			})
			return
		}

		// Validate token format "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(IngestLogsResponse{
				Success: false,
				Message: "Invalid authorization header format",
			})
			return
		}

		apiToken := parts[1]
		// TODO: Validate API token against database
		_ = apiToken

		// Parse request body
		var req IngestLogsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(IngestLogsResponse{
				Success: false,
				Message: "Malformed request body",
			})
			return
		}

		// Validate collector_id and instance_id
		if req.CollectorID == "" || req.InstanceID <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(IngestLogsResponse{
				Success: false,
				Message: "Invalid collector_id or instance_id",
			})
			return
		}

		// Process each log
		ingestedCount := 0
		errors := []string{}

		for i, logData := range req.Logs {
			// Validate required fields
			timestamp, ok := logData["timestamp"].(string)
			if !ok {
				errors = append(errors, "Log "+string(rune(i))+": missing timestamp")
				continue
			}

			level, ok := logData["level"].(string)
			if !ok {
				errors = append(errors, "Log "+string(rune(i))+": missing level")
				continue
			}

			message, ok := logData["message"].(string)
			if !ok {
				errors = append(errors, "Log "+string(rune(i))+": missing message")
				continue
			}

			// Validate level is ERROR or SLOW_QUERY
			if level != "ERROR" && level != "SLOW_QUERY" {
				errors = append(errors, "Log "+string(rune(i))+": invalid level "+level)
				continue
			}

			// Parse timestamp
			parsedTime, err := time.Parse(time.RFC3339, timestamp)
			if err != nil {
				errors = append(errors, "Log "+string(rune(i))+": invalid timestamp")
				continue
			}

			// Validate timestamp not in future or too old (>24h)
			now := time.Now()
			if parsedTime.After(now) {
				errors = append(errors, "Log "+string(rune(i))+": timestamp in future")
				continue
			}

			if now.Sub(parsedTime) > 24*time.Hour {
				errors = append(errors, "Log "+string(rune(i))+": timestamp older than 24h")
				continue
			}

			// Create PostgreSQLLog model
			pgLog := &models.PostgreSQLLog{
				CollectorID:    req.CollectorID,
				InstanceID:     req.InstanceID,
				LogTimestamp:   parsedTime,
				LogLevel:       level,
				LogMessage:     message,
				SourceLocation: getOptionalString(logData, "source_location"),
				ProcessID:      getOptionalInt(logData, "process_id"),
				QueryText:      getOptionalString(logData, "query_text"),
				QueryHash:      getOptionalInt64(logData, "query_hash"),
				ErrorCode:      getOptionalString(logData, "error_code"),
				ErrorDetail:    getOptionalString(logData, "error_detail"),
				ErrorHint:      getOptionalString(logData, "error_hint"),
				ErrorContext:   getOptionalString(logData, "error_context"),
				UserName:       getOptionalString(logData, "user_name"),
				ConnectionFrom: getOptionalString(logData, "connection_from"),
				SessionID:      getOptionalString(logData, "session_id"),
			}

			// Insert log into database
			// TODO: Use actual storage method when available
			_ = pgLog
			ingestedCount++

			// Broadcast WebSocket event
			wsManager.BroadcastLogEvent(map[string]interface{}{
				"id":        i,
				"timestamp": timestamp,
				"level":     level,
				"message":   message,
				"instance_id": req.InstanceID,
			}, req.InstanceID)

			log.Printf("Ingested log: level=%s, instance=%d", level, req.InstanceID)
		}

		// Return response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(IngestLogsResponse{
			Success:  true,
			Ingested: ingestedCount,
			Errors:   errors,
		})
	}
}

// Helper functions
func getOptionalString(data map[string]interface{}, key string) *string {
	if val, ok := data[key].(string); ok {
		return &val
	}
	return nil
}

func getOptionalInt(data map[string]interface{}, key string) *int {
	if val, ok := data[key].(float64); ok {
		intVal := int(val)
		return &intVal
	}
	return nil
}

func getOptionalInt64(data map[string]interface{}, key string) *int64 {
	if val, ok := data[key].(float64); ok {
		int64Val := int64(val)
		return &int64Val
	}
	return nil
}
