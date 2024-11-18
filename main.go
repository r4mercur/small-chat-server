package main

import (
	"context"
	"errors"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"sync"
	"time"
)

type Chat struct {
	ChatID  string `bson:"chat_id"`
	Message string `bson:"message"`
}

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
	client       *mongo.Client
	collection   *mongo.Collection
)

func init() {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err = mongo.Connect(ctx, options.Client().ApplyURI("mongodb://admin:Lw4EGf67a6WJ@localhost:27017"))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB: ", err)
	}
	collection = client.Database("chatDatabase").Collection("chats")
}

func broadcastMessage(chatID string, message []byte) error {
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

	chat := Chat{ChatID: chatID, Message: string(message)}
	_, err := collection.InsertOne(context.TODO(), chat)
	if err != nil {
		log.Printf("Failed to insert chat message into MongoDB: %v", err)
		return err
	}
	return nil
}

func chatHandler(w http.ResponseWriter, r *http.Request) {
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		http.Error(w, "chat_id is required", http.StatusBadRequest)
		return
	}

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

		err2 := broadcastMessage(chatID, msg)
		if err2 != nil {
			log.Fatal(err2)
		}
	}
}

func main() {
	http.HandleFunc("/chat", chatHandler)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
