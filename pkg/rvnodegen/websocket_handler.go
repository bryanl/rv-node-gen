package rvnodegen

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"k8s.io/apimachinery/pkg/util/json"

	"github.com/bryanl/rv-node-gen/internal/log"
)

// Command is a web socket command.
type Command struct {
	// Type is the type of command.
	Type string `json:"type"`
	// Payload is the command's payload.
	Payload Payload `json:"payload"`

	messageType int
}

// MessageType returns the web socket message type for the command.
func (c *Command) MessageType() int {
	return c.messageType
}

// CreateResponse creates a web socket response.
func (c *Command) CreateResponse(payload Payload) WebsocketResponse {
	return WebsocketResponse{
		MessageType: c.messageType,
		Payload:     payload,
	}
}

// WebsocketResponse is a web socket response.
type WebsocketResponse struct {
	// MessageType is the type of message.
	MessageType int `json:"messageType"`
	// Payload is the response payload.
	Payload Payload `json:"payload"`
}

// Payload is a web socket response payload.
type Payload map[string]interface{}

// ParseCommand parses a web socket message and creates a command.
func ParseCommand(messageType int, data []byte) (Command, error) {
	var c Command
	if err := json.Unmarshal(data, &c); err != nil {
		return Command{}, fmt.Errorf("unmarshal command: %w", err)
	}

	c.messageType = messageType
	return c, nil
}

// WebsocketHandler is a HTTP handler for handling web socket messages.
type WebsocketHandler struct {
	lister   Lister
	upgrader websocket.Upgrader
}

var _ http.Handler = &WebsocketHandler{}

// NewWebsocketHandler creates an instance of WebsocketHandler.
func NewWebsocketHandler(lister Lister) *WebsocketHandler {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// TODO: this is not safe
			return true
		},
	}
	w := &WebsocketHandler{
		lister:   lister,
		upgrader: upgrader,
	}
	return w
}

// ServeHTTP serves the handler.
func (h *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	commands := CommandsFactory(h.lister)

	logger := log.From(r.Context())

	wc, err := NewWebsocketConn(w, r, h.upgrader)
	if err != nil {
		logger.Error(err, "Create websocket connection")
		return
	}

	defer func() {
		if cErr := wc.Close(); cErr != nil {
			logger.Error(err, "Close web socket connection")
		}
	}()

	for {
		c, err := wc.Read()
		if err != nil {
			logger.Error(err, "Read message")
			break
		}

		if c == nil {
			break
		}

		for _, command := range commands {
			if command.Name() == c.Type {
				if err := command.Run(r.Context(), wc, *c); err != nil {
					logger.Error(err, "Handle command")
					break
				}
			}
		}
	}
}

// WebsocketConn is a web socket connection.
type WebsocketConn struct {
	conn *websocket.Conn
}

// NewWebsocketConn creates an instance of WebsocketConn.
func NewWebsocketConn(w http.ResponseWriter, r *http.Request, upgrader websocket.Upgrader) (*WebsocketConn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, fmt.Errorf("upgrade websocket connection: %w", err)
	}

	wc := &WebsocketConn{
		conn: conn,
	}

	return wc, nil
}

// Close closes the web socket connection.
func (wc *WebsocketConn) Close() error {
	return wc.conn.Close()
}

// Read reads the next message from the connection. It will block until a message is read.
func (wc *WebsocketConn) Read() (*Command, error) {
	messageType, message, err := wc.conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
			return nil, err
		}
		return nil, nil
	}

	c, err := ParseCommand(messageType, message)
	if err != nil {
		return nil, fmt.Errorf("parse command from message: %w", err)
	}

	return &c, nil
}

// Write writes a message to the web socket connection.
func (wc *WebsocketConn) Write(r WebsocketResponse) error {
	data, err := json.Marshal(r.Payload)
	if err != nil {
		return err
	}

	if err := wc.conn.WriteMessage(r.MessageType, data); err != nil {
		return err
	}

	return nil
}
