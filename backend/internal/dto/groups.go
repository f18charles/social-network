package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

// CreateGroupRequest is the API payload for creating a group.
type CreateGroupRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Avatar      string `json:"avatar,omitempty"`
}

// InviteUserRequest is the API payload for inviting a user to a group.
type InviteUserRequest struct {
	UserID string `json:"user_id"`
}

// RespondMembershipRequest is the API payload for accepting or rejecting group membership.
type RespondMembershipRequest struct {
	UserID string `json:"user_id"`
	Action string `json:"action"`
}

// GroupResponse is the API representation of a group with viewer membership state.
type GroupResponse struct {
	ID          uuid.UUID `json:"id"`
	CreatorID   uuid.UUID `json:"creator_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Avatar      string    `json:"avatar,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	IsMember    bool      `json:"is_member"`
	Status      string    `json:"status,omitempty"`
	Role        string    `json:"role,omitempty"`
}

// MapGroupResponse maps a group domain model plus viewer status to an API DTO.
func MapGroupResponse(group *models.Group, status string) *GroupResponse {
	if group == nil {
		return nil
	}
	return &GroupResponse{
		ID:          group.ID,
		CreatorID:   group.CreatorID,
		Title:       group.Title,
		Description: group.Description,
		Avatar:      group.Avatar,
		CreatedAt:   group.CreatedAt,
		IsMember:    status == "accepted",
		Status:      status,
	}
}
