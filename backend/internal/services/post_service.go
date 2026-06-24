package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/repositories"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/storage"
)

const (
	// DefaultFeedLimit is the limit used when a feed request omits limit.
	DefaultFeedLimit = 20
	// MaxFeedLimit is the largest accepted page size for post feeds.
	MaxFeedLimit = 100
)

var (
	// ErrForbidden means the current user is not allowed to access the feed.
	ErrForbidden = errors.New("forbidden")
	// ErrInvalidPagination means limit or offset is outside accepted bounds.
	ErrInvalidPagination = errors.New("invalid pagination")
	// ErrPostNotFound means the requested post does not exist.
	ErrPostNotFound = errors.New("post not found")
	// ErrPostForbidden means the current user is not allowed to access a post.
	ErrPostForbidden = errors.New("access to this post is forbidden")
	// ErrNotFollower means a selected private post audience is not an accepted follower.
	ErrNotFollower = errors.New("all audience members must be accepted followers")
	// ErrInvalidPrivacy means the privacy value is invalid.
	ErrInvalidPrivacy = errors.New("invalid privacy value")
	// ErrPostOrCommentDeleted means the post or parent comment is soft-deleted.
	ErrPostOrCommentDeleted = errors.New("post or selected parent comment is deleted")
	// ErrCrossPostParent means parent comment belongs to a different post.
	ErrCrossPostParent = errors.New("parent comment belongs to a different post")
	// ErrCommentNotFound means the requested comment does not exist.
	ErrCommentNotFound = errors.New("parent comment not found")
)

// PostService reads timeline, profile, group, and single-post views.
type PostService interface {
	CreatePost(ctx context.Context, req *models.CreatePostRequest, authorID uuid.UUID) (models.PostResponse, error)
	GetSinglePost(ctx context.Context, postID string, viewerID *string) (models.PostResponse, error)
	GetHomeFeed(viewerID uuid.UUID, limit, offset int) (*models.PostListResponse, error)
	GetProfilePosts(profileUserID, viewerID uuid.UUID, limit, offset int) (*models.PostListResponse, error)
	GetGroupFeed(groupID, viewerID uuid.UUID, limit, offset int) (*models.PostListResponse, error)
	GetCommentsByPost(ctx context.Context, postID string, viewerID uuid.UUID, limit, offset int) (*models.CommentListResponse, error)
	CreateComment(ctx context.Context, req *models.CreateCommentRequest, authorID uuid.UUID) (models.CommentResponse, error)
	UpdatePost(ctx context.Context, postID string, req *models.UpdatePostRequest, authorID uuid.UUID) (models.PostResponse, error)
	DeletePost(ctx context.Context, postID string, authorID uuid.UUID) (models.PostResponse, error)
	UpdateComment(ctx context.Context, commentID string, req *models.UpdateCommentRequest, authorID uuid.UUID) (models.CommentResponse, error)
	DeleteComment(ctx context.Context, commentID string, authorID uuid.UUID) (models.CommentResponse, error)
}

type postService struct {
	postRepo        repositories.PostRepository
	userRepo        repositories.UserRepository
	followerRepo    repositories.FollowersRepository
	groupMemberRepo repositories.GroupMembershipRepository
	commentRepo     repositories.CommentRepository
}

// NewPostService creates a service for authenticated post reads.
func NewPostService(
	postRepo repositories.PostRepository,
	userRepo repositories.UserRepository,
	followerRepo repositories.FollowersRepository,
	groupMemberRepo repositories.GroupMembershipRepository,
	commentRepo repositories.CommentRepository,
) PostService {
	return &postService{
		postRepo:        postRepo,
		userRepo:        userRepo,
		followerRepo:    followerRepo,
		groupMemberRepo: groupMemberRepo,
		commentRepo:     commentRepo,
	}
}

func (s *postService) GetSinglePost(ctx context.Context, postID string, viewerID *string) (models.PostResponse, error) {
	id, err := uuid.FromString(postID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	viewerUUID, ok := parseOptionalViewerID(viewerID)
	if !ok {
		return nil, ErrPostForbidden
	}

	row, err := s.postRepo.GetPostByID(id, viewerUUID)
	if err != nil {
		return nil, ErrPostNotFound
	}
	if err := s.canViewPost(row, viewerUUID); err != nil {
		return nil, err
	}

	post, err := models.MapPostResponse(row)
	if err != nil {
		return nil, fmt.Errorf("map single post: %w", err)
	}
	return post, nil
}

func (s *postService) GetHomeFeed(viewerID uuid.UUID, limit, offset int) (*models.PostListResponse, error) {
	page, err := normalizeFeedPagination(limit, offset)
	if err != nil {
		return nil, err
	}
	rows, err := s.postRepo.ListHomeFeed(viewerID, page.fetchLimit(), page.offset)
	if err != nil {
		return nil, err
	}
	return mapPostFeed("Posts returned.", rows, page)
}

func (s *postService) GetProfilePosts(profileUserID, viewerID uuid.UUID, limit, offset int) (*models.PostListResponse, error) {
	page, err := normalizeFeedPagination(limit, offset)
	if err != nil {
		return nil, err
	}
	if err := s.canViewProfile(profileUserID, viewerID); err != nil {
		return nil, err
	}
	rows, err := s.postRepo.ListProfilePosts(profileUserID, viewerID, page.fetchLimit(), page.offset)
	if err != nil {
		return nil, err
	}
	return mapPostFeed("Posts returned.", rows, page)
}

func (s *postService) GetGroupFeed(groupID, viewerID uuid.UUID, limit, offset int) (*models.PostListResponse, error) {
	page, err := normalizeFeedPagination(limit, offset)
	if err != nil {
		return nil, err
	}
	accepted, err := s.groupMemberRepo.IsAcceptedGroupMember(groupID, viewerID)
	if err != nil {
		return nil, err
	}
	if !accepted {
		return nil, ErrForbidden
	}
	rows, err := s.postRepo.ListGroupFeed(groupID, viewerID, page.fetchLimit(), page.offset)
	if err != nil {
		return nil, err
	}
	return mapPostFeed("Posts returned.", rows, page)
}

func (s *postService) canViewPost(row *models.PostWithAuthor, viewerID uuid.UUID) error {
	if row == nil {
		return ErrPostNotFound
	}
	if row.Post.GroupID != nil {
		accepted, err := s.groupMemberRepo.IsAcceptedGroupMember(*row.Post.GroupID, viewerID)
		if err != nil {
			return err
		}
		if !accepted {
			return ErrPostForbidden
		}
		return nil
	}
	if row.Post.UserID == nil {
		return ErrPostForbidden
	}
	if *row.Post.UserID == viewerID {
		return nil
	}
	switch row.Post.Privacy {
	case models.PostPrivacyPublic:
		return nil
	case models.PostPrivacyAlmostPrivate:
		status, err := s.followerRepo.GetStatus(viewerID, *row.Post.UserID)
		if err != nil {
			return err
		}
		if status == models.Accepted {
			return nil
		}
	case models.PostPrivacyPrivate:
		if audRepo, ok := s.postRepo.(repositories.PostAudienceRepository); ok {
			member, err := audRepo.IsPostAudienceMember(row.Post.ID, viewerID)
			if err != nil {
				return err
			}
			if member {
				return nil
			}
		}
		return ErrPostForbidden
	}
	return ErrPostForbidden
}

func (s *postService) canViewProfile(profileUserID, viewerID uuid.UUID) error {
	if profileUserID == viewerID {
		return nil
	}
	profileUser, err := s.userRepo.GetUserByID(profileUserID)
	if err != nil {
		return err
	}
	if profileUser.IsPublic {
		return nil
	}
	status, err := s.followerRepo.GetStatus(viewerID, profileUserID)
	if err != nil {
		return err
	}
	if status != models.Accepted {
		return ErrForbidden
	}
	return nil
}

type feedPage struct {
	limit  int
	offset int
}

func (p feedPage) fetchLimit() int {
	return p.limit + 1
}

func normalizeFeedPagination(limit, offset int) (feedPage, error) {
	if limit == 0 {
		limit = DefaultFeedLimit
	}
	if limit < 1 || limit > MaxFeedLimit || offset < 0 {
		return feedPage{}, ErrInvalidPagination
	}
	return feedPage{limit: limit, offset: offset}, nil
}

func mapPostFeed(message string, rows []*models.PostWithAuthor, page feedPage) (*models.PostListResponse, error) {
	hasMore := len(rows) > page.limit
	if hasMore {
		rows = rows[:page.limit]
	}

	posts := make([]models.PostResponse, 0, len(rows))
	for _, row := range rows {
		post, err := models.MapPostResponse(row)
		if err != nil {
			return nil, fmt.Errorf("map post feed: %w", err)
		}
		posts = append(posts, post)
	}

	return &models.PostListResponse{
		Status:  "success",
		Message: message,
		Data:    posts,
		Errors:  nil,
		Pagination: models.Pagination{
			Limit:   page.limit,
			Offset:  page.offset,
			HasMore: hasMore,
		},
	}, nil
}

func parseOptionalViewerID(viewerID *string) (uuid.UUID, bool) {
	if viewerID == nil || *viewerID == "" {
		return uuid.Nil, false
	}
	id, err := uuid.FromString(*viewerID)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func (s *postService) CreatePost(ctx context.Context, req *models.CreatePostRequest, authorID uuid.UUID) (models.PostResponse, error) {
	// 1. Enforce validation for group posts vs profile posts
	if req.GroupID != nil {
		// require accepted membership
		accepted, err := s.groupMemberRepo.IsAcceptedGroupMember(*req.GroupID, authorID)
		if err != nil {
			return nil, err
		}
		if !accepted {
			return nil, ErrForbidden
		}
		// ignore ordinary audience selection
		req.AudienceIDs = nil
		// enforce group-only visibility
		req.Privacy = models.PostPrivacyPublic
	} else {
		// validate privacy value
		if req.Privacy != models.PostPrivacyPublic &&
			req.Privacy != models.PostPrivacyAlmostPrivate &&
			req.Privacy != models.PostPrivacyPrivate {
			return nil, ErrInvalidPrivacy
		}

		// require each audience user to be an accepted follower for private posts
		if req.Privacy == models.PostPrivacyPrivate {
			for _, userID := range req.AudienceIDs {
				status, err := s.followerRepo.GetStatus(userID, authorID)
				if err != nil {
					return nil, err
				}
				if status != models.Accepted {
					return nil, ErrNotFollower
				}
			}
		} else {
			// ignore audience selection for non-private posts
			req.AudienceIDs = nil
		}
	}

	post := &models.Post{
		ID:           uuid.Must(uuid.NewV4()),
		UserID:       &authorID,
		GroupID:      req.GroupID,
		Content:      req.Content,
		ImageURL:     req.ImageURL,
		Privacy:      req.Privacy,
		CommentCount: 0,
		LikeCount:    0,
		DislikeCount: 0,
		CreatedAt:    time.Now().UTC(),
	}

	err := s.postRepo.CreatePostWithAudience(post, req.AudienceIDs)
	if err != nil {
		return nil, err
	}

	row, err := s.postRepo.GetPostByID(post.ID, authorID)
	if err != nil {
		return nil, err
	}

	return models.MapPostResponse(row)
}

func (s *postService) GetCommentsByPost(ctx context.Context, postID string, viewerID uuid.UUID, limit, offset int) (*models.CommentListResponse, error) {
	pID, err := uuid.FromString(postID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	// 1. Fetch the post row (even soft-deleted post rows are returned by GetPostByID)
	row, err := s.postRepo.GetPostByID(pID, viewerID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	// 2. Enforce post viewer permission checks
	if err := s.canViewPost(row, viewerID); err != nil {
		return nil, err
	}

	// Bounded top-level pagination
	if limit <= 0 {
		limit = DefaultFeedLimit
	}
	if limit > MaxFeedLimit {
		limit = MaxFeedLimit
	}
	if offset < 0 {
		offset = 0
	}

	// 3. Query the flat list of comments with 1 extra root to determine hasMore
	flatComments, err := s.commentRepo.ListCommentTreeByPost(pID, viewerID, limit+1, offset)
	if err != nil {
		return nil, err
	}

	// 4. Map flat collection to recursively nested tree structure in Go
	tree, err := models.MapCommentTree(flatComments)
	if err != nil {
		return nil, err
	}

	// Bounded pagination check
	hasMore := len(tree) > limit
	if hasMore {
		tree = tree[:limit]
	}

	return &models.CommentListResponse{
		Status:  "success",
		Message: "Comments returned.",
		Data:    tree,
		Errors:  nil,
		Pagination: models.Pagination{
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}, nil
}

func (s *postService) CreateComment(ctx context.Context, req *models.CreateCommentRequest, authorID uuid.UUID) (models.CommentResponse, error) {
	// 1. Fetch the post
	row, err := s.postRepo.GetPostByID(req.PostID, authorID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	// 2. Reject if the post is deleted
	if row.Post.DeletedAt != nil {
		return nil, ErrPostOrCommentDeleted
	}

	// 3. Enforce post permission checks
	if err := s.canViewPost(row, authorID); err != nil {
		return nil, err
	}

	// 4. Validate parent comment if provided
	if req.ParentCommentID != nil {
		parent, err := s.commentRepo.GetCommentByID(*req.ParentCommentID, authorID)
		if err != nil {
			return nil, ErrCommentNotFound
		}
		// Reject if parent comment is deleted (tombstone)
		if parent.Comment.DeletedAt != nil {
			return nil, ErrPostOrCommentDeleted
		}
		// Reject cross-post parent IDs
		if parent.Comment.PostID != req.PostID {
			return nil, ErrCrossPostParent
		}
	}

	// 5. Create the comment model
	commentID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	now := time.Now()

	comment := &models.Comment{
		ID:              commentID,
		PostID:          req.PostID,
		UserID:          &authorID,
		ParentCommentID: req.ParentCommentID,
		Content:         req.Content,
		ImageURL:        req.ImageURL,
		LikeCount:       0,
		DislikeCount:    0,
		CreatedAt:       now,
	}

	// 6. Save comment and update post count atomically
	err = s.commentRepo.CreateComment(comment)
	if err != nil {
		return nil, err
	}

	// 7. Map to response DTO
	author, err := s.userRepo.GetUserByID(authorID)
	if err != nil {
		return nil, err
	}
	var nicknamePtr *string
	if author.Nickname != "" {
		n := author.Nickname
		nicknamePtr = &n
	}
	var avatarPtr *string
	if author.Avatar != "" {
		a := author.Avatar
		avatarPtr = &a
	}

	publicAuthor := &models.PublicUser{
		ID:        author.ID,
		FirstName: author.FirstName,
		LastName:  author.LastName,
		Nickname:  nicknamePtr,
		Avatar:    avatarPtr,
	}

	commentWithAuthor := &models.CommentWithAuthor{
		Comment:    *comment,
		Author:     publicAuthor,
		ViewerVote: models.ViewerVoteNone,
	}

	return models.MapCommentResponse(commentWithAuthor, []models.CommentResponse{})
}

func (s *postService) UpdatePost(ctx context.Context, postID string, req *models.UpdatePostRequest, authorID uuid.UUID) (models.PostResponse, error) {
	pID, err := uuid.FromString(postID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	row, err := s.postRepo.GetPostByID(pID, authorID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	if row.Post.UserID == nil || *row.Post.UserID != authorID {
		return nil, ErrForbidden
	}

	if row.Post.DeletedAt != nil {
		return nil, ErrPostOrCommentDeleted
	}

	updatedPost := row.Post
	oldImageURL := row.Post.ImageURL

	// If content is being updated
	if req.Content != nil {
		content := *req.Content
		trimmedContent := strings.TrimSpace(content)

		hasImage := false
		if req.ImageURL != nil {
			hasImage = true
		} else if updatedPost.ImageURL != nil && !req.RemoveImage {
			hasImage = true
		}

		if trimmedContent == "" && !hasImage {
			return nil, errors.New("either content or image is required")
		}
		updatedPost.Content = trimmedContent
	}

	// Image update
	if req.RemoveImage {
		updatedPost.ImageURL = nil
	}
	if req.ImageURL != nil {
		updatedPost.ImageURL = req.ImageURL
	}

	// Privacy transition
	var audienceIDs []uuid.UUID = req.AudienceIDs
	if req.Privacy != nil {
		newPrivacy := *req.Privacy
		if row.Post.GroupID != nil {
			return nil, errors.New("group post privacy cannot be changed")
		}

		if newPrivacy == models.PostPrivacyPrivate {
			for _, followerID := range audienceIDs {
				status, err := s.followerRepo.GetStatus(followerID, authorID)
				if err != nil {
					return nil, err
				}
				if status != models.Accepted {
					return nil, ErrNotFollower
				}
			}
		} else {
			audienceIDs = nil
		}
		updatedPost.Privacy = newPrivacy
	} else {
		if updatedPost.Privacy == models.PostPrivacyPrivate {
			for _, followerID := range audienceIDs {
				status, err := s.followerRepo.GetStatus(followerID, authorID)
				if err != nil {
					return nil, err
				}
				if status != models.Accepted {
					return nil, ErrNotFollower
				}
			}
		} else {
			audienceIDs = nil
		}
	}

	now := time.Now()
	updatedPost.UpdatedAt = &now

	err = s.postRepo.UpdatePostWithAudience(&updatedPost, audienceIDs)
	if err != nil {
		return nil, err
	}

	if req.ImageURL != nil && oldImageURL != nil && *oldImageURL != *req.ImageURL {
		_ = storage.DeleteImage(*oldImageURL)
	}
	if req.RemoveImage && oldImageURL != nil {
		_ = storage.DeleteImage(*oldImageURL)
	}

	updatedRow, err := s.postRepo.GetPostByID(pID, authorID)
	if err != nil {
		return nil, err
	}

	return models.MapPostResponse(updatedRow)
}

func (s *postService) DeletePost(ctx context.Context, postID string, authorID uuid.UUID) (models.PostResponse, error) {
	pID, err := uuid.FromString(postID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	row, err := s.postRepo.GetPostByID(pID, authorID)
	if err != nil {
		return nil, ErrPostNotFound
	}

	if row.Post.UserID == nil || *row.Post.UserID != authorID {
		return nil, ErrForbidden
	}

	if row.Post.DeletedAt != nil {
		return &models.DeletedPostResponse{
			ID:      row.Post.ID,
			Deleted: true,
		}, nil
	}

	err = s.postRepo.DeletePost(pID)
	if err != nil {
		return nil, err
	}

	if row.Post.ImageURL != nil {
		_ = storage.DeleteImage(*row.Post.ImageURL)
	}

	return &models.DeletedPostResponse{
		ID:      row.Post.ID,
		Deleted: true,
	}, nil
}

func (s *postService) UpdateComment(ctx context.Context, commentID string, req *models.UpdateCommentRequest, authorID uuid.UUID) (models.CommentResponse, error) {
	cID, err := uuid.FromString(commentID)
	if err != nil {
		return nil, ErrCommentNotFound
	}

	row, err := s.commentRepo.GetCommentByID(cID, authorID)
	if err != nil {
		return nil, ErrCommentNotFound
	}

	// 1. Comment author check
	if row.Comment.UserID == nil || *row.Comment.UserID != authorID {
		return nil, ErrForbidden
	}

	// 2. Reject edits to deleted comments
	if row.Comment.DeletedAt != nil {
		return nil, ErrPostOrCommentDeleted
	}

	// 3. Reject edits if the comment belongs to a deleted post
	postRow, err := s.postRepo.GetPostByID(row.Comment.PostID, authorID)
	if err != nil {
		return nil, ErrPostNotFound
	}
	if postRow.Post.DeletedAt != nil {
		return nil, ErrPostOrCommentDeleted
	}

	updatedComment := row.Comment
	oldImageURL := row.Comment.ImageURL

	// 4. Update content
	if req.Content != nil {
		content := *req.Content
		trimmedContent := strings.TrimSpace(content)

		hasImage := false
		if req.ImageURL != nil {
			hasImage = true
		} else if updatedComment.ImageURL != nil && !req.RemoveImage {
			hasImage = true
		}

		if trimmedContent == "" && !hasImage {
			return nil, errors.New("either content or image is required")
		}
		updatedComment.Content = trimmedContent
	}

	// 5. Update image
	if req.RemoveImage {
		updatedComment.ImageURL = nil
	}
	if req.ImageURL != nil {
		updatedComment.ImageURL = req.ImageURL
	}

	now := time.Now()
	updatedComment.UpdatedAt = &now

	err = s.commentRepo.UpdateComment(&updatedComment)
	if err != nil {
		return nil, err
	}

	if req.ImageURL != nil && oldImageURL != nil && *oldImageURL != *req.ImageURL {
		_ = storage.DeleteImage(*oldImageURL)
	}
	if req.RemoveImage && oldImageURL != nil {
		_ = storage.DeleteImage(*oldImageURL)
	}

	// Fetch updated comment to return
	updatedRow, err := s.commentRepo.GetCommentByID(cID, authorID)
	if err != nil {
		return nil, err
	}

	return models.MapCommentResponse(updatedRow, []models.CommentResponse{})
}

func (s *postService) DeleteComment(ctx context.Context, commentID string, authorID uuid.UUID) (models.CommentResponse, error) {
	cID, err := uuid.FromString(commentID)
	if err != nil {
		return nil, ErrCommentNotFound
	}

	row, err := s.commentRepo.GetCommentByID(cID, authorID)
	if err != nil {
		return nil, ErrCommentNotFound
	}

	// Idempotency: if already soft-deleted, return the minimal tombstone.
	if row.Comment.DeletedAt != nil {
		return &models.DeletedCommentResponse{
			ID:      row.Comment.ID,
			Deleted: true,
			Replies: []models.CommentResponse{},
		}, nil
	}

	// 1. Authorization: Allowed for the comment author OR parent post author.
	isCommentAuthor := row.Comment.UserID != nil && *row.Comment.UserID == authorID
	isPostAuthor := false

	postRow, err := s.postRepo.GetPostByID(row.Comment.PostID, authorID)
	if err == nil && postRow.Post.UserID != nil && *postRow.Post.UserID == authorID {
		isPostAuthor = true
	}

	if !isCommentAuthor && !isPostAuthor {
		return nil, ErrForbidden
	}

	// 2. Perform soft-delete
	now := time.Now()
	err = s.commentRepo.DeleteComment(cID, now)
	if err != nil {
		return nil, err
	}

	// 3. Remove media file if existed
	if row.Comment.ImageURL != nil {
		_ = storage.DeleteImage(*row.Comment.ImageURL)
	}

	return &models.DeletedCommentResponse{
		ID:      row.Comment.ID,
		Deleted: true,
		Replies: []models.CommentResponse{},
	}, nil
}
