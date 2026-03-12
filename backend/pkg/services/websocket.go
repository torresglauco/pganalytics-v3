package services

import (
	"sync"

	"github.com/gorilla/websocket"
)

// ConnectionManager manages all active WebSocket connections
type ConnectionManager struct {
	connections map[string][]*Connection // userID -> list of connections
	mu          sync.RWMutex
}

// Connection represents a single WebSocket connection
type Connection struct {
	userID    string
	instances []int // instances this user can access
	conn      *websocket.Conn
	send      chan interface{}
	done      chan bool
}

// WebSocketEvent represents an event sent to clients
type WebSocketEvent struct {
	Type string      `json:"type"` // log:new, metric:update, alert:triggered
	Data interface{} `json:"data"`
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string][]*Connection),
	}
}

// RegisterConnection registers a new WebSocket connection
func (cm *ConnectionManager) RegisterConnection(userID string, instances []int, conn *websocket.Conn) *Connection {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	c := &Connection{
		userID:    userID,
		instances: instances,
		conn:      conn,
		send:      make(chan interface{}, 256),
		done:      make(chan bool),
	}

	cm.connections[userID] = append(cm.connections[userID], c)
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
func (cm *ConnectionManager) BroadcastLogEvent(log interface{}, instanceID int) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	event := WebSocketEvent{
		Type: "log:new",
		Data: log,
	}

	for _, conns := range cm.connections {
		for _, c := range conns {
			if c.hasAccessToInstance(instanceID) {
				select {
				case c.send <- event:
				default:
					// Channel full, skip
				}
			}
		}
	}
}

// BroadcastMetricEvent sends a metric:update event
func (cm *ConnectionManager) BroadcastMetricEvent(data interface{}, instanceID int) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	event := WebSocketEvent{
		Type: "metric:update",
		Data: data,
	}

	for _, conns := range cm.connections {
		for _, c := range conns {
			if c.hasAccessToInstance(instanceID) {
				select {
				case c.send <- event:
				default:
				}
			}
		}
	}
}

// BroadcastAlertEvent sends an alert:triggered event
func (cm *ConnectionManager) BroadcastAlertEvent(data interface{}, instanceID int) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	event := WebSocketEvent{
		Type: "alert:triggered",
		Data: data,
	}

	for _, conns := range cm.connections {
		for _, c := range conns {
			if c.hasAccessToInstance(instanceID) {
				select {
				case c.send <- event:
				default:
				}
			}
		}
	}
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
	defer func() {
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				return
			}

		case <-c.done:
			return
		}
	}
}

func (c *Connection) Close() {
	close(c.done)
}

// SendMessage sends a message to the client
func (c *Connection) SendMessage(msg interface{}) {
	select {
	case c.send <- msg:
	default:
		// Channel full, skip
	}
}
