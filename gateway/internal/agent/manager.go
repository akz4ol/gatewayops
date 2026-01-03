package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// Manager handles agent connections and message routing.
type Manager struct {
	logger      zerolog.Logger
	connections map[uuid.UUID]*Connection
	mu          sync.RWMutex
	upgrader    websocket.Upgrader

	// Metrics
	totalConnections    int64
	activeConnections   int64
	totalMessages       int64
	connectionsByPlatform map[string]int
}

// NewManager creates a new agent connection manager.
func NewManager(logger zerolog.Logger) *Manager {
	return &Manager{
		logger:      logger,
		connections: make(map[uuid.UUID]*Connection),
		connectionsByPlatform: make(map[string]int),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking
				return true
			},
		},
	}
}

// Connect establishes a new agent connection.
func (m *Manager) Connect(ctx context.Context, req ConnectRequest, orgID, userID uuid.UUID) (*Connection, error) {
	conn := &Connection{
		ID:           uuid.New(),
		AgentID:      req.AgentID,
		Platform:     req.Platform,
		OrgID:        orgID,
		UserID:       userID,
		Transport:    req.Transport,
		State:        StateConnecting,
		Capabilities: req.Capabilities,
		CallbackURL:  req.CallbackURL,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
		LastActiveAt: time.Now(),
		sendCh:       make(chan []byte, 256),
		done:         make(chan struct{}),
	}

	m.mu.Lock()
	m.connections[conn.ID] = conn
	m.totalConnections++
	m.activeConnections++
	m.connectionsByPlatform[req.Platform]++
	m.mu.Unlock()

	m.logger.Info().
		Str("connection_id", conn.ID.String()).
		Str("platform", req.Platform).
		Str("agent_id", req.AgentID).
		Str("transport", string(req.Transport)).
		Msg("Agent connection created")

	return conn, nil
}

// UpgradeToWebSocket upgrades an HTTP connection to WebSocket.
func (m *Manager) UpgradeToWebSocket(w http.ResponseWriter, r *http.Request, connID uuid.UUID) error {
	m.mu.Lock()
	conn, exists := m.connections[connID]
	m.mu.Unlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	ws, err := m.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade failed: %w", err)
	}

	conn.mu.Lock()
	conn.ws = ws
	conn.State = StateConnected
	conn.mu.Unlock()

	// Start read and write pumps
	go m.readPump(conn)
	go m.writePump(conn)

	m.logger.Info().
		Str("connection_id", conn.ID.String()).
		Msg("WebSocket connection established")

	return nil
}

// readPump reads messages from the WebSocket connection.
func (m *Manager) readPump(conn *Connection) {
	defer func() {
		m.Disconnect(conn.ID)
	}()

	conn.ws.SetReadLimit(512 * 1024) // 512KB max message size
	conn.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.ws.SetPongHandler(func(string) error {
		conn.ws.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := conn.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				m.logger.Warn().Err(err).Str("connection_id", conn.ID.String()).Msg("WebSocket read error")
			}
			break
		}

		conn.mu.Lock()
		conn.LastActiveAt = time.Now()
		conn.mu.Unlock()

		m.mu.Lock()
		m.totalMessages++
		m.mu.Unlock()

		// Process message
		go m.handleMessage(conn, message)
	}
}

// writePump writes messages to the WebSocket connection.
func (m *Manager) writePump(conn *Connection) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		conn.ws.Close()
	}()

	for {
		select {
		case message, ok := <-conn.sendCh:
			conn.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				conn.ws.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := conn.ws.WriteMessage(websocket.TextMessage, message); err != nil {
				m.logger.Warn().Err(err).Str("connection_id", conn.ID.String()).Msg("WebSocket write error")
				return
			}

		case <-ticker.C:
			conn.ws.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.ws.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}

		case <-conn.done:
			return
		}
	}
}

// handleMessage processes an incoming WebSocket message.
func (m *Manager) handleMessage(conn *Connection, data []byte) {
	var msg WSMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		m.sendError(conn, "", "invalid_message", "Failed to parse message")
		return
	}

	switch msg.Type {
	case WSTypePing:
		m.send(conn, WSMessage{Type: WSTypePong})

	case WSTypeToolCall:
		m.handleToolCall(conn, msg)

	case WSTypeCancel:
		m.handleCancel(conn, msg)

	default:
		m.sendError(conn, msg.ID, "unknown_message_type", fmt.Sprintf("Unknown message type: %s", msg.Type))
	}
}

// handleToolCall processes a tool call request.
func (m *Manager) handleToolCall(conn *Connection, msg WSMessage) {
	// Extract tool call from payload
	payloadBytes, err := json.Marshal(msg.Payload)
	if err != nil {
		m.sendError(conn, msg.ID, "invalid_payload", "Failed to process payload")
		return
	}

	var call ToolCall
	if err := json.Unmarshal(payloadBytes, &call); err != nil {
		m.sendError(conn, msg.ID, "invalid_tool_call", "Invalid tool call format")
		return
	}
	call.ID = msg.ID

	// TODO: Execute tool call through MCP handler
	// For now, send a mock response
	m.send(conn, WSMessage{
		Type: WSTypeToolResult,
		ID:   msg.ID,
		Payload: ToolResult{
			ID:     msg.ID,
			Status: "success",
			Content: []ContentBlock{
				{Type: "text", Text: fmt.Sprintf("Tool %s.%s executed (mock)", call.Server, call.Tool)},
			},
			DurationMs: 50,
			Cost:       0.0001,
		},
	})
}

// handleCancel processes a cancel request.
func (m *Manager) handleCancel(conn *Connection, msg WSMessage) {
	// TODO: Implement cancellation
	m.logger.Info().
		Str("connection_id", conn.ID.String()).
		Str("message_id", msg.ID).
		Msg("Cancel request received")
}

// send sends a message to the connection.
func (m *Manager) send(conn *Connection, msg WSMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		m.logger.Error().Err(err).Msg("Failed to marshal message")
		return
	}

	select {
	case conn.sendCh <- data:
	default:
		m.logger.Warn().Str("connection_id", conn.ID.String()).Msg("Send channel full, dropping message")
	}
}

// sendError sends an error message to the connection.
func (m *Manager) sendError(conn *Connection, msgID, code, message string) {
	m.send(conn, WSMessage{
		Type: WSTypeError,
		ID:   msgID,
		Payload: ErrorInfo{
			Code:    code,
			Message: message,
		},
	})
}

// Disconnect closes and removes a connection.
func (m *Manager) Disconnect(connID uuid.UUID) {
	m.mu.Lock()
	conn, exists := m.connections[connID]
	if !exists {
		m.mu.Unlock()
		return
	}

	delete(m.connections, connID)
	m.activeConnections--
	m.connectionsByPlatform[conn.Platform]--
	m.mu.Unlock()

	conn.mu.Lock()
	conn.State = StateDisconnected
	close(conn.done)
	if conn.ws != nil {
		conn.ws.Close()
	}
	conn.mu.Unlock()

	m.logger.Info().
		Str("connection_id", connID.String()).
		Msg("Agent connection closed")
}

// GetConnection returns a connection by ID.
func (m *Manager) GetConnection(connID uuid.UUID) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[connID]
	return conn, exists
}

// GetStats returns connection statistics.
func (m *Manager) GetStats() ConnectionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	byPlatform := make(map[string]int)
	for k, v := range m.connectionsByPlatform {
		byPlatform[k] = v
	}

	return ConnectionStats{
		Active:     int(m.activeConnections),
		Total:      int(m.totalConnections),
		Messages:   int(m.totalMessages),
		ByPlatform: byPlatform,
	}
}

// ConnectionStats holds connection statistics.
type ConnectionStats struct {
	Active     int            `json:"active"`
	Total      int            `json:"total"`
	Messages   int            `json:"messages"`
	ByPlatform map[string]int `json:"by_platform"`
}
