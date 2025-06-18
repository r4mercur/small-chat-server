package websocket

import (
	"encoding/json"
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

	// Parse the message to extract sender and content
	// Expected format: {"sender":"username","content":"message text"}
	var msgData struct {
		Sender  string `json:"sender"`
		Content string `json:"content"`
	}

	err := json.Unmarshal(message, &msgData)
	if err != nil {
		// If not in JSON format, treat the whole message as content with "anonymous" sender
		msgData.Content = string(message)
		msgData.Sender = "anonymous"

		// Convert back to JSON for consistent format
		message, _ = json.Marshal(msgData)
	}

	// Send the message to all clients
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			delete(clients, client)
			err := client.Close()
			if err != nil {
				return errors.New("failed to close client connection")
			}
		}
	}

	// Save the message to the database
	err = database.SaveMessage(chatID, msgData.Content, msgData.Sender)
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

	// Fetch existing messages for this chat and send them to the client
	messages, err := database.GetMessages(chatID)
	if err != nil {
		log.Printf("Error fetching messages: %v", err)
	} else {
		for _, msg := range messages {
			// Create a JSON message with sender and content
			jsonMsg, err := json.Marshal(map[string]string{
				"sender":  msg.Sender,
				"content": msg.Content,
			})
			if err != nil {
				log.Printf("Error marshaling message: %v", err)
				continue
			}

			err = conn.WriteMessage(websocket.TextMessage, jsonMsg)
			if err != nil {
				log.Printf("Error sending existing message: %v", err)
				break
			}
		}
	}

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
