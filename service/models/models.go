package models

import "time"

// User represents a user in the system
type User struct {
	Name  string `json:"name"`
	Token string `json:"identifier"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversationId"`
	SenderID       string    `json:"senderId"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"createdAt"`
	Deleted        bool      `json:"deleted,omitempty"`
	Comment        string    `json:"comment,omitempty"`
	CommentedAt    time.Time `json:"commentedAt,omitempty"`
	ForwardedFrom  string    `json:"forwardedFrom,omitempty"`
}

// Conversation represents a conversation between users
type Conversation struct {
	ID           string    `json:"conversationId"`
	Participants []string  `json:"participants"`
	Messages     []Message `json:"messages"`
	IsGroup      bool      `json:"isGroup,omitempty"`
	Name         string    `json:"name,omitempty"`
}

// Participant represents a user participating in a conversation
type Participant struct {
	ConversationID string `json:"conversationId"`
	Username       string `json:"username"`
}
