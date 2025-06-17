package main

import (
	"log"
	"net/http"

	"chat-server/database"
	"chat-server/websocket"
)

func main() {
	// Initialize MongoDB connection
	database.InitMongoDB()

	// Health check endpoint for the database connection
	http.HandleFunc("/health", websocket.HealthCheck)

	// Set up HTTP routes
	http.HandleFunc("/chat", websocket.HandleConnection)

	// Start the server
	log.Println("Server starting on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
