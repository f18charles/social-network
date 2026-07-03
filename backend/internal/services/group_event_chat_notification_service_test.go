package services

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

func TestGroupServiceMembershipTransitionsAndNotifications(t *testing.T) {
	creatorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	memberID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	requesterID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	inviteeID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000004"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))

	groups := newServiceTestGroupRepo()
	groups.groups[groupID] = &models.Group{ID: groupID, CreatorID: creatorID, Title: "Testing guild", CreatedAt: time.Now()}
	memberships := newServiceTestMembershipRepo()
	memberships.members[groupMemberKey{groupID: groupID, userID: creatorID}] = "accepted"
	memberships.roles[groupMemberKey{groupID: groupID, userID: creatorID}] = "admin"
	memberships.members[groupMemberKey{groupID: groupID, userID: memberID}] = "accepted"
	users := newServiceTestUserRepo()
	users.users[inviteeID] = &models.User{ID: inviteeID, Email: "invitee@example.test"}
	notifications := &serviceTestNotificationService{}
	service := NewGroupService(groups, memberships, users, notifications)

	if err := service.RequestJoin(groupID, requesterID); err != nil {
		t.Fatalf("RequestJoin returned error: %v", err)
	}
	if got := memberships.members[groupMemberKey{groupID: groupID, userID: requesterID}]; got != "pending_request" {
		t.Fatalf("requester membership = %q, want pending_request", got)
	}
	if !notifications.has(creatorID, "group_request", requesterID) {
		t.Fatalf("expected group_request notification for creator, got %#v", notifications.created)
	}

	if err := service.RespondToMembership(groupID, requesterID, memberID, "accept"); err == nil {
		t.Fatal("expected non-creator to be blocked from accepting join request")
	}
	if err := service.RespondToMembership(groupID, requesterID, creatorID, "accept"); err != nil {
		t.Fatalf("creator accept returned error: %v", err)
	}
	if got := memberships.members[groupMemberKey{groupID: groupID, userID: requesterID}]; got != "accepted" {
		t.Fatalf("requester membership after accept = %q, want accepted", got)
	}

	if err := service.InviteUser(groupID, requesterID, inviteeID); err == nil {
		t.Fatal("expected non-admin accepted member invite to be blocked")
	}
	if err := service.InviteUser(groupID, creatorID, inviteeID); err != nil {
		t.Fatalf("admin invite returned error: %v", err)
	}
	if got := memberships.members[groupMemberKey{groupID: groupID, userID: inviteeID}]; got != "pending_invite" {
		t.Fatalf("invitee membership = %q, want pending_invite", got)
	}
	if !notifications.has(inviteeID, "group_invite", groupID) {
		t.Fatalf("expected group_invite notification for invitee, got %#v", notifications.created)
	}

	if err := service.RespondToMembership(groupID, inviteeID, creatorID, "accept"); err == nil {
		t.Fatal("expected creator to be blocked from accepting invite for invitee")
	}
	if err := service.RespondToMembership(groupID, inviteeID, inviteeID, "accept"); err != nil {
		t.Fatalf("invitee accept returned error: %v", err)
	}
}

func TestGroupServiceLeaveGroupRules(t *testing.T) {
	adminID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	memberID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	otherAdminID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	pendingID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000004"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000010"))

	groups := newServiceTestGroupRepo()
	groups.groups[groupID] = &models.Group{ID: groupID, CreatorID: adminID, Title: "Leaving", CreatedAt: time.Now()}
	memberships := newServiceTestMembershipRepo()
	memberships.members[groupMemberKey{groupID: groupID, userID: adminID}] = "accepted"
	memberships.roles[groupMemberKey{groupID: groupID, userID: adminID}] = "admin"
	memberships.members[groupMemberKey{groupID: groupID, userID: memberID}] = "accepted"
	memberships.members[groupMemberKey{groupID: groupID, userID: pendingID}] = "pending_request"
	service := NewGroupService(groups, memberships, newServiceTestUserRepo(), &serviceTestNotificationService{})

	if err := service.LeaveGroup(groupID, pendingID); err == nil {
		t.Fatal("expected pending member to be blocked from leaving")
	}
	if err := service.LeaveGroup(groupID, adminID); err == nil {
		t.Fatal("expected last admin with other members to be blocked from leaving")
	}
	if err := service.LeaveGroup(groupID, memberID); err != nil {
		t.Fatalf("member LeaveGroup returned error: %v", err)
	}
	if got := memberships.members[groupMemberKey{groupID: groupID, userID: memberID}]; got != "" {
		t.Fatalf("member status after leave = %q, want removed", got)
	}

	memberships.members[groupMemberKey{groupID: groupID, userID: otherAdminID}] = "accepted"
	memberships.roles[groupMemberKey{groupID: groupID, userID: otherAdminID}] = "admin"
	if err := service.LeaveGroup(groupID, adminID); err != nil {
		t.Fatalf("admin LeaveGroup with another admin returned error: %v", err)
	}
}

func TestGroupServiceLeaveDeletesOnlyMemberGroup(t *testing.T) {
	memberID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000011"))
	groups := newServiceTestGroupRepo()
	groups.groups[groupID] = &models.Group{ID: groupID, CreatorID: memberID, Title: "Solo", CreatedAt: time.Now()}
	memberships := newServiceTestMembershipRepo()
	memberships.members[groupMemberKey{groupID: groupID, userID: memberID}] = "accepted"
	memberships.roles[groupMemberKey{groupID: groupID, userID: memberID}] = "admin"
	service := NewGroupService(groups, memberships, newServiceTestUserRepo(), &serviceTestNotificationService{})

	if err := service.LeaveGroup(groupID, memberID); err != nil {
		t.Fatalf("solo LeaveGroup returned error: %v", err)
	}
	if _, ok := groups.groups[groupID]; ok {
		t.Fatal("group remained after only member left")
	}
}

func TestEventServiceRequiresGroupMembershipAndMaintainsRSVPs(t *testing.T) {
	creatorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	memberID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	outsiderID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))
	when := time.Date(2026, 7, 12, 15, 30, 0, 0, time.UTC)

	events := newServiceTestEventRepo()
	memberships := newServiceTestMembershipRepo()
	memberships.members[groupMemberKey{groupID: groupID, userID: creatorID}] = "accepted"
	memberships.roles[groupMemberKey{groupID: groupID, userID: creatorID}] = "admin"
	memberships.members[groupMemberKey{groupID: groupID, userID: memberID}] = "accepted"
	memberships.users[groupID] = []*models.User{{ID: creatorID}, {ID: memberID}}
	notifications := &serviceTestNotificationService{}
	service := NewEventService(events, memberships, notifications)

	if _, err := service.CreateEvent(outsiderID, groupID, "Blocked", "", when); err == nil {
		t.Fatal("expected non-member to be blocked from creating event")
	}

	created, err := service.CreateEvent(creatorID, groupID, "Planning session", "Discuss scope", when)
	if err != nil {
		t.Fatalf("CreateEvent returned error: %v", err)
	}
	if created.UserRSVP != "going" || created.GoingCount != 1 || created.NotGoingCount != 0 {
		t.Fatalf("created event RSVP summary = %#v", created)
	}
	if !notifications.has(memberID, "event_created", created.ID) {
		t.Fatalf("expected event_created notification for other member, got %#v", notifications.created)
	}

	if err := service.RespondToEvent(created.ID, outsiderID, "going"); err == nil {
		t.Fatal("expected non-member RSVP to be blocked")
	}
	if err := service.RespondToEvent(created.ID, memberID, "maybe"); err == nil {
		t.Fatal("expected invalid RSVP status to be rejected")
	}
	if err := service.RespondToEvent(created.ID, memberID, "not_going"); err != nil {
		t.Fatalf("member RSVP returned error: %v", err)
	}

	read, err := service.GetEvent(created.ID, memberID)
	if err != nil {
		t.Fatalf("GetEvent returned error: %v", err)
	}
	if read.UserRSVP != "not_going" || read.GoingCount != 1 || read.NotGoingCount != 1 {
		t.Fatalf("event RSVP summary = %#v, want creator going and member not going", read)
	}
}

func TestChatServiceEnforcesMessageBusinessRules(t *testing.T) {
	senderID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	recipientID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))
	messages := newServiceTestMessageRepo()
	followers := newServiceTestFollowerRepo()
	memberships := newServiceTestMembershipRepo()
	service := NewChatService(messages, followers, memberships, newServiceTestUserRepo(), newServiceTestGroupRepo(), &serviceTestNotificationService{}, &serviceTestMessageReactionRepo{})
	recipient := recipientID.String()

	if _, err := service.SendMessage(senderID, dto.SendMessageRequest{RecipientID: &recipient, Content: ""}); err == nil {
		t.Fatal("expected empty message content to be rejected")
	}
	if _, err := service.SendMessage(senderID, dto.SendMessageRequest{RecipientID: &recipient, Content: "hello"}); err == nil {
		t.Fatal("expected DM without follow relationship to be rejected")
	}

	followers.status[followerKey{followerID: recipientID, followeeID: senderID}] = models.Accepted
	message, err := service.SendMessage(senderID, dto.SendMessageRequest{RecipientID: &recipient, Content: "hello"})
	if err != nil {
		t.Fatalf("SendMessage with reverse follow returned error: %v", err)
	}
	if message.Content != "hello" || message.DMThreadID == nil {
		t.Fatalf("DM response = %#v, want stored thread message", message)
	}

	group := groupID.String()
	if _, err := service.SendMessage(senderID, dto.SendMessageRequest{GroupID: &group, Content: "group hello"}); err == nil {
		t.Fatal("expected non-member group message to be rejected")
	}
	memberships.members[groupMemberKey{groupID: groupID, userID: senderID}] = "accepted"
	groupMessage, err := service.SendMessage(senderID, dto.SendMessageRequest{GroupID: &group, Content: "group hello"})
	if err != nil {
		t.Fatalf("SendMessage accepted group member returned error: %v", err)
	}
	if groupMessage.GroupID == nil || *groupMessage.GroupID != groupID {
		t.Fatalf("group response = %#v, want group id %s", groupMessage, groupID)
	}
}

func TestChatServiceListDMCandidatesMapsUsers(t *testing.T) {
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	candidateID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	messages := newServiceTestMessageRepo()
	messages.candidates = []*models.User{{ID: candidateID, Email: "candidate@example.test", FirstName: "Candidate", LastName: "User", DOB: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC), CreatedAt: time.Now()}}
	service := NewChatService(messages, newServiceTestFollowerRepo(), newServiceTestMembershipRepo(), newServiceTestUserRepo(), newServiceTestGroupRepo(), &serviceTestNotificationService{}, &serviceTestMessageReactionRepo{})

	candidates, err := service.ListDMCandidates(userID, 10)
	if err != nil {
		t.Fatalf("ListDMCandidates returned error: %v", err)
	}
	if len(candidates) != 1 || candidates[0].ID != candidateID || candidates[0].Email != "candidate@example.test" {
		t.Fatalf("candidates = %#v", candidates)
	}
}

func TestNotificationServiceOwnershipFormattingAndPush(t *testing.T) {
	ownerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	otherID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	sourceID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	notifications := newServiceTestNotificationRepo()
	users := newServiceTestUserRepo()
	users.users[sourceID] = &models.User{ID: sourceID, FirstName: "Ada", LastName: "Lovelace", Email: "ada@example.test"}
	service := NewNotificationService(notifications, users, newServiceTestGroupRepo(), newServiceTestEventRepo())

	var pushedUser uuid.UUID
	var pushedPayload any
	service.RegisterPushHandler(func(userID uuid.UUID, payload any) {
		pushedUser = userID
		pushedPayload = payload
	})

	if err := service.CreateNotification(ownerID, "follow_request", sourceID, nil); err != nil {
		t.Fatalf("CreateNotification returned error: %v", err)
	}
	if pushedUser != ownerID || pushedPayload == nil {
		t.Fatalf("push payload user/payload = %s/%#v, want owner and payload", pushedUser, pushedPayload)
	}

	list, err := service.GetNotifications(ownerID)
	if err != nil {
		t.Fatalf("GetNotifications returned error: %v", err)
	}
	if len(list) != 1 || !strings.Contains(list[0].Message, "Ada Lovelace requested to follow you") {
		t.Fatalf("formatted notifications = %#v", list)
	}

	if err := service.MarkAsRead(list[0].ID, otherID); err == nil {
		t.Fatal("expected other user to be blocked from marking notification read")
	}
	if err := service.MarkAsRead(list[0].ID, ownerID); err != nil {
		t.Fatalf("owner MarkAsRead returned error: %v", err)
	}
	if !notifications.notifications[list[0].ID].IsRead {
		t.Fatal("notification was not marked read")
	}
}

type serviceTestGroupRepo struct {
	groups map[uuid.UUID]*models.Group
}

func newServiceTestGroupRepo() *serviceTestGroupRepo {
	return &serviceTestGroupRepo{groups: map[uuid.UUID]*models.Group{}}
}

func (r *serviceTestGroupRepo) CreateGroup(group *models.Group) error {
	r.groups[group.ID] = group
	return nil
}

func (r *serviceTestGroupRepo) GetGroupByID(id uuid.UUID) (*models.Group, error) {
	group, ok := r.groups[id]
	if !ok {
		return nil, errors.New("group not found")
	}
	return group, nil
}

func (r *serviceTestGroupRepo) ListGroups() ([]*models.Group, error) {
	groups := make([]*models.Group, 0, len(r.groups))
	for _, group := range r.groups {
		groups = append(groups, group)
	}
	return groups, nil
}

func (r *serviceTestGroupRepo) UpdateGroup(group *models.Group) error {
	r.groups[group.ID] = group
	return nil
}

func (r *serviceTestGroupRepo) DeleteGroup(id uuid.UUID) error {
	delete(r.groups, id)
	return nil
}

type serviceTestMembershipRepo struct {
	members map[groupMemberKey]string
	roles   map[groupMemberKey]string
	users   map[uuid.UUID][]*models.User
}

func newServiceTestMembershipRepo() *serviceTestMembershipRepo {
	return &serviceTestMembershipRepo{
		members: map[groupMemberKey]string{},
		roles:   map[groupMemberKey]string{},
		users:   map[uuid.UUID][]*models.User{},
	}
}

func (r *serviceTestMembershipRepo) IsAcceptedGroupMember(groupID, userID uuid.UUID) (bool, error) {
	return r.members[groupMemberKey{groupID: groupID, userID: userID}] == "accepted", nil
}

func (r *serviceTestMembershipRepo) IsGroupAdmin(groupID, userID uuid.UUID) (bool, error) {
	key := groupMemberKey{groupID: groupID, userID: userID}
	return r.members[key] == "accepted" && r.roles[key] == "admin", nil
}

func (r *serviceTestMembershipRepo) CountGroupAdmins(groupID uuid.UUID) (int, error) {
	count := 0
	for key, role := range r.roles {
		if key.groupID == groupID && role == "admin" && r.members[key] == "accepted" {
			count++
		}
	}
	return count, nil
}

func (r *serviceTestMembershipRepo) GetMembershipRole(groupID, userID uuid.UUID) (string, error) {
	return r.roles[groupMemberKey{groupID: groupID, userID: userID}], nil
}

func (r *serviceTestMembershipRepo) UpdateMembershipRole(groupID, userID uuid.UUID, role string) error {
	r.roles[groupMemberKey{groupID: groupID, userID: userID}] = role
	return nil
}

func (r *serviceTestMembershipRepo) GetMembership(groupID, userID uuid.UUID) (string, error) {
	status := r.members[groupMemberKey{groupID: groupID, userID: userID}]
	if status == "" {
		return "none", nil
	}
	return status, nil
}

func (r *serviceTestMembershipRepo) AddMembership(groupID, userID uuid.UUID, status string) error {
	r.members[groupMemberKey{groupID: groupID, userID: userID}] = status
	return nil
}

func (r *serviceTestMembershipRepo) UpdateMembershipStatus(groupID, userID uuid.UUID, status string) error {
	r.members[groupMemberKey{groupID: groupID, userID: userID}] = status
	return nil
}

func (r *serviceTestMembershipRepo) RemoveMembership(groupID, userID uuid.UUID) error {
	delete(r.members, groupMemberKey{groupID: groupID, userID: userID})
	return nil
}

func (r *serviceTestMembershipRepo) ListGroupMembers(groupID uuid.UUID) ([]*models.User, error) {
	if users := r.users[groupID]; users != nil {
		return users, nil
	}
	var users []*models.User
	for key, status := range r.members {
		if key.groupID == groupID && status == "accepted" {
			users = append(users, &models.User{ID: key.userID})
		}
	}
	return users, nil
}

func (r *serviceTestMembershipRepo) ListPendingRequests(groupID uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	for key, status := range r.members {
		if key.groupID == groupID && status == "pending_request" {
			users = append(users, &models.User{ID: key.userID})
		}
	}
	return users, nil
}

func (r *serviceTestMembershipRepo) ListGroupMembersWithRoles(groupID uuid.UUID) ([]*models.GroupMemberUser, error) {
	users, err := r.ListGroupMembers(groupID)
	if err != nil {
		return nil, err
	}
	members := make([]*models.GroupMemberUser, 0, len(users))
	for _, user := range users {
		members = append(members, &models.GroupMemberUser{User: *user, Status: "accepted", Role: r.roles[groupMemberKey{groupID: groupID, userID: user.ID}]})
	}
	return members, nil
}

func (r *serviceTestMembershipRepo) ListPendingInvitations(groupID uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	for key, status := range r.members {
		if key.groupID == groupID && status == "pending_invite" {
			users = append(users, &models.User{ID: key.userID})
		}
	}
	return users, nil
}

type serviceTestUserRepo struct {
	users map[uuid.UUID]*models.User
}

func newServiceTestUserRepo() *serviceTestUserRepo {
	return &serviceTestUserRepo{users: map[uuid.UUID]*models.User{}}
}

func (r *serviceTestUserRepo) CreateUser(user *models.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *serviceTestUserRepo) GetUserByID(id uuid.UUID) (*models.User, error) {
	user, ok := r.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (r *serviceTestUserRepo) GetUserByEmail(email string) (*models.User, error) {
	for _, user := range r.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

func (r *serviceTestUserRepo) ListUsers(query string, excludeID uuid.UUID) ([]*models.User, error) {
	var users []*models.User
	for _, user := range r.users {
		if user.ID != excludeID {
			users = append(users, user)
		}
	}
	return users, nil
}

func (r *serviceTestUserRepo) UpdateUserProfile(user *models.User) error {
	r.users[user.ID] = user
	return nil
}

func (r *serviceTestUserRepo) DeleteUser(id uuid.UUID) error {
	delete(r.users, id)
	return nil
}

type serviceTestNotification struct {
	userID   uuid.UUID
	nType    string
	sourceID uuid.UUID
	groupID  *uuid.UUID
}

type serviceTestNotificationService struct {
	created     []serviceTestNotification
	pushHandler func(userID uuid.UUID, payload any)
}

func (s *serviceTestNotificationService) CreateNotification(userID uuid.UUID, nType string, sourceID uuid.UUID, groupID *uuid.UUID) error {
	s.created = append(s.created, serviceTestNotification{userID: userID, nType: nType, sourceID: sourceID, groupID: groupID})
	if s.pushHandler != nil {
		s.pushHandler(userID, nType)
	}
	return nil
}

func (s *serviceTestNotificationService) GetNotifications(userID uuid.UUID) ([]*models.NotificationResponse, error) {
	return nil, nil
}

func (s *serviceTestNotificationService) MarkAsRead(id, userID uuid.UUID) error {
	return nil
}

func (s *serviceTestNotificationService) MarkAllAsRead(userID uuid.UUID) error {
	return nil
}

func (s *serviceTestNotificationService) RegisterPushHandler(handler func(userID uuid.UUID, payload any)) {
	s.pushHandler = handler
}

func (s *serviceTestNotificationService) has(userID uuid.UUID, nType string, sourceID uuid.UUID) bool {
	for _, created := range s.created {
		if created.userID == userID && created.nType == nType && created.sourceID == sourceID {
			return true
		}
	}
	return false
}

type serviceTestEventRepo struct {
	events map[uuid.UUID]*models.Event
	rsvps  map[groupMemberKey]string
}

func newServiceTestEventRepo() *serviceTestEventRepo {
	return &serviceTestEventRepo{
		events: map[uuid.UUID]*models.Event{},
		rsvps:  map[groupMemberKey]string{},
	}
}

func (r *serviceTestEventRepo) CreateEvent(event *models.Event) error {
	r.events[event.ID] = event
	return nil
}

func (r *serviceTestEventRepo) GetEventByID(id uuid.UUID) (*models.Event, error) {
	event, ok := r.events[id]
	if !ok {
		return nil, errors.New("event not found")
	}
	return event, nil
}

func (r *serviceTestEventRepo) ListEventsByGroup(groupID uuid.UUID) ([]*models.Event, error) {
	var events []*models.Event
	for _, event := range r.events {
		if event.GroupID == groupID {
			events = append(events, event)
		}
	}
	return events, nil
}

func (r *serviceTestEventRepo) GetRSVP(eventID, userID uuid.UUID) (string, error) {
	return r.rsvps[groupMemberKey{groupID: eventID, userID: userID}], nil
}

func (r *serviceTestEventRepo) SetRSVP(eventID, userID uuid.UUID, status string) error {
	r.rsvps[groupMemberKey{groupID: eventID, userID: userID}] = status
	return nil
}

func (r *serviceTestEventRepo) GetRSVPSummaries(eventID uuid.UUID) (int, int, error) {
	var going int
	var notGoing int
	for key, status := range r.rsvps {
		if key.groupID != eventID {
			continue
		}
		switch status {
		case "going":
			going++
		case "not_going":
			notGoing++
		}
	}
	return going, notGoing, nil
}

type serviceTestFollowerRepo struct {
	status map[followerKey]models.Status
}

func newServiceTestFollowerRepo() *serviceTestFollowerRepo {
	return &serviceTestFollowerRepo{status: map[followerKey]models.Status{}}
}

func (r *serviceTestFollowerRepo) Follow(followerID, followeeID uuid.UUID, status models.Status) error {
	r.status[followerKey{followerID: followerID, followeeID: followeeID}] = status
	return nil
}

func (r *serviceTestFollowerRepo) Unfollow(followerID, followeeID uuid.UUID) error {
	delete(r.status, followerKey{followerID: followerID, followeeID: followeeID})
	return nil
}

func (r *serviceTestFollowerRepo) AcceptFollower(followerID, followeeID uuid.UUID) error {
	r.status[followerKey{followerID: followerID, followeeID: followeeID}] = models.Accepted
	return nil
}

func (r *serviceTestFollowerRepo) RejectFollower(followerID, followeeID uuid.UUID) error {
	delete(r.status, followerKey{followerID: followerID, followeeID: followeeID})
	return nil
}

func (r *serviceTestFollowerRepo) GetFollowers(userID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}
func (r *serviceTestFollowerRepo) GetFollowing(userID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}
func (r *serviceTestFollowerRepo) GetPendingFollowers(userID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}
func (r *serviceTestFollowerRepo) GetStatus(followerID, followeeID uuid.UUID) (models.Status, error) {
	status := r.status[followerKey{followerID: followerID, followeeID: followeeID}]
	if status == "" {
		return "none", nil
	}
	return status, nil
}

type serviceTestMessageRepo struct {
	messages   []models.Message
	threads    map[uuid.UUID]*models.DMThread
	byPair     map[[2]uuid.UUID]uuid.UUID
	candidates []*models.User
}

func newServiceTestMessageRepo() *serviceTestMessageRepo {
	return &serviceTestMessageRepo{
		threads: map[uuid.UUID]*models.DMThread{},
		byPair:  map[[2]uuid.UUID]uuid.UUID{},
	}
}

func (r *serviceTestMessageRepo) CreateMessage(message *models.Message) error {
	r.messages = append(r.messages, *message)
	return nil
}

func (r *serviceTestMessageRepo) GetMessageByID(id uuid.UUID) (*models.Message, error) {
	for _, message := range r.messages {
		if message.ID == id {
			copy := message
			return &copy, nil
		}
	}
	return nil, errors.New("message not found")
}

func (r *serviceTestMessageRepo) ListMessagesByGroup(groupID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	return nil, nil
}

func (r *serviceTestMessageRepo) ListMessagesByThread(threadID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	return nil, nil
}

func (r *serviceTestMessageRepo) GetOrCreateDMThread(user1ID, user2ID uuid.UUID) (*models.DMThread, error) {
	key := [2]uuid.UUID{user1ID, user2ID}
	if existingID, ok := r.byPair[key]; ok {
		return r.threads[existingID], nil
	}
	threadID := uuid.Must(uuid.NewV4())
	thread := &models.DMThread{ID: threadID, User1ID: user1ID, User2ID: user2ID, LastMessageAt: time.Now()}
	r.threads[threadID] = thread
	r.byPair[key] = threadID
	return thread, nil
}

func (r *serviceTestMessageRepo) GetDMThreadByID(id uuid.UUID) (*models.DMThread, error) {
	thread, ok := r.threads[id]
	if !ok {
		return nil, errors.New("thread not found")
	}
	return thread, nil
}

func (r *serviceTestMessageRepo) ListConversations(userID uuid.UUID) ([]*models.Conversation, error) {
	return nil, nil
}

func (r *serviceTestMessageRepo) ListDMCandidates(userID uuid.UUID, limit int) ([]*models.User, error) {
	if limit > 0 && len(r.candidates) > limit {
		return r.candidates[:limit], nil
	}
	return r.candidates, nil
}

func (r *serviceTestMessageRepo) GetMessageReactions(messageID, viewerID uuid.UUID) ([]*models.MessageReactionSummary, error) {
	return []*models.MessageReactionSummary{}, nil
}

type serviceTestMessageReactionRepo struct {
}

func (r *serviceTestMessageReactionRepo) SetMessageReaction(messageID, userID uuid.UUID, emoji string) error {
	return nil
}

func (r *serviceTestMessageReactionRepo) DeleteMessageReaction(messageID, userID uuid.UUID, emoji string) error {
	return nil
}

func (r *serviceTestMessageReactionRepo) GetMessageReactionSummary(messageID, viewerID uuid.UUID) ([]*models.MessageReactionSummary, error) {
	return []*models.MessageReactionSummary{}, nil
}

type serviceTestNotificationRepo struct {
	notifications map[uuid.UUID]*models.Notification
}

func newServiceTestNotificationRepo() *serviceTestNotificationRepo {
	return &serviceTestNotificationRepo{notifications: map[uuid.UUID]*models.Notification{}}
}

func (r *serviceTestNotificationRepo) CreateNotification(n *models.Notification) error {
	copy := *n
	r.notifications[n.ID] = &copy
	return nil
}

func (r *serviceTestNotificationRepo) GetNotificationByID(id uuid.UUID) (*models.Notification, error) {
	notification, ok := r.notifications[id]
	if !ok {
		return nil, errors.New("notification not found")
	}
	return notification, nil
}

func (r *serviceTestNotificationRepo) ListNotificationsByUser(userID uuid.UUID) ([]*models.Notification, error) {
	var notifications []*models.Notification
	for _, notification := range r.notifications {
		if notification.UserID == userID {
			notifications = append(notifications, notification)
		}
	}
	return notifications, nil
}

func (r *serviceTestNotificationRepo) MarkAsRead(id uuid.UUID) error {
	r.notifications[id].IsRead = true
	return nil
}

func (r *serviceTestNotificationRepo) MarkAllAsRead(userID uuid.UUID) error {
	for _, notification := range r.notifications {
		if notification.UserID == userID {
			notification.IsRead = true
		}
	}
	return nil
}

