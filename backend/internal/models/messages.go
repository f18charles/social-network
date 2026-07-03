package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type DMThread struct {
	ID            uuid.UUID `db:"id"`
	User1ID       uuid.UUID `db:"user1_id"`
	User2ID       uuid.UUID `db:"user2_id"`
	LastMessageAt time.Time `db:"last_message_at"`
}

type Message struct {
	ID         uuid.UUID                 `db:"id"`
	SenderID   uuid.UUID                 `db:"sender_id"`
	DMThreadID *uuid.UUID                `db:"dm_thread_id"`
	GroupID    *uuid.UUID                `db:"group_id"`
	Content    string                    `db:"content"`
	CreatedAt  time.Time                 `db:"created_at"`
	Reactions  []*MessageReactionSummary `db:"-"`
}

// Conversation is a repository read model for a chat conversation summary.
type Conversation struct {
	ThreadID      *uuid.UUID
	GroupID       *uuid.UUID
	Type          string
	TargetName    string
	TargetAvatar  string
	LastMessage   string
	LastMessageAt time.Time
}
