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
	CreateGroup(creatorID uuid.UUID, title, description, avatar string) (*dto.GroupResponse, error)
	GetGroup(id, viewerID uuid.UUID) (*dto.GroupResponse, error)
	UpdateGroup(groupID, actorID uuid.UUID, title, description, avatar string) (*dto.GroupResponse, error)
	PromoteMember(groupID, actorID, userID uuid.UUID) error
	DemoteMember(groupID, actorID, userID uuid.UUID) error
	ListGroups(viewerID uuid.UUID) ([]*dto.GroupResponse, error)
	RequestJoin(groupID, userID uuid.UUID) error
	InviteUser(groupID, inviterID, inviteeID uuid.UUID) error
	RespondToMembership(groupID, userID, deciderID uuid.UUID, action string) error
	LeaveGroup(groupID, userID uuid.UUID) error
	ListMembers(groupID, viewerID uuid.UUID) ([]*dto.UserResponse, error)
	ListPendingRequests(groupID, actorID uuid.UUID) ([]*dto.UserResponse, error)
	ListPendingInvitations(groupID, actorID uuid.UUID) ([]*dto.UserResponse, error)
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

func (s *groupService) CreateGroup(creatorID uuid.UUID, title, description, avatar string) (*dto.GroupResponse, error) {
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
		Avatar:      avatar,
		CreatedAt:   time.Now(),
	}

	if err := s.groupRepo.CreateGroup(g); err != nil {
		return nil, err
	}

	// Add creator as accepted admin member.
	if err := s.membershipRepo.AddMembership(groupID, creatorID, "accepted"); err != nil {
		return nil, err
	}
	if err := s.membershipRepo.UpdateMembershipRole(groupID, creatorID, "admin"); err != nil {
		return nil, err
	}

	response := dto.MapGroupResponse(g, "accepted")
	response.Role = "admin"
	return response, nil
}

func (s *groupService) GetGroup(id, viewerID uuid.UUID) (*dto.GroupResponse, error) {
	g, err := s.groupRepo.GetGroupByID(id)
	if err != nil {
		return nil, err
	}
	status, err := s.membershipRepo.GetMembership(id, viewerID)
	if err != nil {
		return nil, err
	}
	response := dto.MapGroupResponse(g, status)
	if status == "accepted" {
		role, err := s.membershipRepo.GetMembershipRole(id, viewerID)
		if err != nil {
			return nil, err
		}
		response.Role = role
	}
	return response, nil
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

		item := dto.MapGroupResponse(g, status)
		if status == "accepted" {
			role, err := s.membershipRepo.GetMembershipRole(g.ID, viewerID)
			if err != nil {
				return nil, err
			}
			item.Role = role
		}
		response = append(response, item)
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
	// Verify inviter is a group admin.
	isAdmin, err := s.membershipRepo.IsGroupAdmin(groupID, inviterID)
	if err != nil || !isAdmin {
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
	if _, err := s.groupRepo.GetGroupByID(groupID); err != nil {
		return errors.New("group not found")
	}

	status, err := s.membershipRepo.GetMembership(groupID, userID)
	if err != nil || status == "none" {
		return errors.New("no pending membership request/invite found")
	}

	if status == "pending_request" {
		// Group admins can decide on requests.
		isAdmin, err := s.membershipRepo.IsGroupAdmin(groupID, deciderID)
		if err != nil || !isAdmin {
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

func (s *groupService) LeaveGroup(groupID, userID uuid.UUID) error {
	if _, err := s.groupRepo.GetGroupByID(groupID); err != nil {
		return errors.New("group not found")
	}
	status, err := s.membershipRepo.GetMembership(groupID, userID)
	if err != nil {
		return err
	}
	if status != "accepted" {
		return errors.New("only accepted group members can leave")
	}
	members, err := s.membershipRepo.ListGroupMembers(groupID)
	if err != nil {
		return err
	}
	if len(members) <= 1 {
		return s.groupRepo.DeleteGroup(groupID)
	}
	isAdmin, err := s.membershipRepo.IsGroupAdmin(groupID, userID)
	if err != nil {
		return err
	}
	if isAdmin {
		adminCount, err := s.membershipRepo.CountGroupAdmins(groupID)
		if err != nil {
			return err
		}
		if adminCount <= 1 {
			return errors.New("assign another group admin before leaving")
		}
	}
	return s.membershipRepo.RemoveMembership(groupID, userID)
}

func (s *groupService) ListMembers(groupID, viewerID uuid.UUID) ([]*dto.UserResponse, error) {
	// Verify viewer is an accepted member
	isMember, err := s.membershipRepo.IsAcceptedGroupMember(groupID, viewerID)
	if err != nil || !isMember {
		return nil, errors.New("unauthorized: must be a group member to view members list")
	}

	members, err := s.membershipRepo.ListGroupMembersWithRoles(groupID)
	if err != nil {
		return nil, err
	}

	var response []*dto.UserResponse
	for _, m := range members {
		item := mapUserResponse(&m.User)
		item.Role = m.Role
		response = append(response, item)
	}

	return response, nil
}

func (s *groupService) ListPendingRequests(groupID, actorID uuid.UUID) ([]*dto.UserResponse, error) {
	isAdmin, err := s.membershipRepo.IsGroupAdmin(groupID, actorID)
	if err != nil || !isAdmin {
		return nil, errors.New("unauthorized: only group admins can view pending requests")
	}

	requests, err := s.membershipRepo.ListPendingRequests(groupID)
	if err != nil {
		return nil, err
	}

	var response []*dto.UserResponse
	for _, m := range requests {
		response = append(response, mapUserResponse(m))
	}

	return response, nil
}

func (s *groupService) UpdateGroup(groupID, actorID uuid.UUID, title, description, avatar string) (*dto.GroupResponse, error) {
	isAdmin, err := s.membershipRepo.IsGroupAdmin(groupID, actorID)
	if err != nil || !isAdmin {
		return nil, errors.New("unauthorized: only group admins can edit group details")
	}
	g, err := s.groupRepo.GetGroupByID(groupID)
	if err != nil {
		return nil, errors.New("group not found")
	}
	if title != "" {
		g.Title = title
	}
	g.Description = description
	g.Avatar = avatar
	if err := s.groupRepo.UpdateGroup(g); err != nil {
		return nil, err
	}
	response := dto.MapGroupResponse(g, "accepted")
	response.Role = "admin"
	return response, nil
}

func (s *groupService) PromoteMember(groupID, actorID, userID uuid.UUID) error {
	isAdmin, err := s.membershipRepo.IsGroupAdmin(groupID, actorID)
	if err != nil || !isAdmin {
		return errors.New("unauthorized: only group admins can promote members")
	}
	status, err := s.membershipRepo.GetMembership(groupID, userID)
	if err != nil {
		return err
	}
	if status != "accepted" {
		return errors.New("only accepted members can be promoted")
	}
	return s.membershipRepo.UpdateMembershipRole(groupID, userID, "admin")
}

func (s *groupService) DemoteMember(groupID, actorID, userID uuid.UUID) error {
	isAdmin, err := s.membershipRepo.IsGroupAdmin(groupID, actorID)
	if err != nil || !isAdmin {
		return errors.New("unauthorized: only group admins can demote members")
	}
	role, err := s.membershipRepo.GetMembershipRole(groupID, userID)
	if err != nil {
		return err
	}
	if role != "admin" {
		return errors.New("member is not an admin")
	}
	adminCount, err := s.membershipRepo.CountGroupAdmins(groupID)
	if err != nil {
		return err
	}
	if adminCount <= 1 {
		return errors.New("cannot demote the last group admin")
	}
	return s.membershipRepo.UpdateMembershipRole(groupID, userID, "member")
}

func (s *groupService) ListPendingInvitations(groupID, actorID uuid.UUID) ([]*dto.UserResponse, error) {
	isAdmin, err := s.membershipRepo.IsGroupAdmin(groupID, actorID)
	if err != nil || !isAdmin {
		return nil, errors.New("unauthorized: only group admins can view pending invitations")
	}
	invites, err := s.membershipRepo.ListPendingInvitations(groupID)
	if err != nil {
		return nil, err
	}
	response := make([]*dto.UserResponse, 0, len(invites))
	for _, invite := range invites {
		response = append(response, mapUserResponse(invite))
	}
	return response, nil
}

func mapUserResponse(user *models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:          user.ID,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		DateOfBirth: user.DOB.Format("2006-01-02"),
		Avatar:      user.Avatar,
		Nickname:    user.Nickname,
		AboutMe:     user.AboutMe,
		IsPublic:    user.IsPublic,
		CreatedAt:   user.CreatedAt,
	}
}
