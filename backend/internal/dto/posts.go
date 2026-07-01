package dto

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

// CreatePostRequest is the normalized service input for creating a post.
type CreatePostRequest struct {
	Content     string
	Privacy     models.PostPrivacy
	GroupID     *uuid.UUID
	AudienceIDs []uuid.UUID
	ImageURL    *string
}

// UpdatePostRequest contains normalized fields to edit an existing post.
type UpdatePostRequest struct {
	Content     *string
	Privacy     *models.PostPrivacy
	AudienceIDs []uuid.UUID
	ImageURL    *string
	RemoveImage bool
}

// CreateCommentRequest contains normalized fields to write a comment or reply.
type CreateCommentRequest struct {
	PostID          uuid.UUID
	ParentCommentID *uuid.UUID
	Content         string
	ImageURL        *string
}

// UpdateCommentRequest contains normalized fields to edit an existing comment.
type UpdateCommentRequest struct {
	Content     *string
	ImageURL    *string
	RemoveImage bool
}

// Pagination describes offset pagination state for list responses.
type Pagination struct {
	Limit   int  `json:"limit"`
	Offset  int  `json:"offset"`
	HasMore bool `json:"has_more"`
}

// PostListResponse is the API response envelope for paginated post feeds.
type PostListResponse struct {
	Status     string            `json:"status"`
	Message    string            `json:"message"`
	Data       []PostResponse    `json:"data"`
	Errors     map[string]string `json:"errors"`
	Pagination Pagination        `json:"pagination"`
}

// CommentListResponse is the API response envelope for nested comment trees.
type CommentListResponse struct {
	Status     string            `json:"status"`
	Message    string            `json:"message"`
	Data       []CommentResponse `json:"data"`
	Errors     map[string]string `json:"errors"`
	Pagination Pagination        `json:"pagination"`
}

// PostResponse is implemented by active and deleted post response DTOs.
type PostResponse interface {
	isPostResponse()
}

// ActivePostResponse is the full API DTO for a visible, non-deleted post.
type ActivePostResponse struct {
	ID           uuid.UUID          `json:"id"`
	Deleted      bool               `json:"deleted"`
	Author       PublicUserResponse `json:"author"`
	GroupID      *uuid.UUID         `json:"group_id"`
	Content      string             `json:"content"`
	ImageURL     *string            `json:"image_url"`
	Privacy      models.PostPrivacy `json:"privacy"`
	CommentCount int                `json:"comment_count"`
	LikeCount    int                `json:"like_count"`
	DislikeCount int                `json:"dislike_count"`
	ViewerVote   models.ViewerVote  `json:"viewer_vote"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    *time.Time         `json:"updated_at"`
}

// DeletedPostResponse is the minimal API tombstone for a soft-deleted post.
type DeletedPostResponse struct {
	ID      uuid.UUID `json:"id"`
	Deleted bool      `json:"deleted"`
}

func (*ActivePostResponse) isPostResponse()  {}
func (*DeletedPostResponse) isPostResponse() {}

// MapPostResponse maps a repository post row to a safe active or tombstone DTO.
func MapPostResponse(row *models.PostWithAuthor) (PostResponse, error) {
	if row == nil {
		return nil, errors.New("post row is required")
	}
	if row.Post.DeletedAt != nil {
		return &DeletedPostResponse{ID: row.Post.ID, Deleted: true}, nil
	}

	author, err := MapPublicUserResponse(row.Author)
	if err != nil {
		return nil, fmt.Errorf("map active post author: %w", err)
	}

	return &ActivePostResponse{
		ID:           row.Post.ID,
		Deleted:      false,
		Author:       author,
		GroupID:      row.Post.GroupID,
		Content:      row.Post.Content,
		ImageURL:     row.Post.ImageURL,
		Privacy:      row.Post.Privacy,
		CommentCount: row.Post.CommentCount,
		LikeCount:    row.Post.LikeCount,
		DislikeCount: row.Post.DislikeCount,
		ViewerVote:   normalizeViewerVote(row.ViewerVote),
		CreatedAt:    row.Post.CreatedAt,
		UpdatedAt:    row.Post.UpdatedAt,
	}, nil
}

// CommentResponse is implemented by active and deleted comment response DTOs.
type CommentResponse interface {
	isCommentResponse()
}

// ActiveCommentResponse is the full API DTO for a visible, non-deleted comment.
type ActiveCommentResponse struct {
	ID              uuid.UUID          `json:"id"`
	Deleted         bool               `json:"deleted"`
	PostID          uuid.UUID          `json:"post_id"`
	ParentCommentID *uuid.UUID         `json:"parent_comment_id"`
	Author          PublicUserResponse `json:"author"`
	Content         string             `json:"content"`
	ImageURL        *string            `json:"image_url"`
	LikeCount       int                `json:"like_count"`
	DislikeCount    int                `json:"dislike_count"`
	ViewerVote      models.ViewerVote  `json:"viewer_vote"`
	CreatedAt       time.Time          `json:"created_at"`
	UpdatedAt       *time.Time         `json:"updated_at"`
	Replies         []CommentResponse  `json:"replies"`
}

// DeletedCommentResponse is the minimal API tombstone for a soft-deleted comment.
type DeletedCommentResponse struct {
	ID      uuid.UUID         `json:"id"`
	Deleted bool              `json:"deleted"`
	Replies []CommentResponse `json:"replies"`
}

func (*ActiveCommentResponse) isCommentResponse()  {}
func (*DeletedCommentResponse) isCommentResponse() {}

// MapCommentResponse maps one comment row and already-mapped replies to a safe DTO.
func MapCommentResponse(row *models.CommentWithAuthor, replies []CommentResponse) (CommentResponse, error) {
	if row == nil {
		return nil, errors.New("comment row is required")
	}
	if replies == nil {
		replies = []CommentResponse{}
	}
	if row.Comment.DeletedAt != nil {
		return &DeletedCommentResponse{ID: row.Comment.ID, Deleted: true, Replies: replies}, nil
	}

	author, err := MapPublicUserResponse(row.Author)
	if err != nil {
		return nil, fmt.Errorf("map active comment author: %w", err)
	}

	return &ActiveCommentResponse{
		ID:              row.Comment.ID,
		Deleted:         false,
		PostID:          row.Comment.PostID,
		ParentCommentID: row.Comment.ParentCommentID,
		Author:          author,
		Content:         row.Comment.Content,
		ImageURL:        row.Comment.ImageURL,
		LikeCount:       row.Comment.LikeCount,
		DislikeCount:    row.Comment.DislikeCount,
		ViewerVote:      normalizeViewerVote(row.ViewerVote),
		CreatedAt:       row.Comment.CreatedAt,
		UpdatedAt:       row.Comment.UpdatedAt,
		Replies:         replies,
	}, nil
}

// MapCommentTree maps a flat repository result into recursively nested safe DTOs.
func MapCommentTree(rows []*models.CommentWithAuthor) ([]CommentResponse, error) {
	nodes := make(map[uuid.UUID]*commentNode, len(rows))
	order := make([]uuid.UUID, 0, len(rows))

	for _, row := range rows {
		if row == nil {
			return nil, errors.New("comment tree contains nil row")
		}
		if _, exists := nodes[row.Comment.ID]; exists {
			return nil, fmt.Errorf("duplicate comment row %s", row.Comment.ID)
		}
		nodes[row.Comment.ID] = &commentNode{row: row, replies: []CommentResponse{}}
		order = append(order, row.Comment.ID)
	}

	sort.SliceStable(order, func(i, j int) bool {
		return nodes[order[i]].row.Comment.CreatedAt.Before(nodes[order[j]].row.Comment.CreatedAt)
	})

	roots := make([]CommentResponse, 0)
	for i := len(order) - 1; i >= 0; i-- {
		id := order[i]
		node := nodes[id]
		response, err := MapCommentResponse(node.row, node.replies)
		if err != nil {
			return nil, err
		}
		if node.row.Comment.ParentCommentID == nil {
			roots = append([]CommentResponse{response}, roots...)
			continue
		}
		parent, ok := nodes[*node.row.Comment.ParentCommentID]
		if !ok {
			return nil, fmt.Errorf("comment %s references missing parent %s", id, *node.row.Comment.ParentCommentID)
		}
		parent.replies = append([]CommentResponse{response}, parent.replies...)
	}

	if roots == nil {
		return []CommentResponse{}, nil
	}
	return roots, nil
}

type commentNode struct {
	row     *models.CommentWithAuthor
	replies []CommentResponse
}

func normalizeViewerVote(vote models.ViewerVote) models.ViewerVote {
	switch vote {
	case models.ViewerVoteLike, models.ViewerVoteDislike:
		return vote
	default:
		return models.ViewerVoteNone
	}
}
