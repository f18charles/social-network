package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type DMThread struct {
	ID            uuid.UUID `db:"id" json:"id"`
	User1ID       uuid.UUID `db:"user1_id" json:"user1_id"`
	User2ID       uuid.UUID `db:"user2_id" json:"user2_id"`
	LastMessageAt time.Time `db:"last_message_at" json:"last_message_at"`
}

type Message struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	SenderID   uuid.UUID  `db:"sender_id" json:"sender_id"`
	DMThreadID *uuid.UUID `db:"dm_thread_id" json:"dm_thread_id,omitempty"`
	GroupID    *uuid.UUID `db:"group_id" json:"group_id,omitempty"`
	Content    string     `db:"content" json:"content"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

type SendMessageRequest struct {
	Content     string  `json:"content"`
	DMThreadID  *string `json:"dm_thread_id,omitempty"`
	RecipientID *string `json:"recipient_id,omitempty"` // For starting/replying to a direct message without thread ID
	GroupID     *string `json:"group_id,omitempty"`     // For group chat
}

type ConversationResponse struct {
	ThreadID      *uuid.UUID `json:"thread_id,omitempty"` // For DMs
	GroupID       *uuid.UUID `json:"group_id,omitempty"`  // For group chats
	Type          string     `json:"type"`                // 'dm' or 'group'
	TargetName    string     `json:"target_name"`         // Group name or User nickname/name
	TargetAvatar  string     `json:"target_avatar,omitempty"`
	LastMessage   string     `json:"last_message"`
	LastMessageAt time.Time  `json:"last_message_at"`
}

// WSMessage represents a message wrapper sent over WebSocket.
type WSMessage struct {
	Type    string `json:"type"` // 'chat', 'notification'
	Payload any    `json:"payload"`
}
