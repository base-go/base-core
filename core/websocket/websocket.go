package websocket

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins. In production, you might want to restrict this.
	},
}

// Client represents a WebSocket client
type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan []byte
}

// Message represents a message structure
type Message struct {
	Type    string      `json:"type"`
	Content interface{} `json:"content"`
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      *sync.Mutex
}

// NewHub creates a new Hub instance
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		mutex:      &sync.Mutex{},
	}
}

// Run starts the Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mutex.Unlock()
		case message := <-h.broadcast:
			h.mutex.Lock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mutex.Unlock()
		}
	}
}

// ServeWs handles WebSocket requests from the peer
func ServeWs(hub *Hub, c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error(err)
		return
	}
	client := &Client{ID: c.Query("id"), Conn: conn, Send: make(chan []byte, 256)}
	hub.register <- client

	go client.writePump()
	go client.readPump(hub)
}

func (c *Client) readPump(hub *Hub) {
	defer func() {
		hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Errorf("error: %v", err)
			}
			break
		}
		hub.broadcast <- message
	}
}

func (c *Client) writePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		w, err := c.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		// Add queued messages to the current websocket message
		n := len(c.Send)
		for i := 0; i < n; i++ {
			w.Write([]byte{'\n'})
			w.Write(<-c.Send)
		}

		if err := w.Close(); err != nil {
			return
		}
	}
}

// BroadcastMessage sends a message to all connected clients
func (h *Hub) BroadcastMessage(messageType string, content interface{}) {
	message := Message{
		Type:    messageType,
		Content: content,
	}
	jsonMessage, err := json.Marshal(message)
	if err != nil {
		log.Errorf("Failed to marshal message: %v", err)
		return
	}
	h.broadcast <- jsonMessage
}

// InitWebSocketModule initializes the WebSocket module
func InitWebSocketModule(router *gin.RouterGroup) *Hub {
	log.Info("Initializing WebSocket module")
	hub := NewHub()
	go hub.Run()
	SetupWebSocketRoutes(router, hub)
	return hub
}

// SetupWebSocketRoutes sets up the WebSocket routes
func SetupWebSocketRoutes(router *gin.RouterGroup, hub *Hub) {
	router.GET("/ws", WebSocketHandler(hub))
}

// WebSocketHandler returns a gin.HandlerFunc for handling WebSocket connections
// @Summary Connect to WebSocket
// @Description Establishes a WebSocket connection
// @Tags Core/Websocket
// @Accept  json
// @Produce  json
// @Param id query string false "Client ID"
// @Success 101 {string} string "Switching Protocols"
// @Failure 400 {object} ErrorResponse
// @Router /ws [get]
func WebSocketHandler(hub *Hub) gin.HandlerFunc {
	return func(c *gin.Context) {
		ServeWs(hub, c)
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
