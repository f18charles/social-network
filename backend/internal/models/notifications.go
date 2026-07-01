package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Notification struct {
	ID        uuid.UUID  `db:"id" json:"id"`
	UserID    uuid.UUID  `db:"user_id" json:"user_id"`
	Type      string     `db:"type" json:"type"` // 'follow_request', 'group_invite', 'event_invite', 'group_request', 'event_created'
	SourceID  uuid.UUID  `db:"source_id" json:"source_id"`
	GroupID   *uuid.UUID `db:"group_id" json:"group_id,omitempty"` // only set for 'group_request', where source_id is the requester, not the group
	IsRead    bool       `db:"is_read" json:"is_read"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
}

type NotificationResponse struct {
	ID        uuid.UUID  `json:"id"`
	Type      string     `json:"type"`
	SourceID  uuid.UUID  `json:"source_id"`
	GroupID   *uuid.UUID `json:"group_id,omitempty"`
	IsRead    bool       `json:"is_read"`
	CreatedAt time.Time  `json:"created_at"`
	Message   string     `json:"message"`
}
