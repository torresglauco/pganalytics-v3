package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/torresglauco/pganalytics-v3/backend/internal/auth"
	"github.com/torresglauco/pganalytics-v3/backend/pkg/services"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, validate origin properly
		return true
	},
}

// WebSocketHandler handles WebSocket upgrades and manages connections
func WebSocketHandler(wsManager *services.ConnectionManager, jwtManager *auth.JWTManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract and validate JWT token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Parse "Bearer {token}"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate JWT
		claims, err := jwtManager.ValidateUserToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userID := claims.Subject
		// For now, assume user has access to all instances
		// TODO: Query database for user's accessible instances
		instances := []int{1, 2, 3}

		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade error: %v", err)
			return
		}

		// Register connection
		wsConn := wsManager.RegisterConnection(userID, instances, conn)
		log.Printf("WebSocket connection established for user %s", userID)

		// Listen for client messages (heartbeat, etc)
		go func() {
			defer func() {
				wsManager.UnregisterConnection(userID, wsConn)
				log.Printf("WebSocket connection closed for user %s", userID)
			}()

			for {
				message := make(map[string]interface{})
				if err := conn.ReadJSON(&message); err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("WebSocket error: %v", err)
					}
					return
				}

				// Handle client messages (ping/heartbeat)
				if msgType, ok := message["type"].(string); ok && msgType == "ping" {
					wsConn.SendMessage(map[string]string{"type": "pong"})
				}
			}
		}()
	}
}
