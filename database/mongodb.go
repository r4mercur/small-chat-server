package database

import (
	"context"
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
func SaveMessage(chatID string, message string) error {
	chat := models.Chat{
		ChatID:  chatID,
		Message: message,
	}

	_, err := Collection.InsertOne(context.TODO(), chat)
	if err != nil {
		log.Printf("Failed to insert chat message into MongoDB: %v", err)
		return err
	}
	return nil
}
