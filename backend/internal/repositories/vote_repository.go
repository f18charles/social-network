package repositories

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type sqlitePostVoteRepository struct {
	db *sql.DB
}

type sqliteCommentVoteRepository struct {
	db *sql.DB
}

// NewPostVoteRepository creates a SQLite-backed post vote repository.
func NewPostVoteRepository(db *sql.DB) PostVoteRepository {
	return &sqlitePostVoteRepository{db: db}
}

// NewCommentVoteRepository creates a SQLite-backed comment vote repository.
func NewCommentVoteRepository(db *sql.DB) CommentVoteRepository {
	return &sqliteCommentVoteRepository{db: db}
}

func (r *sqlitePostVoteRepository) SetPostVote(postID, userID uuid.UUID, vote models.VoteValue) (*models.VoteSummary, error) {
	if err := validateVoteValue(vote); err != nil {
		return nil, err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	_, err = tx.Exec(`
		INSERT INTO post_votes (post_id, user_id, vote, created_at, updated_at)
		VALUES (?, ?, ?, ?, NULL)
		ON CONFLICT(post_id, user_id) DO UPDATE SET
			vote = excluded.vote,
			updated_at = excluded.created_at
	`, postID.String(), userID.String(), string(vote), now)
	if err != nil {
		return nil, err
	}
	if err := refreshPostVoteCounts(tx, postID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetPostVoteSummary(postID, userID)
}

func (r *sqlitePostVoteRepository) DeletePostVote(postID, userID uuid.UUID) (*models.VoteSummary, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM post_votes WHERE post_id = ? AND user_id = ?`, postID.String(), userID.String()); err != nil {
		return nil, err
	}
	if err := refreshPostVoteCounts(tx, postID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetPostVoteSummary(postID, userID)
}

func (r *sqlitePostVoteRepository) GetPostVoteSummary(postID, viewerID uuid.UUID) (*models.VoteSummary, error) {
	row := r.db.QueryRow(`
		SELECT
			p.like_count,
			p.dislike_count,
			p.heart_count,
			COALESCE(pv.vote, 'none')
		FROM posts p
		LEFT JOIN post_votes pv ON pv.post_id = p.id AND pv.user_id = ?
		WHERE p.id = ?
	`, viewerID.String(), postID.String())

	summary, err := scanVoteSummary(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("post not found")
	}
	return summary, err
}

func (r *sqliteCommentVoteRepository) SetCommentVote(commentID, userID uuid.UUID, vote models.VoteValue) (*models.VoteSummary, error) {
	if err := validateVoteValue(vote); err != nil {
		return nil, err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	_, err = tx.Exec(`
		INSERT INTO comment_votes (comment_id, user_id, vote, created_at, updated_at)
		VALUES (?, ?, ?, ?, NULL)
		ON CONFLICT(comment_id, user_id) DO UPDATE SET
			vote = excluded.vote,
			updated_at = excluded.created_at
	`, commentID.String(), userID.String(), string(vote), now)
	if err != nil {
		return nil, err
	}
	if err := refreshCommentVoteCounts(tx, commentID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetCommentVoteSummary(commentID, userID)
}

func (r *sqliteCommentVoteRepository) DeleteCommentVote(commentID, userID uuid.UUID) (*models.VoteSummary, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM comment_votes WHERE comment_id = ? AND user_id = ?`, commentID.String(), userID.String()); err != nil {
		return nil, err
	}
	if err := refreshCommentVoteCounts(tx, commentID); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetCommentVoteSummary(commentID, userID)
}

func (r *sqliteCommentVoteRepository) GetCommentVoteSummary(commentID, viewerID uuid.UUID) (*models.VoteSummary, error) {
	row := r.db.QueryRow(`
		SELECT
			c.like_count,
			c.dislike_count,
			0,
			COALESCE(cv.vote, 'none')
		FROM comments c
		LEFT JOIN comment_votes cv ON cv.comment_id = c.id AND cv.user_id = ?
		WHERE c.id = ?
	`, viewerID.String(), commentID.String())

	summary, err := scanVoteSummary(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("comment not found")
	}
	return summary, err
}

func validateVoteValue(vote models.VoteValue) error {
	switch vote {
	case models.VoteValueLike, models.VoteValueDislike, models.VoteValueLove:
		return nil
	default:
		return errors.New("invalid vote value")
	}
}

func refreshPostVoteCounts(tx *sql.Tx, postID uuid.UUID) error {
	_, err := tx.Exec(`
		UPDATE posts
		SET
			like_count = (SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote = 'like'),
			dislike_count = (SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote = 'dislike'),
			heart_count = (SELECT COUNT(*) FROM post_votes WHERE post_id = ? AND vote = 'love')
		WHERE id = ?
	`, postID.String(), postID.String(), postID.String(), postID.String())
	return err
}

func refreshCommentVoteCounts(tx *sql.Tx, commentID uuid.UUID) error {
	_, err := tx.Exec(`
		UPDATE comments
		SET
			like_count = (SELECT COUNT(*) FROM comment_votes WHERE comment_id = ? AND vote = 'like'),
			dislike_count = (SELECT COUNT(*) FROM comment_votes WHERE comment_id = ? AND vote = 'dislike')
		WHERE id = ?
	`, commentID.String(), commentID.String(), commentID.String())
	return err
}

func scanVoteSummary(scanner rowScanner) (*models.VoteSummary, error) {
	var summary models.VoteSummary
	var viewerVote string
	if err := scanner.Scan(&summary.LikeCount, &summary.DislikeCount, &summary.HeartCount, &viewerVote); err != nil {
		return nil, err
	}
	summary.ViewerVote = models.ViewerVote(viewerVote)
	if summary.ViewerVote != models.ViewerVoteLike && summary.ViewerVote != models.ViewerVoteDislike && summary.ViewerVote != models.ViewerVoteLove {
		summary.ViewerVote = models.ViewerVoteNone
	}
	return &summary, nil
}
