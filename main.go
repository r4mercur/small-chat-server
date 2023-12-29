package main

import (
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

var (
	upgrade = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	clients      = make(map[*websocket.Conn]bool)
	clientsMutex = sync.Mutex{}
)

func broadcastMessage(message []byte) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			delete(clients, client)
			err := client.Close()
			if err != nil {
				return
			}
		}
	}
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	defer func(conn *websocket.Conn) {
		err := conn.Close()
		if err != nil {

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
		broadcastMessage(msg)
	}
}

func main() {
	http.HandleFunc("/chat", chatHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
