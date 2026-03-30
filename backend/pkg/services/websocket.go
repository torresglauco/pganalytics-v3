package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// Configuration constants
const (
	writeWait      = 10 * time.Second   // Time allowed to write a message to the peer
	pongWait       = 60 * time.Second   // Time allowed to read the next pong message
	pingInterval   = 30 * time.Second   // Send pings to peer with this period
	maxMessageSize = 512 * 1024         // Max message size in bytes (512KB)
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker provides circuit breaking functionality
type CircuitBreaker struct {
	mu           sync.RWMutex
	state        CircuitBreakerState
	failureCount int
	maxFailures  int
	lastFailTime time.Time
	timeout      time.Duration
	successCount int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.state == StateClosed {
		return false
	}

	if cb.state == StateOpen {
		// Check if timeout has passed
		if time.Since(cb.lastFailTime) > cb.timeout {
			// Transition to half-open
			cb.mu.RUnlock()
			cb.mu.Lock()
			cb.state = StateHalfOpen
			cb.failureCount = 0
			cb.successCount = 0
			cb.mu.Unlock()
			cb.mu.RLock()
			return false
		}
		return true
	}

	// Half-open state
	return false
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= 3 {
			// Close the circuit after 3 successes
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	} else if cb.state == StateClosed {
		// Reset failure count on success in closed state
		cb.failureCount = 0
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailTime = time.Now()

	if cb.state == StateClosed && cb.failureCount >= cb.maxFailures {
		// Open the circuit
		cb.state = StateOpen
	} else if cb.state == StateHalfOpen {
		// Go back to open
		cb.state = StateOpen
		cb.failureCount = 0
		cb.successCount = 0
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// ConnectionManager manages all active WebSocket connections
type ConnectionManager struct {
	connections map[string][]*Connection // userID -> list of connections
	mu          sync.RWMutex

	broadcastCB *CircuitBreaker
	logger      *zap.Logger
}

// Connection represents a single WebSocket connection
type Connection struct {
	id           string
	userID       string
	instances    []int // instances this user can access
	conn         *websocket.Conn
	send         chan interface{}
	done         chan struct{}
	lastPongTime time.Time
}

// WebSocketEvent represents an event sent to clients
type WebSocketEvent struct {
	Type string      `json:"type"` // log:new, metric:update, alert:triggered
	Data interface{} `json:"data"`
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(logger *zap.Logger) *ConnectionManager {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ConnectionManager{
		connections: make(map[string][]*Connection),
		broadcastCB: NewCircuitBreaker(5, 30*time.Second), // Open after 5 failures, retry after 30s
		logger:      logger,
	}
}

// RegisterConnection registers a new WebSocket connection
func (cm *ConnectionManager) RegisterConnection(userID string, instances []int, conn *websocket.Conn) *Connection {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	c := &Connection{
		id:           fmt.Sprintf("%s-%d", userID, time.Now().UnixNano()),
		userID:       userID,
		instances:    instances,
		conn:         conn,
		send:         make(chan interface{}, 256),
		done:         make(chan struct{}),
		lastPongTime: time.Now(),
	}

	cm.connections[userID] = append(cm.connections[userID], c)
	go c.readPump(cm)
	go c.writePump()
	return c
}

// UnregisterConnection removes a connection
func (cm *ConnectionManager) UnregisterConnection(userID string, c *Connection) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conns, ok := cm.connections[userID]
	if !ok {
		return
	}

	for i, conn := range conns {
		if conn == c {
			cm.connections[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	if len(cm.connections[userID]) == 0 {
		delete(cm.connections, userID)
	}
}

// BroadcastLogEvent sends a log:new event to all connected users with access
func (cm *ConnectionManager) BroadcastLogEvent(log interface{}, instanceID int) error {
	event := WebSocketEvent{
		Type: "log:new",
		Data: log,
	}

	return cm.broadcastWithFallback(event, instanceID)
}

// BroadcastMetricEvent sends a metric:update event
func (cm *ConnectionManager) BroadcastMetricEvent(data interface{}, instanceID int) error {
	event := WebSocketEvent{
		Type: "metric:update",
		Data: data,
	}

	return cm.broadcastWithFallback(event, instanceID)
}

// BroadcastAlertEvent sends an alert:triggered event
func (cm *ConnectionManager) BroadcastAlertEvent(data interface{}, instanceID int) error {
	event := WebSocketEvent{
		Type: "alert:triggered",
		Data: data,
	}

	return cm.broadcastWithFallback(event, instanceID)
}

// broadcastWithFallback sends a message with fallback mechanism for backpressure
func (cm *ConnectionManager) broadcastWithFallback(event WebSocketEvent, instanceID int) error {
	// Check circuit breaker
	if cm.broadcastCB.IsOpen() {
		cm.logger.Warn("Broadcast circuit breaker is open",
			zap.String("event_type", event.Type),
			zap.Int("instance_id", instanceID))
		return fmt.Errorf("broadcast circuit breaker open")
	}

	cm.mu.RLock()
	// Create a copy of connections to avoid holding lock during send
	var connections []*Connection
	for _, conns := range cm.connections {
		for _, c := range conns {
			if c.hasAccessToInstance(instanceID) {
				connections = append(connections, c)
			}
		}
	}
	cm.mu.RUnlock()

	if len(connections) == 0 {
		return nil
	}

	successCount := 0
	timeoutCount := 0

	for _, conn := range connections {
		// Try to send with timeout to prevent blocking broadcaster
		select {
		case conn.send <- event:
			successCount++
		case <-time.After(100 * time.Millisecond):
			timeoutCount++
			cm.logger.Warn("WebSocket send queue full",
				zap.String("connection_id", conn.id),
				zap.String("user_id", conn.userID),
				zap.String("event_type", event.Type))
		}
	}

	// Log if we had to drop messages
	if timeoutCount > 0 {
		cm.logger.Warn("Dropped WebSocket messages due to backpressure",
			zap.String("event_type", event.Type),
			zap.Int("instance_id", instanceID),
			zap.Int("successful", successCount),
			zap.Int("timed_out", timeoutCount),
			zap.Int("total_connections", len(connections)))
	}

	// Record failure if more than half the connections timed out
	if timeoutCount > len(connections)/2 {
		cm.broadcastCB.RecordFailure()
	} else {
		cm.broadcastCB.RecordSuccess()
	}

	return nil
}

// Connection methods

func (c *Connection) hasAccessToInstance(instanceID int) bool {
	for _, id := range c.instances {
		if id == instanceID {
			return true
		}
	}
	return false
}

func (c *Connection) writePump() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// Set pong handler to track heartbeat responses
	c.conn.SetPongHandler(func(appData string) error {
		c.lastPongTime = time.Now()
		return nil
	})

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// Channel closed, send close message
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Create a new writer for the message
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			// Marshal message to JSON
			data, err := json.Marshal(message)
			if err != nil {
				w.Close()
				return
			}

			_, err = w.Write(data)
			if err != nil {
				w.Close()
				return
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			// Send ping to detect stale connections
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

			// Check for unresponsive connections (no pong in 90 seconds)
			if time.Since(c.lastPongTime) > 90*time.Second {
				c.conn.Close()
				return
			}

		case <-c.done:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}
	}
}

// readPump handles incoming messages from the client
func (c *Connection) readPump(manager *ConnectionManager) {
	defer func() {
		manager.UnregisterConnection(c.userID, c)
		c.conn.Close()
	}()

	// Set read limit
	c.conn.SetReadLimit(maxMessageSize)

	// Set initial deadline
	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	// Handle pong messages
	c.conn.SetPongHandler(func(string) error {
		c.lastPongTime = time.Now()
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg map[string]interface{}

		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				manager.logger.Warn("WebSocket unexpected close error",
					zap.String("connection_id", c.id),
					zap.String("user_id", c.userID),
					zap.Error(err))
			}
			break
		}

		// Process message (client can send ping/heartbeat messages)
		if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
			c.SendMessage(map[string]string{"type": "pong"})
		}
	}
}

// Close closes the connection
func (c *Connection) Close() {
	select {
	case <-c.done:
		// Already closed
	default:
		close(c.done)
	}
}

// SendMessage sends a message to the client
func (c *Connection) SendMessage(msg interface{}) {
	select {
	case c.send <- msg:
	default:
		// Channel full, skip
	}
}
