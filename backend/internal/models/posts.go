package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

// ViewerVote is the current viewer's vote state for a post or comment.
type ViewerVote string

const (
	// ViewerVoteLike means the current viewer liked the resource.
	ViewerVoteLike ViewerVote = "like"
	// ViewerVoteDislike means the current viewer disliked the resource.
	ViewerVoteDislike ViewerVote = "dislike"
	// ViewerVoteNone means the current viewer has not voted on the resource.
	ViewerVoteNone ViewerVote = "none"
)

// VoteValue is a persisted vote value. It excludes the derived "none" state.
type VoteValue string

const (
	// VoteValueLike stores a like vote.
	VoteValueLike VoteValue = "like"
	// VoteValueDislike stores a dislike vote.
	VoteValueDislike VoteValue = "dislike"
)

// PostPrivacy is the visibility mode for non-group profile posts.
type PostPrivacy string

const (
	// PostPrivacyPublic makes a post visible to all authenticated users.
	PostPrivacyPublic PostPrivacy = "public"
	// PostPrivacyAlmostPrivate makes a post visible to accepted followers.
	PostPrivacyAlmostPrivate PostPrivacy = "almost_private"
	// PostPrivacyPrivate makes a post visible only to selected accepted followers.
	PostPrivacyPrivate PostPrivacy = "private"
)

// PublicUser contains only identity fields safe to embed in another response.
type PublicUser struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Nickname  *string
	Avatar    *string
}

// Post is the database representation of a social post.
type Post struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	GroupID      *uuid.UUID
	Content      string
	ImageURL     *string
	Privacy      PostPrivacy
	CommentCount int
	LikeCount    int
	DislikeCount int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time
}

// PostAudience records one selected follower allowed to view a private post.
type PostAudience struct {
	PostID uuid.UUID
	UserID uuid.UUID
}

// PostVote records one user's current vote on a post.
type PostVote struct {
	PostID    uuid.UUID
	UserID    uuid.UUID
	Vote      VoteValue
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// VoteSummary contains current vote counts plus the current viewer's vote.
type VoteSummary struct {
	LikeCount    int
	DislikeCount int
	ViewerVote   ViewerVote
}

// PostQuery filters reusable post-list repository reads.
type PostQuery struct {
	AuthorID *uuid.UUID
	GroupID  *uuid.UUID
	Limit    int
	Offset   int
}

// PostWithAuthor is a repository read model with hydrated author and viewer state.
type PostWithAuthor struct {
	Post       Post
	Author     *PublicUser
	ViewerVote ViewerVote
}

// Comment is the database representation of a post comment or nested reply.
type Comment struct {
	ID              uuid.UUID
	PostID          uuid.UUID
	UserID          *uuid.UUID
	ParentCommentID *uuid.UUID
	Content         string
	ImageURL        *string
	LikeCount       int
	DislikeCount    int
	CreatedAt       time.Time
	DeletedAt       *time.Time
	UpdatedAt       *time.Time
}

// CommentVote records one user's current vote on a comment.
type CommentVote struct {
	CommentID uuid.UUID
	UserID    uuid.UUID
	Vote      VoteValue
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// CommentWithAuthor is a repository read model with hydrated author and viewer state.
type CommentWithAuthor struct {
	Comment      Comment
	Author       *PublicUser
	ViewerVote   ViewerVote
	RepliesCount int
}
