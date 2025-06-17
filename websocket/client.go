package websocket

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"

	"chat-server/database"
)

var (
	// Upgrader is used to upgrade HTTP connections to WebSocket connections
	Upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	// Clients holds all connected WebSocket clients
	clients      = make(map[*websocket.Conn]bool)
	clientsMutex = sync.Mutex{}
)

// BroadcastMessage sends a message to all connected clients and saves it to the database
func BroadcastMessage(chatID string, message []byte) error {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			delete(clients, client)
			err := client.Close()
			if err != nil {
				return errors.New("failed to close client connection")
			}
		}
	}

	err := database.SaveMessage(chatID, string(message))
	if err != nil {
		log.Printf("Failed to save message: %v", err)
		return err
	}
	return nil
}

// HandleConnection handles a new WebSocket connection
func HandleConnection(w http.ResponseWriter, r *http.Request) {
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}(conn)

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			break
		}

		err = BroadcastMessage(chatID, msg)
		if err != nil {
			log.Printf("Error broadcasting message: %v", err)
			break
		}
	}
}

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	if err := database.Client.Ping(r.Context(), nil); err != nil {
		http.Error(w, "Database connection failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Database connection is healthy"))
}
