package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Notification struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	Type      string    `db:"type"` // 'follow_request', 'group_invite', 'event_invite', 'group_request', 'event_created'
	SourceID  uuid.UUID `db:"source_id"`
	IsRead    bool      `db:"is_read"`
	CreatedAt time.Time `db:"created_at"`
}
