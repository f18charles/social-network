package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/api/middleware"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

func TestGroupHandlerRespondMembershipDefaultsTargetToCurrentUser(t *testing.T) {
	currentUserID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))
	service := &handlerTestGroupService{}
	handler := NewGroupHandler(service)
	request := authenticatedRequestWithBody(http.MethodPost, "/api/groups/"+groupID.String()+"/respond", currentUserID, `{"action":"accept"}`)
	request.SetPathValue("id", groupID.String())
	recorder := httptest.NewRecorder()

	handler.RespondMembership(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.respondGroupID != groupID || service.respondUserID != currentUserID || service.respondDeciderID != currentUserID || service.respondAction != "accept" {
		t.Fatalf("service call = group %s user %s decider %s action %q", service.respondGroupID, service.respondUserID, service.respondDeciderID, service.respondAction)
	}
}

func TestGroupHandlerInviteUserRejectsMalformedInviteeID(t *testing.T) {
	currentUserID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))
	handler := NewGroupHandler(&handlerTestGroupService{})
	request := authenticatedRequestWithBody(http.MethodPost, "/api/groups/"+groupID.String()+"/invite", currentUserID, `{"user_id":"bad"}`)
	request.SetPathValue("id", groupID.String())
	recorder := httptest.NewRecorder()

	handler.InviteUser(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestEventHandlerCreateEventParsesRFC3339AndRejectsInvalidDate(t *testing.T) {
	currentUserID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))
	eventDate := "2026-07-12T15:30:00Z"
	service := &handlerTestEventService{createResponse: &dto.EventResponse{ID: uuid.Must(uuid.NewV4())}}
	handler := NewEventHandler(service)
	request := authenticatedRequestWithBody(http.MethodPost, "/api/groups/"+groupID.String()+"/events", currentUserID, `{"title":"Planning","event_date":"`+eventDate+`"}`)
	request.SetPathValue("id", groupID.String())
	recorder := httptest.NewRecorder()

	handler.CreateEvent(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusCreated, recorder.Body.String())
	}
	if service.createCreatorID != currentUserID || service.createGroupID != groupID || service.createTitle != "Planning" {
		t.Fatalf("service create call = creator %s group %s title %q", service.createCreatorID, service.createGroupID, service.createTitle)
	}
	if service.createDate.Format(time.RFC3339) != eventDate {
		t.Fatalf("event date = %s, want %s", service.createDate.Format(time.RFC3339), eventDate)
	}

	badRequest := authenticatedRequestWithBody(http.MethodPost, "/api/groups/"+groupID.String()+"/events", currentUserID, `{"title":"Planning","event_date":"not-a-date"}`)
	badRequest.SetPathValue("id", groupID.String())
	badRecorder := httptest.NewRecorder()
	handler.CreateEvent(badRecorder, badRequest)
	if badRecorder.Code != http.StatusBadRequest {
		t.Fatalf("bad date status = %d, want %d", badRecorder.Code, http.StatusBadRequest)
	}
}

func TestNotificationHandlerUsesAuthenticatedUserForReadOperations(t *testing.T) {
	currentUserID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	notificationID := uuid.Must(uuid.FromString("30000000-0000-0000-0000-000000000001"))
	service := &handlerTestNotificationService{}
	handler := NewNotificationHandler(service)
	request := authenticatedRequest(http.MethodPost, "/api/notifications/"+notificationID.String()+"/read", currentUserID)
	request.SetPathValue("id", notificationID.String())
	recorder := httptest.NewRecorder()

	handler.MarkAsRead(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if service.markID != notificationID || service.markUserID != currentUserID {
		t.Fatalf("MarkAsRead call = notification %s user %s", service.markID, service.markUserID)
	}

	allRequest := authenticatedRequest(http.MethodPost, "/api/notifications/read/all", currentUserID)
	allRecorder := httptest.NewRecorder()
	handler.MarkAllAsRead(allRecorder, allRequest)
	if allRecorder.Code != http.StatusOK {
		t.Fatalf("mark all status = %d, want %d", allRecorder.Code, http.StatusOK)
	}
	if service.markAllUserID != currentUserID {
		t.Fatalf("MarkAllAsRead user = %s, want %s", service.markAllUserID, currentUserID)
	}
}

func TestNotificationHandlerRejectsMalformedNotificationID(t *testing.T) {
	handler := NewNotificationHandler(&handlerTestNotificationService{})
	request := authenticatedRequest(http.MethodPost, "/api/notifications/bad/read", uuid.Must(uuid.NewV4()))
	request.SetPathValue("id", "bad")
	recorder := httptest.NewRecorder()

	handler.MarkAsRead(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func authenticatedRequestWithBody(method, target string, userID uuid.UUID, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	user := &models.User{ID: userID, Email: "viewer@example.com"}
	return request.WithContext(context.WithValue(request.Context(), middleware.UserContextKey, user))
}

type handlerTestGroupService struct {
	respondGroupID   uuid.UUID
	respondUserID    uuid.UUID
	respondDeciderID uuid.UUID
	respondAction    string
}

func (s *handlerTestGroupService) CreateGroup(creatorID uuid.UUID, title, description string) (*dto.GroupResponse, error) {
	return &dto.GroupResponse{CreatorID: creatorID, Title: title, Description: description}, nil
}
func (s *handlerTestGroupService) GetGroup(id uuid.UUID) (*models.Group, error) { return nil, nil }
func (s *handlerTestGroupService) ListGroups(viewerID uuid.UUID) ([]*dto.GroupResponse, error) {
	return nil, nil
}
func (s *handlerTestGroupService) RequestJoin(groupID, userID uuid.UUID) error { return nil }
func (s *handlerTestGroupService) InviteUser(groupID, inviterID, inviteeID uuid.UUID) error {
	return nil
}
func (s *handlerTestGroupService) RespondToMembership(groupID, userID, deciderID uuid.UUID, action string) error {
	s.respondGroupID = groupID
	s.respondUserID = userID
	s.respondDeciderID = deciderID
	s.respondAction = action
	return nil
}
func (s *handlerTestGroupService) ListMembers(groupID, viewerID uuid.UUID) ([]*dto.UserResponse, error) {
	return nil, nil
}
func (s *handlerTestGroupService) ListPendingRequests(groupID, creatorID uuid.UUID) ([]*dto.UserResponse, error) {
	return nil, nil
}

type handlerTestEventService struct {
	createCreatorID uuid.UUID
	createGroupID   uuid.UUID
	createTitle     string
	createDate      time.Time
	createResponse  *dto.EventResponse
}

func (s *handlerTestEventService) CreateEvent(creatorID, groupID uuid.UUID, title, description string, eventDate time.Time) (*dto.EventResponse, error) {
	s.createCreatorID = creatorID
	s.createGroupID = groupID
	s.createTitle = title
	s.createDate = eventDate
	return s.createResponse, nil
}
func (s *handlerTestEventService) GetEvent(id, userID uuid.UUID) (*dto.EventResponse, error) {
	return nil, nil
}
func (s *handlerTestEventService) ListGroupEvents(groupID, userID uuid.UUID) ([]*dto.EventResponse, error) {
	return nil, nil
}
func (s *handlerTestEventService) RespondToEvent(eventID, userID uuid.UUID, status string) error {
	if status == "error" {
		return errors.New("forced")
	}
	return nil
}

type handlerTestNotificationService struct {
	markID        uuid.UUID
	markUserID    uuid.UUID
	markAllUserID uuid.UUID
}

func (s *handlerTestNotificationService) CreateNotification(userID uuid.UUID, nType string, sourceID uuid.UUID) error {
	return nil
}
func (s *handlerTestNotificationService) GetNotifications(userID uuid.UUID) ([]*dto.NotificationResponse, error) {
	return []*dto.NotificationResponse{}, nil
}
func (s *handlerTestNotificationService) MarkAsRead(id, userID uuid.UUID) error {
	s.markID = id
	s.markUserID = userID
	return nil
}
func (s *handlerTestNotificationService) MarkAllAsRead(userID uuid.UUID) error {
	s.markAllUserID = userID
	return nil
}
func (s *handlerTestNotificationService) RegisterPushHandler(handler func(userID uuid.UUID, payload any)) {
}
