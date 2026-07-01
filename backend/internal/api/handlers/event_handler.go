package handlers

import (
	"net/http"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/services"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/utils"
)

type EventHandler struct {
	eventService services.EventService
}

func NewEventHandler(es services.EventService) *EventHandler {
	return &EventHandler{eventService: es}
}

func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid group ID format", nil)
		return
	}

	var req dto.CreateEventRequest
	if err := utils.DecodeJSON(r, &req); err != nil || req.Title == "" || req.EventDate == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid request. title and event_date are required.", nil)
		return
	}

	eventDate, err := time.Parse(time.RFC3339, req.EventDate)
	if err != nil {
		eventDate, err = time.Parse("2006-01-02 15:04:05", req.EventDate)
		if err != nil {
			_ = utils.SendError(w, http.StatusBadRequest, "Invalid eventDate format. Use RFC3339 or 'YYYY-MM-DD HH:MM:SS'.", nil)
			return
		}
	}

	e, err := h.eventService.CreateEvent(currentUser.ID, groupID, req.Title, req.Description, eventDate)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusCreated, "Event created successfully", e)
}

func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	groupID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid group ID format", nil)
		return
	}

	events, err := h.eventService.ListGroupEvents(groupID, currentUser.ID)
	if err != nil {
		_ = utils.SendError(w, http.StatusForbidden, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "Events returned.", events)
}

func (h *EventHandler) RespondEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		_ = utils.SendError(w, http.StatusMethodNotAllowed, "Method not allowed", nil)
		return
	}

	currentUser, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		_ = utils.SendError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	eventID, err := uuid.FromString(r.PathValue("id"))
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid event ID format", nil)
		return
	}

	var req dto.EventRSVPRequest
	if err := utils.DecodeJSON(r, &req); err != nil || req.Status == "" {
		_ = utils.SendError(w, http.StatusBadRequest, "Invalid request. status is required.", nil)
		return
	}

	err = h.eventService.RespondToEvent(eventID, currentUser.ID, req.Status)
	if err != nil {
		_ = utils.SendError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	_ = utils.SendSuccess(w, http.StatusOK, "RSVP saved successfully", nil)
}
