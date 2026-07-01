package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID             uuid.UUID `db:"id"`
	Email          string    `db:"email"`
	PassHash       string    `db:"password_hash"`
	FirstName      string    `db:"first_name"`
	LastName       string    `db:"last_name"`
	DOB            time.Time `db:"dob"`
	Avatar         string    `db:"avatar"`
	Nickname       string    `db:"nickname"`
	AboutMe        string    `db:"about_me"`
	IsPublic       bool      `db:"is_public"`
	FollowerCount  int       `db:"follower_count"`
	FollowingCount int       `db:"following_count"`
	CreatedAt      time.Time `db:"created_at"`
}

type Session struct {
	ID        uuid.UUID `db:"id"`
	UserID    uuid.UUID `db:"user_id"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}

type Status string

const (
	Pending  Status = "pending"
	Accepted Status = "accepted"
)

type Follower struct {
	FollowerID uuid.UUID `db:"follower_id"`
	FolloweeID uuid.UUID `db:"followee_id"`
	Status     Status    `db:"status"`
	CreatedAt  time.Time `db:"created_at"`
}
