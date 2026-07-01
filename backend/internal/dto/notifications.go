package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

// NotificationResponse is the API representation of a notification.
type NotificationResponse struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"type"`
	SourceID  uuid.UUID `json:"source_id"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	Message   string    `json:"message"`
}

// MapNotificationResponse maps a notification domain model plus display message to an API DTO.
func MapNotificationResponse(notification *models.Notification, message string) *NotificationResponse {
	if notification == nil {
		return nil
	}
	return &NotificationResponse{
		ID:        notification.ID,
		Type:      notification.Type,
		SourceID:  notification.SourceID,
		IsRead:    notification.IsRead,
		CreatedAt: notification.CreatedAt,
		Message:   message,
	}
}
