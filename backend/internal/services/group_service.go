package services

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

type GroupService interface {
	CreateGroup(creatorID uuid.UUID, title, description string) (*dto.GroupResponse, error)
	GetGroup(id uuid.UUID) (*models.Group, error)
	ListGroups(viewerID uuid.UUID) ([]*dto.GroupResponse, error)
	RequestJoin(groupID, userID uuid.UUID) error
	InviteUser(groupID, inviterID, inviteeID uuid.UUID) error
	RespondToMembership(groupID, userID, deciderID uuid.UUID, action string) error
	ListMembers(groupID, viewerID uuid.UUID) ([]*dto.UserResponse, error)
	ListPendingRequests(groupID, creatorID uuid.UUID) ([]*dto.UserResponse, error)
}

type groupService struct {
	groupRepo        repositories.GroupRepository
	membershipRepo   repositories.GroupMembershipRepository
	userRepo         repositories.UserRepository
	notificationServ NotificationService
}

func NewGroupService(
	gr repositories.GroupRepository,
	mr repositories.GroupMembershipRepository,
	ur repositories.UserRepository,
	ns NotificationService,
) GroupService {
	return &groupService{
		groupRepo:        gr,
		membershipRepo:   mr,
		userRepo:         ur,
		notificationServ: ns,
	}
}

func (s *groupService) CreateGroup(creatorID uuid.UUID, title, description string) (*dto.GroupResponse, error) {
	if title == "" {
		return nil, errors.New("group title is required")
	}

	groupID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	g := &models.Group{
		ID:          groupID,
		CreatorID:   creatorID,
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
	}

	if err := s.groupRepo.CreateGroup(g); err != nil {
		return nil, err
	}

	// Add creator as accepted member
	if err := s.membershipRepo.AddMembership(groupID, creatorID, "accepted"); err != nil {
		return nil, err
	}

	return dto.MapGroupResponse(g, "accepted"), nil
}

func (s *groupService) GetGroup(id uuid.UUID) (*models.Group, error) {
	return s.groupRepo.GetGroupByID(id)
}

func (s *groupService) ListGroups(viewerID uuid.UUID) ([]*dto.GroupResponse, error) {
	list, err := s.groupRepo.ListGroups()
	if err != nil {
		return nil, err
	}

	var response []*dto.GroupResponse
	for _, g := range list {
		status, err := s.membershipRepo.GetMembership(g.ID, viewerID)
		if err != nil {
			return nil, err
		}

		response = append(response, &dto.GroupResponse{
			ID:          g.ID,
			CreatorID:   g.CreatorID,
			Title:       g.Title,
			Description: g.Description,
			CreatedAt:   g.CreatedAt,
			IsMember:    status == "accepted",
			Status:      status,
		})
	}

	return response, nil
}

func (s *groupService) RequestJoin(groupID, userID uuid.UUID) error {
	g, err := s.groupRepo.GetGroupByID(groupID)
	if err != nil {
		return errors.New("group not found")
	}

	status, err := s.membershipRepo.GetMembership(groupID, userID)
	if err != nil {
		return err
	}

	if status == "accepted" {
		return errors.New("membership relationship already exists")
	}

	if status == "pending_invite" {
		// Accept a pending invite when the invited user chooses to join.
		return s.membershipRepo.UpdateMembershipStatus(groupID, userID, "accepted")
	}

	if status != "none" {
		return errors.New("membership relationship already exists")
	}

	if err := s.membershipRepo.AddMembership(groupID, userID, "pending_request"); err != nil {
		return err
	}

	// Notify group creator
	_ = s.notificationServ.CreateNotification(g.CreatorID, "group_request", userID, &groupID)

	return nil
}

func (s *groupService) InviteUser(groupID, inviterID, inviteeID uuid.UUID) error {
	// Verify inviter is an accepted member
	isMember, err := s.membershipRepo.IsAcceptedGroupMember(groupID, inviterID)
	if err != nil || !isMember {
		return errors.New("unauthorized to invite users to this group")
	}

	// Verify invitee exists
	_, err = s.userRepo.GetUserByID(inviteeID)
	if err != nil {
		return errors.New("invitee not found")
	}

	status, err := s.membershipRepo.GetMembership(groupID, inviteeID)
	if err != nil {
		return err
	}

	if status != "none" {
		return errors.New("invitee is already a member or has a pending status")
	}

	if err := s.membershipRepo.AddMembership(groupID, inviteeID, "pending_invite"); err != nil {
		return err
	}

	// Notify invitee
	_ = s.notificationServ.CreateNotification(inviteeID, "group_invite", groupID, nil)

	return nil
}

func (s *groupService) RespondToMembership(groupID, userID, deciderID uuid.UUID, action string) error {
	g, err := s.groupRepo.GetGroupByID(groupID)
	if err != nil {
		return errors.New("group not found")
	}

	status, err := s.membershipRepo.GetMembership(groupID, userID)
	if err != nil || status == "none" {
		return errors.New("no pending membership request/invite found")
	}

	if status == "pending_request" {
		// Only group creator can decide on requests
		if deciderID != g.CreatorID {
			return errors.New("unauthorized to accept join requests")
		}
	} else if status == "pending_invite" {
		// Only the invited user can decide on invites
		if deciderID != userID {
			return errors.New("unauthorized to accept group invites on behalf of this user")
		}
	} else {
		return errors.New("membership is already accepted")
	}

	if action == "accept" {
		return s.membershipRepo.UpdateMembershipStatus(groupID, userID, "accepted")
	} else if action == "decline" || action == "reject" {
		return s.membershipRepo.RemoveMembership(groupID, userID)
	}

	return errors.New("invalid action")
}

func (s *groupService) ListMembers(groupID, viewerID uuid.UUID) ([]*dto.UserResponse, error) {
	// Verify viewer is an accepted member
	isMember, err := s.membershipRepo.IsAcceptedGroupMember(groupID, viewerID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized: must be a group member to view members list")
	}

	members, err := s.membershipRepo.ListGroupMembers(groupID)
	if err != nil {
		return nil, err
	}

	var response []*dto.UserResponse
	for _, m := range members {
		response = append(response, &dto.UserResponse{
			ID:          m.ID,
			Email:       m.Email,
			FirstName:   m.FirstName,
			LastName:    m.LastName,
			DateOfBirth: m.DOB.Format("2006-01-02"),
			Avatar:      m.Avatar,
			Nickname:    m.Nickname,
			AboutMe:     m.AboutMe,
			IsPublic:    m.IsPublic,
			CreatedAt:   m.CreatedAt,
		})
	}

	return response, nil
}

func (s *groupService) ListPendingRequests(groupID, creatorID uuid.UUID) ([]*dto.UserResponse, error) {
	g, err := s.groupRepo.GetGroupByID(groupID)
	if err != nil {
		return nil, errors.New("group not found")
	}

	if g.CreatorID != creatorID {
		return nil, errors.New("unauthorized: only the group creator can view pending requests")
	}

	requests, err := s.membershipRepo.ListPendingRequests(groupID)
	if err != nil {
		return nil, err
	}

	var response []*dto.UserResponse
	for _, m := range requests {
		response = append(response, &dto.UserResponse{
			ID:          m.ID,
			Email:       m.Email,
			FirstName:   m.FirstName,
			LastName:    m.LastName,
			DateOfBirth: m.DOB.Format("2006-01-02"),
			Avatar:      m.Avatar,
			Nickname:    m.Nickname,
			AboutMe:     m.AboutMe,
			IsPublic:    m.IsPublic,
			CreatedAt:   m.CreatedAt,
		})
	}

	return response, nil
}
