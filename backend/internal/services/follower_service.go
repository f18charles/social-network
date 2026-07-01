package services

import (
	"errors"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

type FollowerService interface {
	Follow(followerID, followingID uuid.UUID) (string, error)
	Unfollow(followerID, followingID uuid.UUID) error
	AcceptFollow(followerID, followingID uuid.UUID) error
	RejectFollow(followerID, followingID uuid.UUID) error
	GetFollowers(userID uuid.UUID) ([]*models.User, error)
	GetFollowing(userID uuid.UUID) ([]*models.User, error)
	GetPendingFollowers(userID uuid.UUID) ([]*models.User, error)
	GetFollowStatus(followerID, followingID uuid.UUID) (string, error)
}

type followerService struct {
	followerRepo     repositories.FollowersRepository
	userRepo         repositories.UserRepository
	notificationServ NotificationService
}

func NewFollowerService(fr repositories.FollowersRepository, ur repositories.UserRepository, ns NotificationService) FollowerService {
	return &followerService{
		followerRepo:     fr,
		userRepo:         ur,
		notificationServ: ns,
	}
}

func (s *followerService) Follow(followerID, followingID uuid.UUID) (string, error) {
	if followerID == followingID {
		return "", errors.New("cannot follow yourself")
	}

	// Verify target user exists
	targetUser, err := s.userRepo.GetUserByID(followingID)
	if err != nil {
		return "", errors.New("target user not found")
	}

	// Check current status
	status, err := s.followerRepo.GetStatus(followerID, followingID)
	if err != nil {
		return "", err
	}

	if status != "none" {
		return string(status), errors.New("follow relationship or request already exists")
	}

	// If target user is public, follow directly. If private, send follow request (pending status).
	newStatus := models.Pending
	if targetUser.IsPublic {
		newStatus = models.Accepted
	}

	err = s.followerRepo.Follow(followerID, followingID, newStatus)
	if err != nil {
		return "", err
	}

	if newStatus == models.Pending {
		_ = s.notificationServ.CreateNotification(followingID, "follow_request", followerID, nil)
	}

	return string(newStatus), nil
}

func (s *followerService) Unfollow(followerID, followingID uuid.UUID) error {
	return s.followerRepo.Unfollow(followerID, followingID)
}

func (s *followerService) AcceptFollow(followerID, followingID uuid.UUID) error {
	status, err := s.followerRepo.GetStatus(followerID, followingID)
	if err != nil {
		return err
	}
	if status != models.Pending {
		return errors.New("no pending follow request to accept")
	}
	return s.followerRepo.AcceptFollower(followerID, followingID)
}

func (s *followerService) RejectFollow(followerID, followingID uuid.UUID) error {
	status, err := s.followerRepo.GetStatus(followerID, followingID)
	if err != nil {
		return err
	}
	if status != models.Pending {
		return errors.New("no pending follow request to reject")
	}
	return s.followerRepo.RejectFollower(followerID, followingID)
}

func (s *followerService) GetFollowers(userID uuid.UUID) ([]*models.User, error) {
	return s.followerRepo.GetFollowers(userID)
}

func (s *followerService) GetFollowing(userID uuid.UUID) ([]*models.User, error) {
	return s.followerRepo.GetFollowing(userID)
}

func (s *followerService) GetPendingFollowers(userID uuid.UUID) ([]*models.User, error) {
	return s.followerRepo.GetPendingFollowers(userID)
}

func (s *followerService) GetFollowStatus(followerID, followingID uuid.UUID) (string, error) {
	status, err := s.followerRepo.GetStatus(followerID, followingID)
	if err != nil {
		return "", err
	}
	return string(status), nil
}
