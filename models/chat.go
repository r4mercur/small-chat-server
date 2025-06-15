package models

// Chat represents a chat message in the system
type Chat struct {
	ChatID  string `bson:"chat_id"`
	Message string `bson:"message"`
}
