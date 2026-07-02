package repositories

import (
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

func TestUserRepositoryCRUDAndOptionalProfileFields(t *testing.T) {
	db := newPostCommentTestDB(t)
	repo := NewUserRepository(db)
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000101"))
	user := &models.User{
		ID:        userID,
		Email:     "repo@example.com",
		PassHash:  "hash",
		FirstName: "Repo",
		LastName:  "User",
		DOB:       time.Date(1998, 4, 12, 0, 0, 0, 0, time.UTC),
		Avatar:    "/uploads/avatars/repo.png",
		Nickname:  "repo",
		AboutMe:   "about",
		IsPublic:  true,
		CreatedAt: time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
	}

	if err := repo.CreateUser(user); err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	byID, err := repo.GetUserByID(userID)
	if err != nil {
		t.Fatalf("GetUserByID returned error: %v", err)
	}
	if byID.Email != user.Email || byID.Avatar != user.Avatar || byID.Nickname != user.Nickname || byID.AboutMe != user.AboutMe {
		t.Fatalf("stored user = %#v, want profile fields preserved", byID)
	}

	byEmail, err := repo.GetUserByEmail(user.Email)
	if err != nil {
		t.Fatalf("GetUserByEmail returned error: %v", err)
	}
	if byEmail.ID != userID {
		t.Fatalf("byEmail id = %s, want %s", byEmail.ID, userID)
	}

	user.Email = "updated@example.com"
	user.FirstName = "Updated"
	user.IsPublic = false
	if err := repo.UpdateUserProfile(user); err != nil {
		t.Fatalf("UpdateUserProfile returned error: %v", err)
	}
	updated, err := repo.GetUserByID(userID)
	if err != nil {
		t.Fatalf("GetUserByID after update returned error: %v", err)
	}
	if updated.Email != "updated@example.com" || updated.FirstName != "Updated" || updated.IsPublic {
		t.Fatalf("updated user = %#v", updated)
	}

	if err := repo.DeleteUser(userID); err != nil {
		t.Fatalf("DeleteUser returned error: %v", err)
	}
	if _, err := repo.GetUserByID(userID); err == nil {
		t.Fatal("expected deleted user lookup to fail")
	}
}

func TestSessionRepositoryCRUD(t *testing.T) {
	db := newPostCommentTestDB(t)
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000102"))
	sessionID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000102"))
	insertUser(t, db, userID, "session@example.com", "Session", "User")
	repo := NewSessionRepository(db)
	session := &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC),
		CreatedAt: time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
	}

	if err := repo.CreateSession(session); err != nil {
		t.Fatalf("CreateSession returned error: %v", err)
	}
	found, err := repo.GetSessionByID(sessionID)
	if err != nil {
		t.Fatalf("GetSessionByID returned error: %v", err)
	}
	if found.UserID != userID || !found.ExpiresAt.Equal(session.ExpiresAt) {
		t.Fatalf("session = %#v, want user %s and expiry %s", found, userID, session.ExpiresAt)
	}

	if err := repo.DeleteSession(sessionID); err != nil {
		t.Fatalf("DeleteSession returned error: %v", err)
	}
	if _, err := repo.GetSessionByID(sessionID); err == nil {
		t.Fatal("expected deleted session lookup to fail")
	}
}

func TestFollowerRepositoryTransitionsAndListsAcceptedRelationships(t *testing.T) {
	db := newPostCommentTestDB(t)
	repo := NewFollowerRepository(db)
	followerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000103"))
	followeeID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000104"))
	pendingFollowerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000105"))
	insertUser(t, db, followerID, "follower@example.com", "Follower", "One")
	insertUser(t, db, followeeID, "followee@example.com", "Followee", "One")
	insertUser(t, db, pendingFollowerID, "pending@example.com", "Pending", "One")

	status, err := repo.GetStatus(followerID, followeeID)
	if err != nil {
		t.Fatalf("GetStatus missing returned error: %v", err)
	}
	if status != "none" {
		t.Fatalf("missing status = %q, want none", status)
	}

	if err := repo.Follow(followerID, followeeID, models.Pending); err != nil {
		t.Fatalf("Follow pending returned error: %v", err)
	}
	if err := repo.AcceptFollower(followerID, followeeID); err != nil {
		t.Fatalf("AcceptFollower returned error: %v", err)
	}
	if err := repo.Follow(pendingFollowerID, followeeID, models.Pending); err != nil {
		t.Fatalf("Follow second pending returned error: %v", err)
	}

	followers, err := repo.GetFollowers(followeeID)
	if err != nil {
		t.Fatalf("GetFollowers returned error: %v", err)
	}
	if len(followers) != 1 || followers[0].ID != followerID {
		t.Fatalf("followers = %#v, want only accepted follower", followers)
	}
	following, err := repo.GetFollowing(followerID)
	if err != nil {
		t.Fatalf("GetFollowing returned error: %v", err)
	}
	if len(following) != 1 || following[0].ID != followeeID {
		t.Fatalf("following = %#v, want accepted followee", following)
	}

	if err := repo.RejectFollower(pendingFollowerID, followeeID); err != nil {
		t.Fatalf("RejectFollower returned error: %v", err)
	}
	if status, _ := repo.GetStatus(pendingFollowerID, followeeID); status != "none" {
		t.Fatalf("rejected status = %q, want none", status)
	}
	if err := repo.Unfollow(followerID, followeeID); err != nil {
		t.Fatalf("Unfollow returned error: %v", err)
	}
	if err := repo.Unfollow(followerID, followeeID); err == nil {
		t.Fatal("expected unfollowing a missing relationship to fail")
	}
}

func TestMessageRepositoryListDMCandidatesExcludesActiveThreads(t *testing.T) {
	db := newPostCommentTestDB(t)
	repo := NewMessageRepository(db)
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000120"))
	recentID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000121"))
	activeID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000122"))
	emptyThreadID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000123"))
	pendingID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000124"))
	for _, user := range []struct {
		id    uuid.UUID
		email string
		first string
	}{
		{viewerID, "viewer-dm@example.com", "Viewer"},
		{recentID, "recent-dm@example.com", "Recent"},
		{activeID, "active-dm@example.com", "Active"},
		{emptyThreadID, "empty-dm@example.com", "Empty"},
		{pendingID, "pending-dm@example.com", "Pending"},
	} {
		insertUser(t, db, user.id, user.email, user.first, "Candidate")
	}

	base := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	for _, rel := range []struct {
		follower uuid.UUID
		followee uuid.UUID
		status   string
		created  time.Time
	}{
		{recentID, viewerID, "accepted", base.Add(3 * time.Minute)},
		{viewerID, activeID, "accepted", base.Add(2 * time.Minute)},
		{viewerID, emptyThreadID, "accepted", base.Add(time.Minute)},
		{viewerID, pendingID, "pending", base.Add(4 * time.Minute)},
	} {
		if _, err := db.Exec(`INSERT INTO followers (follower_id, followee_id, status, created_at) VALUES (?, ?, ?, ?)`, rel.follower.String(), rel.followee.String(), rel.status, rel.created.Format(time.RFC3339)); err != nil {
			t.Fatalf("insert follower relation: %v", err)
		}
	}

	activeThread, err := repo.GetOrCreateDMThread(viewerID, activeID)
	if err != nil {
		t.Fatalf("GetOrCreateDMThread active returned error: %v", err)
	}
	if _, err := repo.GetOrCreateDMThread(viewerID, emptyThreadID); err != nil {
		t.Fatalf("GetOrCreateDMThread empty returned error: %v", err)
	}
	messageID := uuid.Must(uuid.FromString("30000000-0000-0000-0000-000000000120"))
	if err := repo.CreateMessage(&models.Message{ID: messageID, SenderID: viewerID, DMThreadID: &activeThread.ID, Content: "already active", CreatedAt: base.Add(5 * time.Minute)}); err != nil {
		t.Fatalf("CreateMessage returned error: %v", err)
	}

	candidates, err := repo.ListDMCandidates(viewerID, 10)
	if err != nil {
		t.Fatalf("ListDMCandidates returned error: %v", err)
	}
	if len(candidates) != 2 {
		t.Fatalf("candidate count = %d, want 2: %#v", len(candidates), candidates)
	}
	if candidates[0].ID != recentID || candidates[1].ID != emptyThreadID {
		t.Fatalf("candidate order = [%s, %s], want recent then empty-thread", candidates[0].ID, candidates[1].ID)
	}
}

func TestGroupMembershipRepositoryChecksAcceptedOnly(t *testing.T) {
	db := newPostCommentTestDB(t)
	repo := NewGroupMembershipRepository(db)
	creatorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000106"))
	acceptedID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000107"))
	pendingID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000108"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000106"))
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)
	insertUser(t, db, creatorID, "creator@example.com", "Group", "Creator")
	insertUser(t, db, acceptedID, "accepted@example.com", "Accepted", "Member")
	insertUser(t, db, pendingID, "pending-member@example.com", "Pending", "Member")
	if _, err := db.Exec(
		`INSERT INTO groups (id, creator_id, title, created_at) VALUES (?, ?, 'Backend Readers', ?)`,
		groupID.String(),
		creatorID.String(),
		now.Format(time.RFC3339),
	); err != nil {
		t.Fatalf("insert group: %v", err)
	}
	for _, row := range []struct {
		userID uuid.UUID
		status string
	}{
		{acceptedID, "accepted"},
		{pendingID, "pending_request"},
	} {
		if _, err := db.Exec(
			`INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)`,
			groupID.String(),
			row.userID.String(),
			row.status,
		); err != nil {
			t.Fatalf("insert group member %s: %v", row.userID, err)
		}
	}

	accepted, err := repo.IsAcceptedGroupMember(groupID, acceptedID)
	if err != nil {
		t.Fatalf("IsAcceptedGroupMember accepted returned error: %v", err)
	}
	if !accepted {
		t.Fatal("expected accepted member to be allowed")
	}
	pending, err := repo.IsAcceptedGroupMember(groupID, pendingID)
	if err != nil {
		t.Fatalf("IsAcceptedGroupMember pending returned error: %v", err)
	}
	if pending {
		t.Fatal("expected pending member not to be accepted")
	}
}
