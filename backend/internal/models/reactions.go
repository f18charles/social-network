package models

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type HeartReaction struct {
	PostID    uuid.UUID  `db:"post_id"`
	UserID    uuid.UUID  `db:"user_id"`
	CreatedAt time.Time  `db:"created_at"`
}

type ReactionSummary struct {
	LikeCount    int        `json:"like_count"`
	DislikeCount int        `json:"dislike_count"`
	HeartCount   int        `json:"heart_count"`
	ViewerVote   ViewerVote `json:"viewer_vote"`
}
