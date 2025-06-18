package models

// Message represents a single message in a chat
type Message struct {
	Sender    string              `bson:"sender"`
	Content   string              `bson:"content"`
	Timestamp int64               `bson:"timestamp,omitempty"`
	Reactions map[string][]string `bson:"reactions,omitempty"` // Map of emoji to list of users who reacted
}

// Chat represents a chat conversation in the system
type Chat struct {
	ChatID   string    `bson:"chat_id"`
	Messages []Message `bson:"messages"`
}
