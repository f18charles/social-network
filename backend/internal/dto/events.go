package dto

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

// CreateEventRequest is the API payload for creating a group event.
type CreateEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	EventDate   string `json:"event_date"`
}

// EventRSVPRequest is the API payload for responding to an event.
type EventRSVPRequest struct {
	Status string `json:"status"`
}

// EventResponse is the API representation of a group event with viewer RSVP state.
type EventResponse struct {
	ID            uuid.UUID `json:"id"`
	GroupID       uuid.UUID `json:"group_id"`
	CreatorID     uuid.UUID `json:"creator_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	EventDate     time.Time `json:"event_date"`
	CreatedAt     time.Time `json:"created_at"`
	UserRSVP      string    `json:"user_rsvp"`
	GoingCount    int       `json:"going_count"`
	NotGoingCount int       `json:"not_going_count"`
}

// MapEventResponse maps an event domain model plus RSVP summary to an API DTO.
func MapEventResponse(event *models.Event, userRSVP string, goingCount, notGoingCount int) *EventResponse {
	if event == nil {
		return nil
	}
	return &EventResponse{
		ID:            event.ID,
		GroupID:       event.GroupID,
		CreatorID:     event.CreatorID,
		Title:         event.Title,
		Description:   event.Description,
		EventDate:     event.EventDate,
		CreatedAt:     event.CreatedAt,
		UserRSVP:      userRSVP,
		GoingCount:    goingCount,
		NotGoingCount: notGoingCount,
	}
}
