package services

import (
	"errors"
	"testing"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

func TestFollowerServicePublicProfileAutoAcceptsFollow(t *testing.T) {
	users := newFakeUserRepository()
	followers := newFakeFollowersRepository()
	notifications := &fakeNotificationServiceForFollower{}
	service := NewFollowerService(followers, users, notifications)
	followerID, targetID := followerTestIDs()
	users.add(&models.User{ID: targetID, Email: "public@example.com", IsPublic: true})

	status, err := service.Follow(followerID, targetID)
	if err != nil {
		t.Fatalf("Follow returned error: %v", err)
	}
	if status != string(models.Accepted) {
		t.Fatalf("status = %q, want accepted", status)
	}
	if followers.status[followerKey{followerID: followerID, followeeID: targetID}] != models.Accepted {
		t.Fatal("expected accepted follower relationship to be persisted")
	}
}

func TestFollowerServicePrivateProfileCreatesPendingRequest(t *testing.T) {
	users := newFakeUserRepository()
	followers := newFakeFollowersRepository()
	notifications := &fakeNotificationServiceForFollower{}
	service := NewFollowerService(followers, users, notifications)
	followerID, targetID := followerTestIDs()
	users.add(&models.User{ID: targetID, Email: "private@example.com", IsPublic: false})

	status, err := service.Follow(followerID, targetID)
	if err != nil {
		t.Fatalf("Follow returned error: %v", err)
	}
	if status != string(models.Pending) {
		t.Fatalf("status = %q, want pending", status)
	}
}

func TestFollowerServiceRejectsSelfFollow(t *testing.T) {
	users := newFakeUserRepository()
	notifications := &fakeNotificationServiceForFollower{}
	service := NewFollowerService(newFakeFollowersRepository(), users, notifications)
	userID := uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
	users.add(&models.User{ID: userID, Email: "amina@example.com", IsPublic: true})

	if _, err := service.Follow(userID, userID); err == nil {
		t.Fatal("expected self-follow to be rejected")
	}
}

func TestFollowerServiceAcceptAndRejectRequirePendingRequests(t *testing.T) {
	users := newFakeUserRepository()
	followers := newFakeFollowersRepository()
	notifications := &fakeNotificationServiceForFollower{}
	service := NewFollowerService(followers, users, notifications)
	followerID, targetID := followerTestIDs()
	users.add(&models.User{ID: targetID, Email: "private@example.com", IsPublic: false})

	if err := service.AcceptFollow(followerID, targetID); err == nil {
		t.Fatal("expected accepting a missing request to fail")
	}

	status, err := service.Follow(followerID, targetID)
	if err != nil {
		t.Fatalf("Follow returned error: %v", err)
	}
	if status != string(models.Pending) {
		t.Fatalf("status = %q, want pending", status)
	}
	if err := service.AcceptFollow(followerID, targetID); err != nil {
		t.Fatalf("AcceptFollow returned error: %v", err)
	}
	if followers.status[followerKey{followerID: followerID, followeeID: targetID}] != models.Accepted {
		t.Fatal("expected request to become accepted")
	}
	if err := service.RejectFollow(followerID, targetID); err == nil {
		t.Fatal("expected rejecting an accepted relationship to fail")
	}
}

func TestFollowerServiceRejectFollowRemovesPendingRequest(t *testing.T) {
	users := newFakeUserRepository()
	followers := newFakeFollowersRepository()
	notifications := &fakeNotificationServiceForFollower{}
	service := NewFollowerService(followers, users, notifications)
	followerID, targetID := followerTestIDs()
	users.add(&models.User{ID: targetID, Email: "private@example.com", IsPublic: false})

	if _, err := service.Follow(followerID, targetID); err != nil {
		t.Fatalf("Follow returned error: %v", err)
	}
	if err := service.RejectFollow(followerID, targetID); err != nil {
		t.Fatalf("RejectFollow returned error: %v", err)
	}
	if followers.status[followerKey{followerID: followerID, followeeID: targetID}] != "" {
		t.Fatal("expected pending request to be removed")
	}
}

func followerTestIDs() (uuid.UUID, uuid.UUID) {
	return uuid.Must(uuid.FromString("0dd6e443-0998-4f50-a4cf-1a40a0536213")),
		uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1"))
}

type followerKey struct {
	followerID uuid.UUID
	followeeID uuid.UUID
}

type fakeFollowersRepository struct {
	status map[followerKey]models.Status
}

func newFakeFollowersRepository() *fakeFollowersRepository {
	return &fakeFollowersRepository{status: map[followerKey]models.Status{}}
}

func (r *fakeFollowersRepository) Follow(followerID, followeeID uuid.UUID, status models.Status) error {
	key := followerKey{followerID: followerID, followeeID: followeeID}
	if _, exists := r.status[key]; exists {
		return errors.New("relationship already exists")
	}
	r.status[key] = status
	return nil
}

func (r *fakeFollowersRepository) Unfollow(followerID, followeeID uuid.UUID) error {
	key := followerKey{followerID: followerID, followeeID: followeeID}
	if _, exists := r.status[key]; !exists {
		return errors.New("relationship not found")
	}
	delete(r.status, key)
	return nil
}

func (r *fakeFollowersRepository) AcceptFollower(followerID, followeeID uuid.UUID) error {
	key := followerKey{followerID: followerID, followeeID: followeeID}
	if r.status[key] != models.Pending {
		return errors.New("follow request not found")
	}
	r.status[key] = models.Accepted
	return nil
}

func (r *fakeFollowersRepository) RejectFollower(followerID, followeeID uuid.UUID) error {
	key := followerKey{followerID: followerID, followeeID: followeeID}
	if r.status[key] != models.Pending {
		return errors.New("follow request not found")
	}
	delete(r.status, key)
	return nil
}

func (r *fakeFollowersRepository) GetFollowers(userID uuid.UUID) ([]*models.User, error) {
	return []*models.User{}, nil
}

func (r *fakeFollowersRepository) GetFollowing(userID uuid.UUID) ([]*models.User, error) {
	return []*models.User{}, nil
}

func (r *fakeFollowersRepository) GetPendingFollowers(userID uuid.UUID) ([]*models.User, error) {
	return []*models.User{}, nil
}

func (r *fakeFollowersRepository) GetStatus(followerID, followeeID uuid.UUID) (models.Status, error) {
	status, exists := r.status[followerKey{followerID: followerID, followeeID: followeeID}]
	if !exists {
		return "none", nil
	}
	return status, nil
}

type fakeNotificationServiceForFollower struct {
	createdCount int
	lastUserID   uuid.UUID
	lastType     string
	lastSourceID uuid.UUID
}

func (s *fakeNotificationServiceForFollower) CreateNotification(userID uuid.UUID, nType string, sourceID uuid.UUID) error {
	s.createdCount++
	s.lastUserID = userID
	s.lastType = nType
	s.lastSourceID = sourceID
	return nil
}

func (s *fakeNotificationServiceForFollower) GetNotifications(userID uuid.UUID) ([]*dto.NotificationResponse, error) {
	return nil, nil
}

func (s *fakeNotificationServiceForFollower) MarkAsRead(id, userID uuid.UUID) error { return nil }
func (s *fakeNotificationServiceForFollower) MarkAllAsRead(userID uuid.UUID) error  { return nil }
func (s *fakeNotificationServiceForFollower) RegisterPushHandler(handler func(userID uuid.UUID, payload any)) {
}
