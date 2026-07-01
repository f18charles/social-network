package dto

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

// CreateUserRequest is the API payload for registering a user.
type CreateUserRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Avatar      string `json:"avatar,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	AboutMe     string `json:"about_me,omitempty"`
	IsPublic    bool   `json:"is_public"`
}

// LoginRequest is the API payload for creating a session.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse is the API payload returned after successful login.
type LoginResponse struct {
	Token string `json:"token"`
}

// UserResponse is the safe API representation of a user.
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

// UpdateUserRequest is the API payload for updating the current user.
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

// FollowRequestInput is the API payload for following or unfollowing a user.
type FollowRequestInput struct {
	FollowingID string `json:"following_id"`
}

// AcceptRejectFollowInput is the API payload for deciding a follow request.
type AcceptRejectFollowInput struct {
	FollowerID string `json:"follower_id"`
}

// FollowStatusResponse is the API representation of a follower relationship.
type FollowStatusResponse struct {
	FollowerID  string `json:"follower_id"`
	FollowingID string `json:"following_id"`
	Status      string `json:"status"`
}

// PublicUserResponse is the safe user shape embedded by other DTOs.
type PublicUserResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Nickname  *string   `json:"nickname"`
	Avatar    *string   `json:"avatar"`
}

// MapUserResponse maps a user domain model to its safe API DTO.
func MapUserResponse(user *models.User) (*UserResponse, error) {
	if user == nil {
		return nil, errors.New("user is required")
	}
	return &UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DOB.Format("2006-01-02"),
		Avatar:      user.Avatar,
		Nickname:    user.Nickname,
		AboutMe:     user.AboutMe,
		IsPublic:    user.IsPublic,
		CreatedAt:   user.CreatedAt,
	}, nil
}

// MustMapUserResponse maps a user domain model where nil has already been excluded.
func MustMapUserResponse(user *models.User) *UserResponse {
	response, err := MapUserResponse(user)
	if err != nil {
		return nil
	}
	return response
}

// MapPublicUserResponse maps a public user read model to its safe embedded DTO.
func MapPublicUserResponse(user *models.PublicUser) (PublicUserResponse, error) {
	if user == nil {
		return PublicUserResponse{}, errors.New("public user is required")
	}
	return PublicUserResponse{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
	}, nil
}
