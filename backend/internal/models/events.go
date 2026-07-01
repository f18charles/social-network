package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Event struct {
	ID          uuid.UUID `db:"id"`
	GroupID     uuid.UUID `db:"group_id"`
	CreatorID   uuid.UUID `db:"creator_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	EventDate   time.Time `db:"event_date"`
	CreatedAt   time.Time `db:"created_at"`
}

type EventRSVP struct {
	EventID uuid.UUID `db:"event_id"`
	UserID  uuid.UUID `db:"user_id"`
	Status  string    `db:"status"` // 'going', 'not_going', 'pending_invite'
}
