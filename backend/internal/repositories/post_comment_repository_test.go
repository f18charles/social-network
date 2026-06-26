package repositories

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	dbpkg "learn.zone01kisumu.ke/git/qquinton/social-network/internal/db"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

func TestPostRepositoryReturnsViewerVoteInReadModel(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)

	repo := NewPostRepository(db)
	post, err := repo.GetPostByID(ids.postID, ids.viewerID)
	if err != nil {
		t.Fatalf("GetPostByID returned error: %v", err)
	}

	if post.Post.ID != ids.postID {
		t.Fatalf("post id = %s, want %s", post.Post.ID, ids.postID)
	}
	if post.Author == nil || post.Author.ID != ids.authorID {
		t.Fatalf("author = %#v, want %s", post.Author, ids.authorID)
	}
	if post.ViewerVote != models.ViewerVoteLike {
		t.Fatalf("viewer vote = %q, want %q", post.ViewerVote, models.ViewerVoteLike)
	}
}

func TestCommentRepositoryReturnsTreeWithViewerVotes(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)

	repo := NewCommentRepository(db)
	comments, err := repo.ListCommentTreeByPost(ids.postID, ids.viewerID, 10, 0)
	if err != nil {
		t.Fatalf("ListCommentTreeByPost returned error: %v", err)
	}
	if len(comments) != 2 {
		t.Fatalf("got %d comments, want 2", len(comments))
	}

	byID := map[uuid.UUID]*models.CommentWithAuthor{}
	for _, comment := range comments {
		byID[comment.Comment.ID] = comment
	}
	if byID[ids.parentCommentID].ViewerVote != models.ViewerVoteNone {
		t.Fatalf("parent viewer vote = %q, want none", byID[ids.parentCommentID].ViewerVote)
	}
	if byID[ids.replyCommentID].ViewerVote != models.ViewerVoteDislike {
		t.Fatalf("reply viewer vote = %q, want dislike", byID[ids.replyCommentID].ViewerVote)
	}
	if byID[ids.replyCommentID].Comment.ParentCommentID == nil || *byID[ids.replyCommentID].Comment.ParentCommentID != ids.parentCommentID {
		t.Fatalf("reply parent = %#v, want %s", byID[ids.replyCommentID].Comment.ParentCommentID, ids.parentCommentID)
	}
}

func TestPostAudienceRepositoryReplacesAndChecksMembership(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)
	memberID := ids.otherViewerID

	repo := NewPostAudienceRepository(db)
	if err := repo.ReplacePostAudience(ids.postID, []uuid.UUID{ids.viewerID}); err != nil {
		t.Fatalf("ReplacePostAudience returned error: %v", err)
	}
	if err := repo.ReplacePostAudience(ids.postID, []uuid.UUID{memberID}); err != nil {
		t.Fatalf("ReplacePostAudience replace returned error: %v", err)
	}

	members, err := repo.ListPostAudience(ids.postID)
	if err != nil {
		t.Fatalf("ListPostAudience returned error: %v", err)
	}
	if len(members) != 1 || members[0] != memberID {
		t.Fatalf("members = %#v, want only %s", members, memberID)
	}

	isMember, err := repo.IsPostAudienceMember(ids.postID, memberID)
	if err != nil {
		t.Fatalf("IsPostAudienceMember returned error: %v", err)
	}
	if !isMember {
		t.Fatal("expected replacement member to be in audience")
	}
	oldMember, err := repo.IsPostAudienceMember(ids.postID, ids.viewerID)
	if err != nil {
		t.Fatalf("IsPostAudienceMember old member returned error: %v", err)
	}
	if oldMember {
		t.Fatal("expected previous member to be removed")
	}
}

func TestVoteRepositoriesUpdateCountsAndViewerVote(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)
	postVotes := NewPostVoteRepository(db)
	commentVotes := NewCommentVoteRepository(db)

	postSummary, err := postVotes.SetPostVote(ids.postID, ids.otherViewerID, models.VoteValueDislike)
	if err != nil {
		t.Fatalf("SetPostVote returned error: %v", err)
	}
	if postSummary.LikeCount != 1 || postSummary.DislikeCount != 1 || postSummary.ViewerVote != models.ViewerVoteDislike {
		t.Fatalf("post summary = %#v, want 1 like, 1 dislike, viewer dislike", postSummary)
	}

	postSummary, err = postVotes.DeletePostVote(ids.postID, ids.viewerID)
	if err != nil {
		t.Fatalf("DeletePostVote returned error: %v", err)
	}
	if postSummary.LikeCount != 0 || postSummary.DislikeCount != 1 || postSummary.ViewerVote != models.ViewerVoteNone {
		t.Fatalf("post summary after delete = %#v, want 0 likes, 1 dislike, viewer none", postSummary)
	}

	commentSummary, err := commentVotes.SetCommentVote(ids.replyCommentID, ids.viewerID, models.VoteValueLike)
	if err != nil {
		t.Fatalf("SetCommentVote returned error: %v", err)
	}
	if commentSummary.LikeCount != 1 || commentSummary.DislikeCount != 0 || commentSummary.ViewerVote != models.ViewerVoteLike {
		t.Fatalf("comment summary = %#v, want 1 like, 0 dislikes, viewer like", commentSummary)
	}
}

func TestPostRepositoryHomeFeedPrivacyAndStableOrdering(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedFeedPrivacyRows(t, db)
	repo := NewPostRepository(db)

	selectedPosts, err := repo.ListHomeFeed(ids.selectedFollowerID, 10, 0)
	if err != nil {
		t.Fatalf("ListHomeFeed selected follower returned error: %v", err)
	}
	assertPostIDs(t, selectedPosts, ids.privatePostID, ids.almostPrivatePostID, ids.publicPostID)
	if selectedPosts[2].ViewerVote != models.ViewerVoteLike {
		t.Fatalf("public post viewer vote = %q, want like", selectedPosts[2].ViewerVote)
	}

	unselectedPosts, err := repo.ListHomeFeed(ids.unselectedFollowerID, 10, 0)
	if err != nil {
		t.Fatalf("ListHomeFeed unselected follower returned error: %v", err)
	}
	assertPostIDs(t, unselectedPosts, ids.almostPrivatePostID, ids.publicPostID)

	nonFollowerPosts, err := repo.ListHomeFeed(ids.nonFollowerID, 10, 0)
	if err != nil {
		t.Fatalf("ListHomeFeed non-follower returned error: %v", err)
	}
	assertPostIDs(t, nonFollowerPosts, ids.publicPostID)

	ownerPosts, err := repo.ListHomeFeed(ids.authorID, 10, 0)
	if err != nil {
		t.Fatalf("ListHomeFeed owner returned error: %v", err)
	}
	assertPostIDs(t, ownerPosts, ids.privatePostID, ids.almostPrivatePostID, ids.publicPostID)
}

func TestPostRepositoryProfileFeedAppliesPostPrivacy(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedFeedPrivacyRows(t, db)
	repo := NewPostRepository(db)

	selectedPosts, err := repo.ListProfilePosts(ids.authorID, ids.selectedFollowerID, 10, 0)
	if err != nil {
		t.Fatalf("ListProfilePosts selected follower returned error: %v", err)
	}
	assertPostIDs(t, selectedPosts, ids.privatePostID, ids.almostPrivatePostID, ids.publicPostID)

	nonFollowerPosts, err := repo.ListProfilePosts(ids.authorID, ids.nonFollowerID, 10, 0)
	if err != nil {
		t.Fatalf("ListProfilePosts non-follower returned error: %v", err)
	}
	assertPostIDs(t, nonFollowerPosts, ids.publicPostID)
}

func TestPostRepositoryGroupFeedReturnsOnlyGroupPosts(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedFeedPrivacyRows(t, db)
	repo := NewPostRepository(db)
	members := NewGroupMembershipRepository(db)

	accepted, err := members.IsAcceptedGroupMember(ids.groupID, ids.groupMemberID)
	if err != nil {
		t.Fatalf("IsAcceptedGroupMember accepted returned error: %v", err)
	}
	if !accepted {
		t.Fatal("expected group member to be accepted")
	}
	notAccepted, err := members.IsAcceptedGroupMember(ids.groupID, ids.nonFollowerID)
	if err != nil {
		t.Fatalf("IsAcceptedGroupMember non-member returned error: %v", err)
	}
	if notAccepted {
		t.Fatal("expected non-member to be rejected")
	}

	posts, err := repo.ListGroupFeed(ids.groupID, ids.groupMemberID, 10, 0)
	if err != nil {
		t.Fatalf("ListGroupFeed returned error: %v", err)
	}
	assertPostIDs(t, posts, ids.groupPostID)
}

type postCommentSeedIDs struct {
	authorID        uuid.UUID
	viewerID        uuid.UUID
	otherViewerID   uuid.UUID
	postID          uuid.UUID
	parentCommentID uuid.UUID
	replyCommentID  uuid.UUID
}

type feedPrivacySeedIDs struct {
	authorID             uuid.UUID
	selectedFollowerID   uuid.UUID
	unselectedFollowerID uuid.UUID
	nonFollowerID        uuid.UUID
	groupMemberID        uuid.UUID
	groupID              uuid.UUID
	publicPostID         uuid.UUID
	almostPrivatePostID  uuid.UUID
	privatePostID        uuid.UUID
	groupPostID          uuid.UUID
}

func newPostCommentTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := dbpkg.InitDB(filepath.Join(t.TempDir(), "repository.db"), filepath.Join("..", "db", "migrations"))
	if err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func seedPostCommentTestRows(t *testing.T, db *sql.DB) postCommentSeedIDs {
	t.Helper()

	ids := postCommentSeedIDs{
		authorID:        uuid.Must(uuid.FromString("6f5d9a18-5c4f-4b7a-9e9a-7a5d2efc44b1")),
		viewerID:        uuid.Must(uuid.FromString("0dd6e443-0998-4f50-a4cf-1a40a0536213")),
		otherViewerID:   uuid.Must(uuid.FromString("d54271d8-e248-4260-acf2-b6f1264830af")),
		postID:          uuid.Must(uuid.FromString("8a6de4c1-2f4a-4c52-a38f-61d76c9e7d11")),
		parentCommentID: uuid.Must(uuid.FromString("623b1e60-babc-44db-a4a1-0c3d071a80a3")),
		replyCommentID:  uuid.Must(uuid.FromString("a1ac5c54-6cb6-41bc-a88f-55ad4fb3a6d2")),
	}
	now := time.Date(2026, 6, 16, 8, 15, 0, 0, time.UTC)

	insertUser(t, db, ids.authorID, "amina@example.com", "Amina", "Njeri")
	insertUser(t, db, ids.viewerID, "viewer@example.com", "Vera", "Viewer")
	insertUser(t, db, ids.otherViewerID, "other@example.com", "Otis", "Other")

	_, err := db.Exec(`
		INSERT INTO posts (
			id, user_id, content, image_url, privacy,
			comment_count, like_count, dislike_count, created_at
		)
		VALUES (?, ?, ?, ?, 'public', 1, 1, 0, ?)
	`, ids.postID.String(), ids.authorID.String(), "First hike", "/uploads/images/hike.gif", now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert post: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO post_votes (post_id, user_id, vote, created_at)
		VALUES (?, ?, 'like', ?)
	`, ids.postID.String(), ids.viewerID.String(), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert post vote: %v", err)
	}

	deletedAt := now.Add(5 * time.Minute)
	_, err = db.Exec(`
		INSERT INTO comments (
			id, post_id, user_id, content, like_count, dislike_count, created_at, deleted_at
		)
		VALUES (?, ?, ?, ?, 0, 0, ?, ?)
	`, ids.parentCommentID.String(), ids.postID.String(), ids.authorID.String(), "deleted retained text", now.Add(time.Minute).Format(time.RFC3339), deletedAt.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert parent comment: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO comments (
			id, post_id, user_id, parent_comment_id, content, like_count, dislike_count, created_at
		)
		VALUES (?, ?, ?, ?, ?, 0, 1, ?)
	`, ids.replyCommentID.String(), ids.postID.String(), ids.authorID.String(), ids.parentCommentID.String(), "That view is unreal.", now.Add(2*time.Minute).Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert reply comment: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO comment_votes (comment_id, user_id, vote, created_at)
		VALUES (?, ?, 'dislike', ?)
	`, ids.replyCommentID.String(), ids.viewerID.String(), now.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert comment vote: %v", err)
	}

	return ids
}

func seedFeedPrivacyRows(t *testing.T, db *sql.DB) feedPrivacySeedIDs {
	t.Helper()

	ids := feedPrivacySeedIDs{
		authorID:             uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001")),
		selectedFollowerID:   uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002")),
		unselectedFollowerID: uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003")),
		nonFollowerID:        uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000004")),
		groupMemberID:        uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000005")),
		groupID:              uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001")),
		publicPostID:         uuid.Must(uuid.FromString("aaaaaaaa-0000-0000-0000-000000000001")),
		almostPrivatePostID:  uuid.Must(uuid.FromString("bbbbbbbb-0000-0000-0000-000000000001")),
		privatePostID:        uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001")),
		groupPostID:          uuid.Must(uuid.FromString("dddddddd-0000-0000-0000-000000000001")),
	}
	now := time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC)

	insertUser(t, db, ids.authorID, "author@example.com", "Author", "One")
	insertUser(t, db, ids.selectedFollowerID, "selected@example.com", "Selected", "Follower")
	insertUser(t, db, ids.unselectedFollowerID, "unselected@example.com", "Unselected", "Follower")
	insertUser(t, db, ids.nonFollowerID, "nonfollower@example.com", "Non", "Follower")
	insertUser(t, db, ids.groupMemberID, "member@example.com", "Group", "Member")

	insertFollow(t, db, ids.selectedFollowerID, ids.authorID, models.Accepted)
	insertFollow(t, db, ids.unselectedFollowerID, ids.authorID, models.Accepted)
	insertPostRow(t, db, ids.publicPostID, ids.authorID, nil, models.PostPrivacyPublic, now)
	insertPostRow(t, db, ids.almostPrivatePostID, ids.authorID, nil, models.PostPrivacyAlmostPrivate, now)
	insertPostRow(t, db, ids.privatePostID, ids.authorID, nil, models.PostPrivacyPrivate, now)
	insertPostAudienceRow(t, db, ids.privatePostID, ids.selectedFollowerID)

	_, err := db.Exec(
		`INSERT INTO post_votes (post_id, user_id, vote, created_at) VALUES (?, ?, 'like', ?)`,
		ids.publicPostID.String(),
		ids.selectedFollowerID.String(),
		now.Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("insert feed post vote: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO groups (id, creator_id, title, created_at) VALUES (?, ?, 'Hikers', ?)`,
		ids.groupID.String(),
		ids.authorID.String(),
		now.Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("insert group: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, 'accepted')`,
		ids.groupID.String(),
		ids.groupMemberID.String(),
	)
	if err != nil {
		t.Fatalf("insert group member: %v", err)
	}
	insertPostRow(t, db, ids.groupPostID, ids.authorID, &ids.groupID, models.PostPrivacyPublic, now.Add(time.Hour))

	return ids
}

func insertUser(t *testing.T, db *sql.DB, id uuid.UUID, email, firstName, lastName string) {
	t.Helper()
	_, err := db.Exec(`
		INSERT INTO users (
			id, email, password_hash, first_name, last_name, dob,
			avatar, nickname, about_me, is_public, created_at
		)
		VALUES (?, ?, 'hash', ?, ?, '1998-04-12', '/uploads/avatars/user.png', 'nick', NULL, 1, ?)
	`, id.String(), email, firstName, lastName, time.Date(2026, 6, 16, 8, 0, 0, 0, time.UTC).Format(time.RFC3339))
	if err != nil {
		t.Fatalf("insert user %s: %v", id, err)
	}
}

func insertFollow(t *testing.T, db *sql.DB, followerID, followeeID uuid.UUID, status models.Status) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO followers (follower_id, followee_id, status, created_at) VALUES (?, ?, ?, ?)`,
		followerID.String(),
		followeeID.String(),
		string(status),
		time.Date(2026, 6, 16, 8, 0, 0, 0, time.UTC).Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("insert follow: %v", err)
	}
}

func insertPostRow(t *testing.T, db *sql.DB, id, authorID uuid.UUID, groupID *uuid.UUID, privacy models.PostPrivacy, createdAt time.Time) {
	t.Helper()
	var groupValue any
	if groupID != nil {
		groupValue = groupID.String()
	}
	_, err := db.Exec(
		`INSERT INTO posts (id, user_id, group_id, content, privacy, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		id.String(),
		authorID.String(),
		groupValue,
		"feed post",
		string(privacy),
		createdAt.Format(time.RFC3339),
	)
	if err != nil {
		t.Fatalf("insert post %s: %v", id, err)
	}
}

func insertPostAudienceRow(t *testing.T, db *sql.DB, postID, userID uuid.UUID) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO post_audiences (post_id, user_id) VALUES (?, ?)`, postID.String(), userID.String())
	if err != nil {
		t.Fatalf("insert post audience: %v", err)
	}
}

func assertPostIDs(t *testing.T, posts []*models.PostWithAuthor, expected ...uuid.UUID) {
	t.Helper()
	if len(posts) != len(expected) {
		t.Fatalf("post count = %d, want %d", len(posts), len(expected))
	}
	for i, post := range posts {
		if post.Post.ID != expected[i] {
			t.Fatalf("post[%d] = %s, want %s", i, post.Post.ID, expected[i])
		}
	}
}

func TestCommentRepositoryCreateCommentAtomic(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)

	commentRepo := NewCommentRepository(db)
	postRepo := NewPostRepository(db)

	postBefore, err := postRepo.GetPostByID(ids.postID, ids.viewerID)
	if err != nil {
		t.Fatalf("failed to fetch post: %v", err)
	}
	initialCount := postBefore.Post.CommentCount

	newCommentID := uuid.Must(uuid.NewV4())
	newComment := &models.Comment{
		ID:              newCommentID,
		PostID:          ids.postID,
		UserID:          &ids.authorID,
		ParentCommentID: nil,
		Content:         "Brand new comment",
		CreatedAt:       time.Now(),
	}

	err = commentRepo.CreateComment(newComment)
	if err != nil {
		t.Fatalf("CreateComment failed: %v", err)
	}

	postAfter, err := postRepo.GetPostByID(ids.postID, ids.viewerID)
	if err != nil {
		t.Fatalf("failed to fetch post after comment creation: %v", err)
	}

	expectedCount := initialCount + 1
	if postAfter.Post.CommentCount != expectedCount {
		t.Fatalf("post comment count = %d, want %d", postAfter.Post.CommentCount, expectedCount)
	}

	insertedComment, err := commentRepo.GetCommentByID(newCommentID, ids.viewerID)
	if err != nil {
		t.Fatalf("failed to fetch inserted comment: %v", err)
	}
	if insertedComment.Comment.Content != "Brand new comment" {
		t.Fatalf("comment content = %q, want %q", insertedComment.Comment.Content, "Brand new comment")
	}
}

func TestPostRepositoryUpdatePostWithAudience(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)

	repo := NewPostRepository(db)
	audRepo := NewPostAudienceRepository(db)

	postRow, err := repo.GetPostByID(ids.postID, ids.viewerID)
	if err != nil {
		t.Fatalf("failed to fetch post: %v", err)
	}

	post := postRow.Post
	post.Content = "Updated content"
	post.Privacy = models.PostPrivacyPrivate
	now := time.Now()
	post.UpdatedAt = &now

	newAudienceMember := ids.otherViewerID
	err = repo.UpdatePostWithAudience(&post, []uuid.UUID{newAudienceMember})
	if err != nil {
		t.Fatalf("UpdatePostWithAudience failed: %v", err)
	}

	updatedRow, err := repo.GetPostByID(ids.postID, ids.viewerID)
	if err != nil {
		t.Fatalf("failed to fetch updated post: %v", err)
	}

	if updatedRow.Post.Content != "Updated content" {
		t.Fatalf("content = %q, want %q", updatedRow.Post.Content, "Updated content")
	}
	if updatedRow.Post.Privacy != models.PostPrivacyPrivate {
		t.Fatalf("privacy = %q, want %q", updatedRow.Post.Privacy, models.PostPrivacyPrivate)
	}

	audience, err := audRepo.ListPostAudience(ids.postID)
	if err != nil {
		t.Fatalf("ListPostAudience failed: %v", err)
	}
	if len(audience) != 1 || audience[0] != newAudienceMember {
		t.Fatalf("audience = %#v, want only %s", audience, newAudienceMember)
	}
}

func TestPostRepositoryDeletePost(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)

	repo := NewPostRepository(db)
	err := repo.DeletePost(ids.postID)
	if err != nil {
		t.Fatalf("DeletePost failed: %v", err)
	}

	deletedRow, err := repo.GetPostByID(ids.postID, ids.viewerID)
	if err != nil {
		t.Fatalf("GetPostByID failed after deletion: %v", err)
	}

	if deletedRow.Post.DeletedAt == nil {
		t.Fatal("expected deleted_at to be set")
	}
	if deletedRow.Post.Content != "" {
		t.Fatalf("content = %q, want empty", deletedRow.Post.Content)
	}
	if deletedRow.Post.ImageURL != nil {
		t.Fatalf("image_url = %v, want nil", deletedRow.Post.ImageURL)
	}
}

func TestCommentRepositoryUpdateComment(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)

	repo := NewCommentRepository(db)
	commentRow, err := repo.GetCommentByID(ids.parentCommentID, ids.viewerID)
	if err != nil {
		t.Fatalf("GetCommentByID failed: %v", err)
	}

	comment := commentRow.Comment
	comment.Content = "Updated comment content"
	now := time.Now()
	comment.UpdatedAt = &now

	err = repo.UpdateComment(&comment)
	if err != nil {
		t.Fatalf("UpdateComment failed: %v", err)
	}

	updatedRow, err := repo.GetCommentByID(ids.parentCommentID, ids.viewerID)
	if err != nil {
		t.Fatalf("failed to fetch updated comment: %v", err)
	}

	if updatedRow.Comment.Content != "Updated comment content" {
		t.Fatalf("content = %q, want %q", updatedRow.Comment.Content, "Updated comment content")
	}
	if updatedRow.Comment.UpdatedAt == nil {
		t.Fatal("expected updated_at to be set")
	}
}

func TestCommentRepositoryDeleteComment(t *testing.T) {
	db := newPostCommentTestDB(t)
	ids := seedPostCommentTestRows(t, db)

	repo := NewCommentRepository(db)
	now := time.Now()
	err := repo.DeleteComment(ids.parentCommentID, now)
	if err != nil {
		t.Fatalf("DeleteComment failed: %v", err)
	}

	deletedRow, err := repo.GetCommentByID(ids.parentCommentID, ids.viewerID)
	if err != nil {
		t.Fatalf("GetCommentByID failed after deletion: %v", err)
	}

	if deletedRow.Comment.DeletedAt == nil {
		t.Fatal("expected deleted_at to be set")
	}
	if deletedRow.Comment.Content != "" {
		t.Fatalf("content = %q, want empty", deletedRow.Comment.Content)
	}
	if deletedRow.Comment.ImageURL != nil {
		t.Fatalf("image_url = %v, want nil", deletedRow.Comment.ImageURL)
	}
}
