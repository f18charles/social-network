package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Event struct {
	ID          uuid.UUID `db:"id" json:"id"`
	GroupID     uuid.UUID `db:"group_id" json:"group_id"`
	CreatorID   uuid.UUID `db:"creator_id" json:"creator_id"`
	Title       string    `db:"title" json:"title"`
	Description string    `db:"description" json:"description"`
	EventDate   time.Time `db:"event_date" json:"event_date"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type EventRSVP struct {
	EventID uuid.UUID `db:"event_id" json:"event_id"`
	UserID  uuid.UUID `db:"user_id" json:"user_id"`
	Status  string    `db:"status" json:"status"` // 'going', 'not_going', 'pending_invite'
}

type CreateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"` // RFC3339 format
}

type EventRSVPRequest struct {
	Status string `json:"status"` // 'going', 'not_going'
}

type EventResponse struct {
	ID            uuid.UUID `json:"id"`
	GroupID       uuid.UUID `json:"group_id"`
	CreatorID     uuid.UUID `json:"creator_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	EventDate     time.Time `json:"event_date"`
	CreatedAt     time.Time `json:"created_at"`
	UserRSVP      string    `json:"user_rsvp"` // 'going', 'not_going', 'pending_invite', 'none'
	GoingCount    int       `json:"going_count"`
	NotGoingCount int       `json:"not_going_count"`
}
