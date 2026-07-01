package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/dto"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
)

func TestPostServiceHomeFeedPaginationDefaultsAndHasMore(t *testing.T) {
	posts := newFakePostRepository()
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	posts.homeRows = makePostRows(t, 21)
	service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

	response, err := service.GetHomeFeed(viewerID, 0, 0)
	if err != nil {
		t.Fatalf("GetHomeFeed returned error: %v", err)
	}
	if posts.lastLimit != DefaultFeedLimit+1 {
		t.Fatalf("repo limit = %d, want %d", posts.lastLimit, DefaultFeedLimit+1)
	}
	if len(response.Data) != DefaultFeedLimit {
		t.Fatalf("response post count = %d, want %d", len(response.Data), DefaultFeedLimit)
	}
	if !response.Pagination.HasMore {
		t.Fatal("expected has_more=true when repository returns limit+1 rows")
	}
	if response.Pagination.Limit != DefaultFeedLimit || response.Pagination.Offset != 0 {
		t.Fatalf("pagination = %#v, want default limit and zero offset", response.Pagination)
	}
}

func TestPostServiceRejectsInvalidPagination(t *testing.T) {
	service := newTestPostService(newFakePostRepository(), newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))

	tests := []struct {
		name   string
		limit  int
		offset int
	}{
		{name: "limit too small", limit: -1, offset: 0},
		{name: "limit too large", limit: MaxFeedLimit + 1, offset: 0},
		{name: "offset negative", limit: 20, offset: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetHomeFeed(viewerID, tt.limit, tt.offset)
			if !errors.Is(err, ErrInvalidPagination) {
				t.Fatalf("error = %v, want ErrInvalidPagination", err)
			}
		})
	}
}

func TestPostServiceProfileVisibility(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	profileID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))

	tests := []struct {
		name      string
		profile   *models.User
		status    models.Status
		wantError error
	}{
		{
			name:    "public profile allowed",
			profile: &models.User{ID: profileID, Email: "public@example.com", IsPublic: true},
		},
		{
			name:    "accepted follower allowed",
			profile: &models.User{ID: profileID, Email: "private@example.com", IsPublic: false},
			status:  models.Accepted,
		},
		{
			name:      "non follower forbidden",
			profile:   &models.User{ID: profileID, Email: "private@example.com", IsPublic: false},
			status:    "none",
			wantError: ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := newFakeUserRepository()
			users.add(tt.profile)
			followers := newFakeFollowersRepository()
			if tt.status != "" && tt.status != "none" {
				followers.status[followerKey{followerID: viewerID, followeeID: profileID}] = tt.status
			}
			posts := newFakePostRepository()
			posts.profileRows = makePostRows(t, 1)
			service := newTestPostService(posts, users, followers, newFakeGroupMembershipRepository())

			_, err := service.GetProfilePosts(profileID, viewerID, 20, 0)
			if !errors.Is(err, tt.wantError) {
				t.Fatalf("error = %v, want %v", err, tt.wantError)
			}
		})
	}
}

func TestPostServiceProfileOwnerCanViewOwnPrivateProfile(t *testing.T) {
	userID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	posts := newFakePostRepository()
	posts.profileRows = makePostRows(t, 1)
	service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

	if _, err := service.GetProfilePosts(userID, userID, 20, 0); err != nil {
		t.Fatalf("GetProfilePosts owner returned error: %v", err)
	}
}

func TestPostServiceGroupFeedRequiresAcceptedMembership(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))
	groups := newFakeGroupMembershipRepository()
	service := newTestPostService(newFakePostRepository(), newFakeUserRepository(), newFakeFollowersRepository(), groups)

	_, err := service.GetGroupFeed(groupID, viewerID, 20, 0)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("error = %v, want ErrForbidden", err)
	}

	groups.accepted[groupMemberKey{groupID: groupID, userID: viewerID}] = true
	if _, err := service.GetGroupFeed(groupID, viewerID, 20, 0); err != nil {
		t.Fatalf("GetGroupFeed accepted member returned error: %v", err)
	}
}

func TestPostServiceGetSinglePostMapsPublicPost(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	postID := uuid.Must(uuid.FromString("aaaaaaaa-0000-0000-0000-000000000001"))
	posts := newFakePostRepository()
	posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, nil)
	service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())
	viewer := viewerID.String()

	response, err := service.GetSinglePost(context.Background(), postID.String(), &viewer)
	if err != nil {
		t.Fatalf("GetSinglePost returned error: %v", err)
	}
	active, ok := response.(*dto.ActivePostResponse)
	if !ok {
		t.Fatalf("response type = %T, want active post", response)
	}
	if active.ID != postID || active.ViewerVote != models.ViewerVoteNone {
		t.Fatalf("active response = %#v", active)
	}
	if posts.lastSingleID != postID || posts.lastSingleViewerID != viewerID {
		t.Fatalf("repo ids = %s/%s, want %s/%s", posts.lastSingleID, posts.lastSingleViewerID, postID, viewerID)
	}
}

func TestPostServiceGetSinglePostEnforcesAlmostPrivateFollowers(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000009"))
	postID := uuid.Must(uuid.FromString("bbbbbbbb-0000-0000-0000-000000000001"))
	posts := newFakePostRepository()
	posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyAlmostPrivate, &authorID)
	followers := newFakeFollowersRepository()
	service := newTestPostService(posts, newFakeUserRepository(), followers, newFakeGroupMembershipRepository())
	viewer := viewerID.String()

	if _, err := service.GetSinglePost(context.Background(), postID.String(), &viewer); !errors.Is(err, ErrPostForbidden) {
		t.Fatalf("error = %v, want ErrPostForbidden", err)
	}

	followers.status[followerKey{followerID: viewerID, followeeID: authorID}] = models.Accepted
	if _, err := service.GetSinglePost(context.Background(), postID.String(), &viewer); err != nil {
		t.Fatalf("GetSinglePost accepted follower returned error: %v", err)
	}
}

func TestPostServiceGetSinglePostRejectsPrivatePostUnlessOwner(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000009"))
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	postID := uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001"))
	posts := newFakePostRepository()
	posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPrivate, &authorID)
	service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())
	viewer := viewerID.String()

	if _, err := service.GetSinglePost(context.Background(), postID.String(), &viewer); !errors.Is(err, ErrPostForbidden) {
		t.Fatalf("error = %v, want ErrPostForbidden", err)
	}

	owner := authorID.String()
	if _, err := service.GetSinglePost(context.Background(), postID.String(), &owner); err != nil {
		t.Fatalf("GetSinglePost owner returned error: %v", err)
	}
}

func TestPostServiceGetSinglePostAllowsPrivatePostIfAudienceMember(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000009"))
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	nonAudienceID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	postID := uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001"))

	posts := newFakePostRepository()
	posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPrivate, &authorID)
	posts.audienceMembers[viewerID] = true // viewer is in the audience

	service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

	viewer := viewerID.String()
	if _, err := service.GetSinglePost(context.Background(), postID.String(), &viewer); err != nil {
		t.Fatalf("GetSinglePost audience member returned error: %v", err)
	}

	nonAudience := nonAudienceID.String()
	if _, err := service.GetSinglePost(context.Background(), postID.String(), &nonAudience); !errors.Is(err, ErrPostForbidden) {
		t.Fatalf("error = %v, want ErrPostForbidden", err)
	}
}

func TestPostServiceGetSinglePostRequiresGroupMembership(t *testing.T) {
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	groupID := uuid.Must(uuid.FromString("20000000-0000-0000-0000-000000000001"))
	postID := uuid.Must(uuid.FromString("dddddddd-0000-0000-0000-000000000001"))
	posts := newFakePostRepository()
	posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, nil)
	posts.singleRow.Post.GroupID = &groupID
	groups := newFakeGroupMembershipRepository()
	service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), groups)
	viewer := viewerID.String()

	if _, err := service.GetSinglePost(context.Background(), postID.String(), &viewer); !errors.Is(err, ErrPostForbidden) {
		t.Fatalf("error = %v, want ErrPostForbidden", err)
	}

	groups.accepted[groupMemberKey{groupID: groupID, userID: viewerID}] = true
	if _, err := service.GetSinglePost(context.Background(), postID.String(), &viewer); err != nil {
		t.Fatalf("GetSinglePost accepted group member returned error: %v", err)
	}
}

type fakePostRepository struct {
	homeRows           []*models.PostWithAuthor
	profileRows        []*models.PostWithAuthor
	groupRows          []*models.PostWithAuthor
	singleRow          *models.PostWithAuthor
	singleErr          error
	lastLimit          int
	lastOffset         int
	lastSingleID       uuid.UUID
	lastSingleViewerID uuid.UUID
	audienceMembers    map[uuid.UUID]bool
}

func newFakePostRepository() *fakePostRepository {
	return &fakePostRepository{
		audienceMembers: make(map[uuid.UUID]bool),
	}
}

func (r *fakePostRepository) CreatePost(post *models.Post) error {
	return nil
}

func (r *fakePostRepository) CreatePostWithAudience(post *models.Post, audience []uuid.UUID) error {
	return nil
}

func (r *fakePostRepository) ReplacePostAudience(postID uuid.UUID, userIDs []uuid.UUID) error {
	return nil
}

func (r *fakePostRepository) ListPostAudience(postID uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (r *fakePostRepository) IsPostAudienceMember(postID, userID uuid.UUID) (bool, error) {
	return r.audienceMembers[userID], nil
}

func (r *fakePostRepository) GetPostByID(id, viewerID uuid.UUID) (*models.PostWithAuthor, error) {
	r.lastSingleID = id
	r.lastSingleViewerID = viewerID
	if r.singleErr != nil {
		return nil, r.singleErr
	}
	if r.singleRow == nil {
		return nil, errors.New("post not found")
	}
	return r.singleRow, nil
}

func (r *fakePostRepository) ListPosts(query models.PostQuery, viewerID uuid.UUID) ([]*models.PostWithAuthor, error) {
	return nil, errors.New("not implemented")
}

func (r *fakePostRepository) ListHomeFeed(viewerID uuid.UUID, limit, offset int) ([]*models.PostWithAuthor, error) {
	r.lastLimit = limit
	r.lastOffset = offset
	return r.homeRows, nil
}

func (r *fakePostRepository) ListProfilePosts(profileUserID, viewerID uuid.UUID, limit, offset int) ([]*models.PostWithAuthor, error) {
	r.lastLimit = limit
	r.lastOffset = offset
	return r.profileRows, nil
}

func (r *fakePostRepository) ListGroupFeed(groupID, viewerID uuid.UUID, limit, offset int) ([]*models.PostWithAuthor, error) {
	r.lastLimit = limit
	r.lastOffset = offset
	return r.groupRows, nil
}

func (r *fakePostRepository) UpdatePostWithAudience(post *models.Post, audience []uuid.UUID) error {
	if r.singleRow != nil && r.singleRow.Post.ID == post.ID {
		r.singleRow.Post.Content = post.Content
		r.singleRow.Post.Privacy = post.Privacy
		r.singleRow.Post.ImageURL = post.ImageURL
		r.singleRow.Post.UpdatedAt = post.UpdatedAt
	}
	return nil
}

func (r *fakePostRepository) DeletePost(id uuid.UUID) error {
	if r.singleRow != nil && r.singleRow.Post.ID == id {
		now := time.Now()
		r.singleRow.Post.DeletedAt = &now
		r.singleRow.Post.Content = ""
		r.singleRow.Post.ImageURL = nil
	}
	return nil
}

type groupMemberKey struct {
	groupID uuid.UUID
	userID  uuid.UUID
}

type fakeGroupMembershipRepository struct {
	accepted map[groupMemberKey]bool
}

func newFakeGroupMembershipRepository() *fakeGroupMembershipRepository {
	return &fakeGroupMembershipRepository{accepted: map[groupMemberKey]bool{}}
}

func (r *fakeGroupMembershipRepository) IsAcceptedGroupMember(groupID, userID uuid.UUID) (bool, error) {
	return r.accepted[groupMemberKey{groupID: groupID, userID: userID}], nil
}

func (r *fakeGroupMembershipRepository) GetMembership(groupID, userID uuid.UUID) (string, error) {
	if r.accepted[groupMemberKey{groupID: groupID, userID: userID}] {
		return "accepted", nil
	}
	return "none", nil
}

func (r *fakeGroupMembershipRepository) AddMembership(groupID, userID uuid.UUID, status string) error {
	if status == "accepted" {
		r.accepted[groupMemberKey{groupID: groupID, userID: userID}] = true
	}
	return nil
}

func (r *fakeGroupMembershipRepository) UpdateMembershipStatus(groupID, userID uuid.UUID, status string) error {
	if status == "accepted" {
		r.accepted[groupMemberKey{groupID: groupID, userID: userID}] = true
	} else {
		r.accepted[groupMemberKey{groupID: groupID, userID: userID}] = false
	}
	return nil
}

func (r *fakeGroupMembershipRepository) RemoveMembership(groupID, userID uuid.UUID) error {
	delete(r.accepted, groupMemberKey{groupID: groupID, userID: userID})
	return nil
}

func (r *fakeGroupMembershipRepository) ListGroupMembers(groupID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}

func (r *fakeGroupMembershipRepository) ListPendingRequests(groupID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}

func makePostRows(t *testing.T, count int) []*models.PostWithAuthor {
	t.Helper()
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000009"))
	rows := make([]*models.PostWithAuthor, 0, count)
	for i := 0; i < count; i++ {
		postID, err := uuid.NewV4()
		if err != nil {
			t.Fatalf("uuid.NewV4 returned error: %v", err)
		}
		rows = append(rows, &models.PostWithAuthor{
			Post: models.Post{
				ID:        postID,
				UserID:    &authorID,
				Content:   "post",
				Privacy:   models.PostPrivacyPublic,
				CreatedAt: time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
			},
			Author: &models.PublicUser{
				ID:        authorID,
				FirstName: "Amina",
				LastName:  "Njeri",
			},
			ViewerVote: models.ViewerVoteNone,
		})
	}
	return rows
}

func makeSinglePostRow(t *testing.T, postID uuid.UUID, privacy models.PostPrivacy, authorID *uuid.UUID) *models.PostWithAuthor {
	t.Helper()
	if authorID == nil {
		defaultAuthorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000009"))
		authorID = &defaultAuthorID
	}
	return &models.PostWithAuthor{
		Post: models.Post{
			ID:        postID,
			UserID:    authorID,
			Content:   "post",
			Privacy:   privacy,
			CreatedAt: time.Date(2026, 6, 16, 12, 0, 0, 0, time.UTC),
		},
		Author: &models.PublicUser{
			ID:        *authorID,
			FirstName: "Amina",
			LastName:  "Njeri",
		},
		ViewerVote: models.ViewerVoteNone,
	}
}

func TestPostServiceCreatePost(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	followerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	nonFollowerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	groupID := uuid.Must(uuid.FromString("30000000-0000-0000-0000-000000000001"))

	t.Run("Create public post success", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, uuid.Nil, models.PostPrivacyPublic, &authorID)
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

		req := &dto.CreatePostRequest{
			Content: "Hello world",
			Privacy: models.PostPrivacyPublic,
		}
		resp, err := service.CreatePost(context.Background(), req, authorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		activePost, ok := resp.(*dto.ActivePostResponse)
		if !ok {
			t.Fatalf("expected ActivePostResponse, got %T", resp)
		}
		if activePost.Content != "post" {
			t.Fatalf("unexpected content: %s", activePost.Content)
		}
	})

	t.Run("Create group post fails if not a member", func(t *testing.T) {
		posts := newFakePostRepository()
		groups := newFakeGroupMembershipRepository()
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), groups)

		req := &dto.CreatePostRequest{
			Content: "Hello group",
			GroupID: &groupID,
		}
		_, err := service.CreatePost(context.Background(), req, authorID)
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("Create group post success if member", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, uuid.Nil, models.PostPrivacyPublic, &authorID)
		groups := newFakeGroupMembershipRepository()
		groups.accepted[groupMemberKey{groupID: groupID, userID: authorID}] = true
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), groups)

		req := &dto.CreatePostRequest{
			Content: "Hello group",
			GroupID: &groupID,
		}
		_, err := service.CreatePost(context.Background(), req, authorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Create private post fails if audience is not accepted follower", func(t *testing.T) {
		posts := newFakePostRepository()
		followers := newFakeFollowersRepository()
		service := newTestPostService(posts, newFakeUserRepository(), followers, newFakeGroupMembershipRepository())

		req := &dto.CreatePostRequest{
			Content:     "Hello private",
			Privacy:     models.PostPrivacyPrivate,
			AudienceIDs: []uuid.UUID{nonFollowerID},
		}
		_, err := service.CreatePost(context.Background(), req, authorID)
		if !errors.Is(err, ErrNotFollower) {
			t.Fatalf("expected ErrNotFollower, got %v", err)
		}
	})

	t.Run("Create private post success if audience is accepted follower", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, uuid.Nil, models.PostPrivacyPrivate, &authorID)
		followers := newFakeFollowersRepository()
		followers.status[followerKey{followerID: followerID, followeeID: authorID}] = models.Accepted
		service := newTestPostService(posts, newFakeUserRepository(), followers, newFakeGroupMembershipRepository())

		req := &dto.CreatePostRequest{
			Content:     "Hello private",
			Privacy:     models.PostPrivacyPrivate,
			AudienceIDs: []uuid.UUID{followerID},
		}
		_, err := service.CreatePost(context.Background(), req, authorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

type fakeCommentRepository struct {
	comments    []*models.CommentWithAuthor
	commentsMap map[uuid.UUID]*models.CommentWithAuthor
	err         error
}

func (f *fakeCommentRepository) CreateComment(comment *models.Comment) error {
	return nil
}

func (f *fakeCommentRepository) GetCommentByID(id, viewerID uuid.UUID) (*models.CommentWithAuthor, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.commentsMap != nil {
		if c, ok := f.commentsMap[id]; ok {
			return c, nil
		}
	}
	return nil, errors.New("comment not found")
}

func (f *fakeCommentRepository) ListCommentTreeByPost(postID, viewerID uuid.UUID, limit, offset int) ([]*models.CommentWithAuthor, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.comments, nil
}

func (f *fakeCommentRepository) UpdateComment(comment *models.Comment) error {
	if f.err != nil {
		return f.err
	}
	if f.commentsMap != nil {
		if c, ok := f.commentsMap[comment.ID]; ok {
			c.Comment.Content = comment.Content
			c.Comment.ImageURL = comment.ImageURL
			c.Comment.UpdatedAt = comment.UpdatedAt
			return nil
		}
	}
	return errors.New("comment not found")
}

func (f *fakeCommentRepository) DeleteComment(id uuid.UUID, deletedAt time.Time) error {
	if f.err != nil {
		return f.err
	}
	if f.commentsMap != nil {
		if c, ok := f.commentsMap[id]; ok {
			c.Comment.DeletedAt = &deletedAt
			c.Comment.Content = ""
			c.Comment.ImageURL = nil
			return nil
		}
	}
	return errors.New("comment not found")
}

func newFakeCommentRepository() *fakeCommentRepository {
	return &fakeCommentRepository{}
}

func newTestPostService(
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	followerRepo repositories.FollowersRepository,
	groupMemberRepo repositories.GroupMembershipRepository,
) PostService {
	return NewPostService(postRepo, userRepo, followerRepo, groupMemberRepo, newFakeCommentRepository())
}

func TestPostServiceGetCommentsByPost(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	viewerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	strangerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	postID := uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001"))
	commentID := uuid.Must(uuid.FromString("11111111-0000-0000-0000-000000000001"))

	t.Run("Get comments fails if not authorized to view post", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyAlmostPrivate, &authorID)
		followers := newFakeFollowersRepository()
		comments := newFakeCommentRepository()
		service := NewPostService(posts, newFakeUserRepository(), followers, newFakeGroupMembershipRepository(), comments)

		_, err := service.GetCommentsByPost(context.Background(), postID.String(), strangerID, 10, 0)
		if !errors.Is(err, ErrPostForbidden) {
			t.Fatalf("expected ErrPostForbidden, got %v", err)
		}
	})

	t.Run("Get comments success if authorized", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyAlmostPrivate, &authorID)
		followers := newFakeFollowersRepository()
		followers.status[followerKey{followerID: viewerID, followeeID: authorID}] = models.Accepted
		comments := newFakeCommentRepository()
		comments.comments = []*models.CommentWithAuthor{
			{
				Comment: models.Comment{
					ID:        commentID,
					PostID:    postID,
					UserID:    &authorID,
					Content:   "First comment",
					CreatedAt: time.Now(),
				},
				Author: &models.PublicUser{
					ID:        authorID,
					FirstName: "Amina",
					LastName:  "Njeri",
				},
				ViewerVote: models.ViewerVoteNone,
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), followers, newFakeGroupMembershipRepository(), comments)

		resp, err := service.GetCommentsByPost(context.Background(), postID.String(), viewerID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Fatalf("expected 1 comment, got %d", len(resp.Data))
		}
		activeComment, ok := resp.Data[0].(*dto.ActiveCommentResponse)
		if !ok {
			t.Fatalf("expected ActiveCommentResponse, got %T", resp.Data[0])
		}
		if activeComment.Content != "First comment" {
			t.Fatalf("unexpected content: %s", activeComment.Content)
		}
	})

	t.Run("Get comments success even if parent post is soft-deleted", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		deletedTime := time.Now()
		posts.singleRow.Post.DeletedAt = &deletedTime
		comments := newFakeCommentRepository()
		comments.comments = []*models.CommentWithAuthor{
			{
				Comment: models.Comment{
					ID:        commentID,
					PostID:    postID,
					UserID:    &authorID,
					Content:   "Historical comment",
					CreatedAt: time.Now(),
				},
				Author: &models.PublicUser{
					ID:        authorID,
					FirstName: "Amina",
					LastName:  "Njeri",
				},
				ViewerVote: models.ViewerVoteNone,
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		resp, err := service.GetCommentsByPost(context.Background(), postID.String(), viewerID, 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Fatalf("expected 1 comment, got %d", len(resp.Data))
		}
	})
}

func TestPostServiceCreateComment(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	strangerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	postID := uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001"))
	commentID := uuid.Must(uuid.FromString("11111111-0000-0000-0000-000000000001"))

	t.Run("Create comment success", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		comments := newFakeCommentRepository()
		users := newFakeUserRepository()
		users.add(&models.User{
			ID:        authorID,
			FirstName: "John",
			LastName:  "Doe",
		})
		service := NewPostService(posts, users, newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		req := &dto.CreateCommentRequest{
			PostID:  postID,
			Content: "Great post!",
		}
		resp, err := service.CreateComment(context.Background(), req, authorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		activeComment, ok := resp.(*dto.ActiveCommentResponse)
		if !ok {
			t.Fatalf("expected ActiveCommentResponse, got %T", resp)
		}
		if activeComment.Content != "Great post!" {
			t.Fatalf("unexpected content: %s", activeComment.Content)
		}
	})

	t.Run("Create comment fails if post is soft-deleted", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		deletedTime := time.Now()
		posts.singleRow.Post.DeletedAt = &deletedTime
		comments := newFakeCommentRepository()
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		req := &dto.CreateCommentRequest{
			PostID:  postID,
			Content: "Great post!",
		}
		_, err := service.CreateComment(context.Background(), req, authorID)
		if !errors.Is(err, ErrPostOrCommentDeleted) {
			t.Fatalf("expected ErrPostOrCommentDeleted, got %v", err)
		}
	})

	t.Run("Create comment fails if user has no view permission", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPrivate, &authorID)
		comments := newFakeCommentRepository()
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		req := &dto.CreateCommentRequest{
			PostID:  postID,
			Content: "Great post!",
		}
		_, err := service.CreateComment(context.Background(), req, strangerID)
		if !errors.Is(err, ErrPostForbidden) {
			t.Fatalf("expected ErrPostForbidden, got %v", err)
		}
	})

	t.Run("Create comment fails if parent comment does not exist", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		comments := newFakeCommentRepository()
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		req := &dto.CreateCommentRequest{
			PostID:          postID,
			ParentCommentID: &commentID,
			Content:         "Great post!",
		}
		_, err := service.CreateComment(context.Background(), req, authorID)
		if !errors.Is(err, ErrCommentNotFound) {
			t.Fatalf("expected ErrCommentNotFound, got %v", err)
		}
	})

	t.Run("Create comment fails if parent comment belongs to different post", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		comments := newFakeCommentRepository()
		otherPostID := uuid.Must(uuid.NewV4())
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:     commentID,
					PostID: otherPostID,
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		req := &dto.CreateCommentRequest{
			PostID:          postID,
			ParentCommentID: &commentID,
			Content:         "Great post!",
		}
		_, err := service.CreateComment(context.Background(), req, authorID)
		if !errors.Is(err, ErrCrossPostParent) {
			t.Fatalf("expected ErrCrossPostParent, got %v", err)
		}
	})

	t.Run("Create comment fails if parent comment is soft-deleted", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		comments := newFakeCommentRepository()
		deletedTime := time.Now()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:        commentID,
					PostID:    postID,
					DeletedAt: &deletedTime,
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		req := &dto.CreateCommentRequest{
			PostID:          postID,
			ParentCommentID: &commentID,
			Content:         "Great post!",
		}
		_, err := service.CreateComment(context.Background(), req, authorID)
		if !errors.Is(err, ErrPostOrCommentDeleted) {
			t.Fatalf("expected ErrPostOrCommentDeleted, got %v", err)
		}
	})
}

func TestPostServiceUpdatePost(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	strangerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	postID := uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001"))

	t.Run("Author updates post successfully", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

		content := "Updated Content"
		privacy := models.PostPrivacyPublic
		req := &dto.UpdatePostRequest{
			Content: &content,
			Privacy: &privacy,
		}

		resp, err := service.UpdatePost(context.Background(), postID.String(), req, authorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		activePost, ok := resp.(*dto.ActivePostResponse)
		if !ok {
			t.Fatalf("expected ActivePostResponse, got %T", resp)
		}
		if activePost.Content != "Updated Content" {
			t.Fatalf("content = %q, want %q", activePost.Content, "Updated Content")
		}
		if activePost.UpdatedAt == nil {
			t.Fatal("expected UpdatedAt to be set")
		}
	})

	t.Run("Non-author edit fails", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

		content := "Updated Content"
		req := &dto.UpdatePostRequest{
			Content: &content,
		}

		_, err := service.UpdatePost(context.Background(), postID.String(), req, strangerID)
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("Edit fails if post is deleted", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		deletedTime := time.Now()
		posts.singleRow.Post.DeletedAt = &deletedTime
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

		content := "Updated Content"
		req := &dto.UpdatePostRequest{
			Content: &content,
		}

		_, err := service.UpdatePost(context.Background(), postID.String(), req, authorID)
		if !errors.Is(err, ErrPostOrCommentDeleted) {
			t.Fatalf("expected ErrPostOrCommentDeleted, got %v", err)
		}
	})
}

func TestPostServiceDeletePost(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	strangerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	postID := uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001"))

	t.Run("Author deletes post successfully", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

		resp, err := service.DeletePost(context.Background(), postID.String(), authorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		deletedPost, ok := resp.(*dto.DeletedPostResponse)
		if !ok {
			t.Fatalf("expected DeletedPostResponse, got %T", resp)
		}
		if !deletedPost.Deleted {
			t.Fatal("expected Deleted = true")
		}
	})

	t.Run("Non-author delete fails", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

		_, err := service.DeletePost(context.Background(), postID.String(), strangerID)
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("Repeated delete is idempotent", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		deletedTime := time.Now()
		posts.singleRow.Post.DeletedAt = &deletedTime
		service := newTestPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository())

		resp, err := service.DeletePost(context.Background(), postID.String(), authorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		deletedPost, ok := resp.(*dto.DeletedPostResponse)
		if !ok {
			t.Fatalf("expected DeletedPostResponse, got %T", resp)
		}
		if !deletedPost.Deleted {
			t.Fatal("expected Deleted = true")
		}
	})
}

func TestPostServiceUpdateComment(t *testing.T) {
	authorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	strangerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	postID := uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001"))
	commentID := uuid.Must(uuid.FromString("11111111-0000-0000-0000-000000000001"))

	t.Run("Update comment success", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		comments := newFakeCommentRepository()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:      commentID,
					PostID:  postID,
					UserID:  &authorID,
					Content: "Old content",
				},
				Author: &models.PublicUser{ID: authorID},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		newContent := "Updated comment content"
		req := &dto.UpdateCommentRequest{
			Content: &newContent,
		}

		resp, err := service.UpdateComment(context.Background(), commentID.String(), req, authorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		activeComment, ok := resp.(*dto.ActiveCommentResponse)
		if !ok {
			t.Fatalf("expected ActiveCommentResponse, got %T", resp)
		}
		if activeComment.Content != newContent {
			t.Fatalf("content = %q, want %q", activeComment.Content, newContent)
		}
		if activeComment.UpdatedAt == nil {
			t.Fatal("expected UpdatedAt to be set")
		}
	})

	t.Run("Non-author edit fails", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		comments := newFakeCommentRepository()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:      commentID,
					PostID:  postID,
					UserID:  &authorID,
					Content: "Old content",
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		newContent := "Fail edit"
		req := &dto.UpdateCommentRequest{
			Content: &newContent,
		}

		_, err := service.UpdateComment(context.Background(), commentID.String(), req, strangerID)
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("Edit to deleted comment is rejected", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		comments := newFakeCommentRepository()
		deletedTime := time.Now()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:        commentID,
					PostID:    postID,
					UserID:    &authorID,
					Content:   "Deleted comment",
					DeletedAt: &deletedTime,
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		newContent := "Fail edit"
		req := &dto.UpdateCommentRequest{
			Content: &newContent,
		}

		_, err := service.UpdateComment(context.Background(), commentID.String(), req, authorID)
		if !errors.Is(err, ErrPostOrCommentDeleted) {
			t.Fatalf("expected ErrPostOrCommentDeleted, got %v", err)
		}
	})

	t.Run("Edit to comment under deleted post is rejected", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &authorID)
		deletedTime := time.Now()
		posts.singleRow.Post.DeletedAt = &deletedTime

		comments := newFakeCommentRepository()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:      commentID,
					PostID:  postID,
					UserID:  &authorID,
					Content: "Active comment",
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		newContent := "Fail edit"
		req := &dto.UpdateCommentRequest{
			Content: &newContent,
		}

		_, err := service.UpdateComment(context.Background(), commentID.String(), req, authorID)
		if !errors.Is(err, ErrPostOrCommentDeleted) {
			t.Fatalf("expected ErrPostOrCommentDeleted, got %v", err)
		}
	})
}

func TestPostServiceDeleteComment(t *testing.T) {
	commentAuthorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000001"))
	postAuthorID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000002"))
	strangerID := uuid.Must(uuid.FromString("10000000-0000-0000-0000-000000000003"))
	postID := uuid.Must(uuid.FromString("cccccccc-0000-0000-0000-000000000001"))
	commentID := uuid.Must(uuid.FromString("11111111-0000-0000-0000-000000000001"))

	t.Run("Delete by comment author succeeds", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &postAuthorID)
		comments := newFakeCommentRepository()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:      commentID,
					PostID:  postID,
					UserID:  &commentAuthorID,
					Content: "Comment content",
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		resp, err := service.DeleteComment(context.Background(), commentID.String(), commentAuthorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		deletedComment, ok := resp.(*dto.DeletedCommentResponse)
		if !ok {
			t.Fatalf("expected DeletedCommentResponse, got %T", resp)
		}
		if !deletedComment.Deleted {
			t.Fatal("expected Deleted = true")
		}
	})

	t.Run("Delete by post author succeeds", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &postAuthorID)
		comments := newFakeCommentRepository()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:      commentID,
					PostID:  postID,
					UserID:  &commentAuthorID,
					Content: "Comment content",
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		resp, err := service.DeleteComment(context.Background(), commentID.String(), postAuthorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		deletedComment, ok := resp.(*dto.DeletedCommentResponse)
		if !ok {
			t.Fatalf("expected DeletedCommentResponse, got %T", resp)
		}
		if !deletedComment.Deleted {
			t.Fatal("expected Deleted = true")
		}
	})

	t.Run("Delete by stranger fails", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &postAuthorID)
		comments := newFakeCommentRepository()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:      commentID,
					PostID:  postID,
					UserID:  &commentAuthorID,
					Content: "Comment content",
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		_, err := service.DeleteComment(context.Background(), commentID.String(), strangerID)
		if !errors.Is(err, ErrForbidden) {
			t.Fatalf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("Repeated delete is idempotent", func(t *testing.T) {
		posts := newFakePostRepository()
		posts.singleRow = makeSinglePostRow(t, postID, models.PostPrivacyPublic, &postAuthorID)
		comments := newFakeCommentRepository()
		deletedTime := time.Now()
		comments.commentsMap = map[uuid.UUID]*models.CommentWithAuthor{
			commentID: {
				Comment: models.Comment{
					ID:        commentID,
					PostID:    postID,
					UserID:    &commentAuthorID,
					Content:   "Comment content",
					DeletedAt: &deletedTime,
				},
			},
		}
		service := NewPostService(posts, newFakeUserRepository(), newFakeFollowersRepository(), newFakeGroupMembershipRepository(), comments)

		resp, err := service.DeleteComment(context.Background(), commentID.String(), commentAuthorID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		deletedComment, ok := resp.(*dto.DeletedCommentResponse)
		if !ok {
			t.Fatalf("expected DeletedCommentResponse, got %T", resp)
		}
		if !deletedComment.Deleted {
			t.Fatal("expected Deleted = true")
		}
	})
}
