package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type MessageReaction struct {
	MessageID uuid.UUID  `db:"message_id"`
	UserID    uuid.UUID  `db:"user_id"`
	Emoji     string     `db:"emoji"`
	CreatedAt time.Time  `db:"created_at"`
}

type MessageReactionSummary struct {
	Emoji       string `json:"emoji"`
	Count       int    `json:"count"`
	UserReacted bool   `json:"user_reacted"`
}
