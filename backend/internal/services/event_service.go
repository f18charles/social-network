package services

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

type EventService interface {
	CreateEvent(creatorID, groupID uuid.UUID, title, description string, eventDate time.Time) (*dto.EventResponse, error)
	GetEvent(id, userID uuid.UUID) (*dto.EventResponse, error)
	ListGroupEvents(groupID, userID uuid.UUID) ([]*dto.EventResponse, error)
	RespondToEvent(eventID, userID uuid.UUID, status string) error
}

type eventService struct {
	eventRepo        repositories.EventRepository
	membershipRepo   repositories.GroupMembershipRepository
	notificationServ NotificationService
}

func NewEventService(
	er repositories.EventRepository,
	mr repositories.GroupMembershipRepository,
	ns NotificationService,
) EventService {
	return &eventService{
		eventRepo:        er,
		membershipRepo:   mr,
		notificationServ: ns,
	}
}

func (s *eventService) CreateEvent(creatorID, groupID uuid.UUID, title, description string, eventDate time.Time) (*dto.EventResponse, error) {
	if title == "" {
		return nil, errors.New("event title is required")
	}

	// Verify creator is an accepted member of the group
	isMember, err := s.membershipRepo.IsAcceptedGroupMember(groupID, creatorID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized: must be group member to create events")
	}

	eventID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	e := &models.Event{
		ID:          eventID,
		GroupID:     groupID,
		CreatorID:   creatorID,
		Title:       title,
		Description: description,
		EventDate:   eventDate,
		CreatedAt:   time.Now(),
	}

	if err := s.eventRepo.CreateEvent(e); err != nil {
		return nil, err
	}

	// RSVP the creator as 'going' automatically
	_ = s.eventRepo.SetRSVP(eventID, creatorID, "going")

	// Notify all other group members
	members, err := s.membershipRepo.ListGroupMembers(groupID)
	if err == nil {
		for _, m := range members {
			if m.ID != creatorID {
				_ = s.notificationServ.CreateNotification(m.ID, "event_created", eventID)
			}
		}
	}

	return dto.MapEventResponse(e, "going", 1, 0), nil
}

func (s *eventService) GetEvent(id, userID uuid.UUID) (*dto.EventResponse, error) {
	e, err := s.eventRepo.GetEventByID(id)
	if err != nil {
		return nil, err
	}

	// Verify viewer is a group member
	isMember, err := s.membershipRepo.IsAcceptedGroupMember(e.GroupID, userID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized: must be group member to view events")
	}

	rsvp, err := s.eventRepo.GetRSVP(e.ID, userID)
	if err != nil {
		return nil, err
	}

	going, notGoing, err := s.eventRepo.GetRSVPSummaries(e.ID)
	if err != nil {
		return nil, err
	}

	return &dto.EventResponse{
		ID:            e.ID,
		GroupID:       e.GroupID,
		CreatorID:     e.CreatorID,
		Title:         e.Title,
		Description:   e.Description,
		EventDate:     e.EventDate,
		CreatedAt:     e.CreatedAt,
		UserRSVP:      rsvp,
		GoingCount:    going,
		NotGoingCount: notGoing,
	}, nil
}

func (s *eventService) ListGroupEvents(groupID, userID uuid.UUID) ([]*dto.EventResponse, error) {
	// Verify viewer is a group member
	isMember, err := s.membershipRepo.IsAcceptedGroupMember(groupID, userID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized: must be group member to view events")
	}

	events, err := s.eventRepo.ListEventsByGroup(groupID)
	if err != nil {
		return nil, err
	}

	var response []*dto.EventResponse
	for _, e := range events {
		rsvp, err := s.eventRepo.GetRSVP(e.ID, userID)
		if err != nil {
			return nil, err
		}

		going, notGoing, err := s.eventRepo.GetRSVPSummaries(e.ID)
		if err != nil {
			return nil, err
		}

		response = append(response, &dto.EventResponse{
			ID:            e.ID,
			GroupID:       e.GroupID,
			CreatorID:     e.CreatorID,
			Title:         e.Title,
			Description:   e.Description,
			EventDate:     e.EventDate,
			CreatedAt:     e.CreatedAt,
			UserRSVP:      rsvp,
			GoingCount:    going,
			NotGoingCount: notGoing,
		})
	}

	return response, nil
}

func (s *eventService) RespondToEvent(eventID, userID uuid.UUID, status string) error {
	if status != "going" && status != "not_going" {
		return errors.New("invalid RSVP status: must be 'going' or 'not_going'")
	}

	e, err := s.eventRepo.GetEventByID(eventID)
	if err != nil {
		return errors.New("event not found")
	}

	// Verify user is group member
	isMember, err := s.membershipRepo.IsAcceptedGroupMember(e.GroupID, userID)
	if err != nil || !isMember {
		return errors.New("unauthorized to RSVP to this event")
	}

	return s.eventRepo.SetRSVP(eventID, userID, status)
}
