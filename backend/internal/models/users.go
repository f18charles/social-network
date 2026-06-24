package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type User struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Email          string    `db:"email" json:"email"`
	PassHash       string    `db:"password_hash" json:"pass_hash"`
	FirstName      string    `db:"first_name" json:"first_name"`
	LastName       string    `db:"last_name" json:"last_name"`
	DOB            time.Time `db:"dob" json:"dob"`
	Avatar         string    `db:"avatar" json:"avatar"`
	Nickname       string    `db:"nickname" json:"nickname"`
	AboutMe        string    `db:"about_me" json:"about_me"`
	IsPublic       bool      `db:"is_public" json:"is_public"`
	FollowerCount  int       `db:"follower_count" json:"follower_count"`
	FollowingCount int       `db:"following_count" json:"following_count"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type Session struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type Status string

const (
	Pending  Status = "pending"
	Accepted Status = "accepted"
)

type Follower struct {
	FollowerID uuid.UUID `db:"follower_id" json:"follower_id"`
	FolloweeID uuid.UUID `db:"followee_id" json:"followee_id"`
	Status     Status    `db:"status" json:"status"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

// Request and Response DTOs for the handlers

type CreateUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"` // YYYY-MM-DD
	Avatar      string `json:"avatar,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	AboutMe     string `json:"about_me,omitempty"`
	IsPublic    bool   `json:"is_public"`
}

type UserResponse struct {
	ID                   uuid.UUID `json:"id"`
	Email                string    `json:"email"`
	FirstName            string    `json:"first_name"`
	LastName             string    `json:"last_name"`
	DateOfBirth          string    `json:"date_of_birth"`
	Avatar               string    `json:"avatar,omitempty"`
	Nickname             string    `json:"nickname,omitempty"`
	AboutMe              string    `json:"about_me,omitempty"`
	IsPublic             bool      `json:"is_public"`
	CreatedAt            time.Time `json:"created_at"`
	IsFollowing          bool      `json:"is_following,omitempty"`
	FollowRequestPending bool      `json:"follow_request_pending,omitempty"`
}

type UpdateUserRequest struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	DateOfBirth     string `json:"date_of_birth"`
	Avatar          string `json:"avatar"`
	Nickname        string `json:"nickname"`
	AboutMe         string `json:"about_me"`
	IsPublic        bool   `json:"is_public"`
}

type FollowRequestInput struct {
	FollowingID string `json:"following_id"`
}

type AcceptRejectFollowInput struct {
	FollowerID string `json:"follower_id"`
}

type FollowStatusResponse struct {
	FollowerID  string `json:"follower_id"`
	FollowingID string `json:"following_id"`
	Status      string `json:"status"` // 'pending', 'accepted', 'none'
}
