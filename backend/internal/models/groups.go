package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type Group struct {
	ID          uuid.UUID `db:"id"`
	CreatorID   uuid.UUID `db:"creator_id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Avatar      string    `db:"avatar"`
	CreatedAt   time.Time `db:"created_at"`
}

type GroupMember struct {
	GroupID uuid.UUID `db:"group_id"`
	UserID  uuid.UUID `db:"user_id"`
	Status  string    `db:"status"` // 'pending_invite', 'pending_request', 'accepted'
	Role    string    `db:"role"`   // 'admin', 'member'
}

// GroupMemberUser is a user row annotated with group membership state.
type GroupMemberUser struct {
	User
	Status string
	Role   string
}
