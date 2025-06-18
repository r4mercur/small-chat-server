package database

import (
	"context"
	"errors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"

	"chat-server/models"
)

var (
	Client     *mongo.Client
	Collection *mongo.Collection
)

// InitMongoDB initializes the MongoDB connection
func InitMongoDB() {
	var err error
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoUser := os.Getenv("MONGO_USER")
	mongoPass := os.Getenv("MONGO_PASS")
	mongoHost := os.Getenv("MONGO_HOST")

	mongoUri := "mongodb://" + mongoUser + ":" + mongoPass + "@" + mongoHost

	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))
	if err != nil {
		log.Fatal("Failed to connect to MongoDB: ", err)
	}
	Collection = Client.Database("chatDatabase").Collection("chats")

	// Check if the connection is successful by inserting a test document
	InitCollection(err, ctx)
}

func InitCollection(err error, ctx context.Context) {
	_, err = Collection.InsertOne(ctx, map[string]interface{}{"init": true})
	if err != nil {
		log.Fatal("Error when creating the collection: ", err)
	}
	_, _ = Collection.DeleteOne(ctx, map[string]interface{}{"init": true})
}

// SaveMessage saves a chat message to MongoDB
func SaveMessage(chatID string, messageContent string, sender string) error {
	ctx := context.TODO()

	// Create a new message
	message := models.Message{
		Sender:    sender,
		Content:   messageContent,
		Timestamp: time.Now().Unix(),
	}

	// Check if chat exists
	filter := map[string]interface{}{
		"chat_id": chatID,
	}

	// Update operation to add the message to the messages array
	update := map[string]interface{}{
		"$push": map[string]interface{}{
			"messages": message,
		},
	}

	// Use upsert to create the chat if it doesn't exist
	opts := options.Update().SetUpsert(true)

	_, err := Collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Failed to save chat message to MongoDB: %v", err)
		return err
	}

	return nil
}

// GetMessages retrieves all messages for a specific chat ID
func GetMessages(chatID string) ([]models.Message, error) {
	ctx := context.TODO()

	filter := map[string]interface{}{
		"chat_id": chatID,
	}

	// Find the chat document
	var chat models.Chat
	err := Collection.FindOne(ctx, filter).Decode(&chat)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No messages yet, return empty slice
			return []models.Message{}, nil
		}
		log.Printf("Failed to retrieve chat from MongoDB: %v", err)
		return nil, err
	}

	return chat.Messages, nil
}

// AddReaction adds an emoji reaction to a specific message
func AddReaction(chatID string, messageID string, emoji string, username string) error {
	ctx := context.TODO()

	// Find the chat document
	filter := map[string]interface{}{
		"chat_id": chatID,
	}

	// Get the current chat document
	var chat models.Chat
	err := Collection.FindOne(ctx, filter).Decode(&chat)
	if err != nil {
		log.Printf("Failed to find chat for reaction: %v", err)
		return err
	}

	// Find the message by its ID (we'll use the timestamp as the ID for simplicity)
	messageIndex := -1
	for i, msg := range chat.Messages {
		if messageID == msg.Sender+"-"+msg.Content {
			messageIndex = i
			break
		}
	}

	if messageIndex == -1 {
		return errors.New("message not found")
	}

	// Initialize the reactions map if it doesn't exist
	if chat.Messages[messageIndex].Reactions == nil {
		chat.Messages[messageIndex].Reactions = make(map[string][]string)
	}

	// Check if user already reacted with this emoji
	userReacted := false
	for _, user := range chat.Messages[messageIndex].Reactions[emoji] {
		if user == username {
			userReacted = true
			break
		}
	}

	// If user hasn't reacted with this emoji yet, add the reaction
	if !userReacted {
		chat.Messages[messageIndex].Reactions[emoji] = append(
			chat.Messages[messageIndex].Reactions[emoji],
			username,
		)
	}

	// Update the chat document in the database
	update := map[string]interface{}{
		"$set": map[string]interface{}{
			"messages": chat.Messages,
		},
	}

	_, err = Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Printf("Failed to update message with reaction: %v", err)
		return err
	}

	return nil
}
