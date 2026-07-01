package services

import (
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

type NotificationService interface {
	CreateNotification(userID uuid.UUID, nType string, sourceID uuid.UUID) error
	GetNotifications(userID uuid.UUID) ([]*dto.NotificationResponse, error)
	MarkAsRead(id, userID uuid.UUID) error
	MarkAllAsRead(userID uuid.UUID) error
	RegisterPushHandler(handler func(userID uuid.UUID, payload any))
}

type notificationService struct {
	notificationRepo repositories.NotificationRepository
	userRepo         repositories.UserRepository
	groupRepo        repositories.GroupRepository
	eventRepo        repositories.EventRepository
	pushHandler      func(userID uuid.UUID, payload any)
}

func NewNotificationService(
	nr repositories.NotificationRepository,
	ur repositories.UserRepository,
	gr repositories.GroupRepository,
	er repositories.EventRepository,
) NotificationService {
	return &notificationService{
		notificationRepo: nr,
		userRepo:         ur,
		groupRepo:        gr,
		eventRepo:        er,
	}
}

func (s *notificationService) RegisterPushHandler(handler func(userID uuid.UUID, payload any)) {
	s.pushHandler = handler
}

func (s *notificationService) CreateNotification(userID uuid.UUID, nType string, sourceID uuid.UUID) error {
	newID, err := uuid.NewV4()
	if err != nil {
		return err
	}

	n := &models.Notification{
		ID:        newID,
		UserID:    userID,
		Type:      nType,
		SourceID:  sourceID,
		IsRead:    false,
		CreatedAt: time.Now(),
	}

	if err := s.notificationRepo.CreateNotification(n); err != nil {
		return err
	}

	// Push real-time over WebSocket if handler is registered
	if s.pushHandler != nil {
		// Prepare a formatted notification response
		resp := s.formatNotification(n)
		s.pushHandler(userID, dto.WSMessage{
			Type:    "notification",
			Payload: resp,
		})
	}

	return nil
}

func (s *notificationService) GetNotifications(userID uuid.UUID) ([]*dto.NotificationResponse, error) {
	list, err := s.notificationRepo.ListNotificationsByUser(userID)
	if err != nil {
		return nil, err
	}

	var response []*dto.NotificationResponse
	for _, n := range list {
		response = append(response, s.formatNotification(n))
	}

	return response, nil
}

func (s *notificationService) MarkAsRead(id, userID uuid.UUID) error {
	n, err := s.notificationRepo.GetNotificationByID(id)
	if err != nil {
		return err
	}
	if n.UserID != userID {
		return fmt.Errorf("unauthorized to mark notification as read")
	}
	return s.notificationRepo.MarkAsRead(id)
}

func (s *notificationService) MarkAllAsRead(userID uuid.UUID) error {
	return s.notificationRepo.MarkAllAsRead(userID)
}

func (s *notificationService) formatNotification(n *models.Notification) *dto.NotificationResponse {
	resp := &dto.NotificationResponse{
		ID:        n.ID,
		Type:      n.Type,
		SourceID:  n.SourceID,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt,
		Message:   "New notification received",
	}

	switch n.Type {
	case "follow_request":
		u, err := s.userRepo.GetUserByID(n.SourceID)
		if err == nil {
			resp.Message = fmt.Sprintf("%s %s requested to follow you.", u.FirstName, u.LastName)
		} else {
			resp.Message = "Someone requested to follow you."
		}
	case "group_invite":
		g, err := s.groupRepo.GetGroupByID(n.SourceID)
		if err == nil {
			resp.Message = fmt.Sprintf("You were invited to join group '%s'.", g.Title)
		} else {
			resp.Message = "You were invited to join a group."
		}
	case "event_invite":
		e, err := s.eventRepo.GetEventByID(n.SourceID)
		if err == nil {
			resp.Message = fmt.Sprintf("You were invited to event '%s'.", e.Title)
		} else {
			resp.Message = "You were invited to an event."
		}
	case "group_request":
		u, err := s.userRepo.GetUserByID(n.SourceID)
		if err == nil {
			resp.Message = fmt.Sprintf("%s %s requested to join your group.", u.FirstName, u.LastName)
		} else {
			resp.Message = "A user requested to join your group."
		}
	case "event_created":
		e, err := s.eventRepo.GetEventByID(n.SourceID)
		if err == nil {
			resp.Message = fmt.Sprintf("A new event '%s' was created in your group.", e.Title)
		} else {
			resp.Message = "A new event was created in your group."
		}
	}

	return resp
}
