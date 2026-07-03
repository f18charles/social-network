package repositories

import (
	"database/sql"
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type sqliteCommentRepository struct {
	db *sql.DB
}

// NewCommentRepository creates a SQLite-backed comment repository.
func NewCommentRepository(db *sql.DB) CommentRepository {
	return &sqliteCommentRepository{db: db}
}

func (r *sqliteCommentRepository) CreateComment(comment *models.Comment) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO comments (
			id, post_id, user_id, parent_comment_id, content, image_url,
			like_count, dislike_count, created_at, deleted_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(
		query,
		comment.ID.String(),
		comment.PostID.String(),
		nullableUUIDArg(comment.UserID),
		nullableUUIDArg(comment.ParentCommentID),
		comment.Content,
		nullableStringArg(comment.ImageURL),
		comment.LikeCount,
		comment.DislikeCount,
		comment.CreatedAt,
		nullableTimeArg(comment.DeletedAt),
		nullableTimeArg(comment.UpdatedAt),
	)
	if err != nil {
		return err
	}

	if comment.ParentCommentID == nil {
		_, err = tx.Exec(
			`UPDATE posts SET comment_count = comment_count + 1 WHERE id = ?`,
			comment.PostID.String(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqliteCommentRepository) GetCommentByID(id, viewerID uuid.UUID) (*models.CommentWithAuthor, error) {
	row := r.db.QueryRow(commentSelectSQL(`c.id = ?`, ""), viewerID.String(), id.String())
	comment, err := scanCommentWithAuthor(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("comment not found")
	}
	return comment, err
}

func (r *sqliteCommentRepository) ListCommentTreeByPost(postID, viewerID uuid.UUID, limit, offset int) ([]*models.CommentWithAuthor, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		WITH RECURSIVE roots AS (
			SELECT id
			FROM comments
			WHERE post_id = ? AND parent_comment_id IS NULL
			ORDER BY created_at ASC
			LIMIT ? OFFSET ?
		),
		tree(id) AS (
			SELECT id FROM roots
			UNION ALL
			SELECT child.id
			FROM comments child
			JOIN tree parent ON child.parent_comment_id = parent.id
		)
	` + commentSelectSQL(`c.id IN (SELECT id FROM tree)`, ` ORDER BY c.created_at ASC`)

	rows, err := r.db.Query(query, postID.String(), limit, offset, viewerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*models.CommentWithAuthor{}
	for rows.Next() {
		comment, err := scanCommentWithAuthor(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *sqliteCommentRepository) ListTopLevelCommentsByPost(postID, viewerID uuid.UUID, limit, offset int) ([]*models.CommentWithAuthor, error) {
	return r.listComments(
		commentSelectSQL(`c.post_id = ? AND c.parent_comment_id IS NULL`, ` ORDER BY c.created_at ASC, c.id ASC LIMIT ? OFFSET ?`),
		viewerID.String(),
		postID.String(),
		limit,
		maxInt(offset, 0),
	)
}

func (r *sqliteCommentRepository) ListRepliesByComment(parentCommentID, viewerID uuid.UUID, limit, offset int) ([]*models.CommentWithAuthor, error) {
	return r.listComments(
		commentSelectSQL(`c.parent_comment_id = ?`, ` ORDER BY c.created_at ASC, c.id ASC LIMIT ? OFFSET ?`),
		viewerID.String(),
		parentCommentID.String(),
		limit,
		maxInt(offset, 0),
	)
}

func (r *sqliteCommentRepository) ListCommentContext(commentID, viewerID uuid.UUID) ([]*models.CommentWithAuthor, error) {
	query := `
		WITH RECURSIVE ancestors(id, parent_comment_id, depth) AS (
			SELECT id, parent_comment_id, 0
			FROM comments
			WHERE id = ?
			UNION ALL
			SELECT parent.id, parent.parent_comment_id, ancestors.depth + 1
			FROM comments parent
			JOIN ancestors ON parent.id = ancestors.parent_comment_id
		)
	` + commentSelectSQL(`c.id IN (SELECT id FROM ancestors)`, ` ORDER BY (SELECT depth FROM ancestors WHERE ancestors.id = c.id) DESC`)

	return r.listComments(query, commentID.String(), viewerID.String())
}

func (r *sqliteCommentRepository) ListCommentsByAuthor(authorID, viewerID uuid.UUID, limit, offset int) ([]*models.CommentWithAuthor, error) {
	whereClause := `
		c.user_id = ?
		AND c.deleted_at IS NULL
		AND EXISTS (
			SELECT 1
			FROM posts p
			WHERE p.id = c.post_id
				AND p.deleted_at IS NULL
				AND (
					p.user_id = ?
					OR (
						p.group_id IS NULL
						AND (
							p.privacy = 'public'
							OR (
								p.privacy = 'almost_private'
								AND EXISTS (
									SELECT 1
									FROM followers f
									WHERE f.follower_id = ?
										AND f.followee_id = p.user_id
										AND f.status = 'accepted'
								)
							)
							OR (
								p.privacy = 'private'
								AND EXISTS (
									SELECT 1
									FROM post_audiences pa
									WHERE pa.post_id = p.id
										AND pa.user_id = ?
								)
							)
						)
					)
					OR (
						p.group_id IS NOT NULL
						AND EXISTS (
							SELECT 1
							FROM group_members gm
							WHERE gm.group_id = p.group_id
								AND gm.user_id = ?
								AND gm.status = 'accepted'
						)
					)
				)
		)`

	return r.listComments(
		commentSelectSQL(whereClause, ` ORDER BY c.created_at DESC, c.id DESC LIMIT ? OFFSET ?`),
		viewerID.String(),
		authorID.String(),
		viewerID.String(),
		viewerID.String(),
		viewerID.String(),
		viewerID.String(),
		limit,
		maxInt(offset, 0),
	)
}

func (r *sqliteCommentRepository) listComments(query string, args ...any) ([]*models.CommentWithAuthor, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []*models.CommentWithAuthor{}
	for rows.Next() {
		comment, err := scanCommentWithAuthor(rows)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *sqliteCommentRepository) UpdateComment(comment *models.Comment) error {
	query := `
		UPDATE comments
		SET content = ?, image_url = ?, updated_at = ?
		WHERE id = ?`
	_, err := r.db.Exec(
		query,
		comment.Content,
		nullableStringArg(comment.ImageURL),
		nullableTimeArg(comment.UpdatedAt),
		comment.ID.String(),
	)
	return err
}

func (r *sqliteCommentRepository) DeleteComment(id uuid.UUID, deletedAt time.Time) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var postID string
	var parentCommentID sql.NullString
	var existingDeletedAt sql.NullString
	if err := tx.QueryRow(`
		SELECT post_id, parent_comment_id, deleted_at
		FROM comments
		WHERE id = ?
	`, id.String()).Scan(&postID, &parentCommentID, &existingDeletedAt); err != nil {
		return err
	}

	query := `
		UPDATE comments
		SET deleted_at = ?, content = '', image_url = NULL
		WHERE id = ?`
	if _, err := tx.Exec(query, deletedAt.Format(time.RFC3339), id.String()); err != nil {
		return err
	}

	if !parentCommentID.Valid && !existingDeletedAt.Valid {
		if _, err := tx.Exec(`
			UPDATE posts
			SET comment_count = CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END
			WHERE id = ?
		`, postID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func commentSelectSQL(whereClause, suffix string) string {
	return `
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			c.parent_comment_id,
			c.content,
			c.image_url,
			c.like_count,
			c.dislike_count,
			c.created_at,
			c.deleted_at,
			c.updated_at,
			u.id,
			u.first_name,
			u.last_name,
			u.nickname,
			u.avatar,
			COALESCE(cv.vote, 'none') AS viewer_vote,
			(
				WITH RECURSIVE descendants(id) AS (
					SELECT child.id
					FROM comments child
					WHERE child.parent_comment_id = c.id
					UNION ALL
					SELECT nested.id
					FROM comments nested
					JOIN descendants d ON nested.parent_comment_id = d.id
				)
				SELECT COUNT(*)
				FROM comments counted
				JOIN descendants d ON counted.id = d.id
				WHERE counted.deleted_at IS NULL
			) AS replies_count
		FROM comments c
		LEFT JOIN users u ON u.id = c.user_id
		LEFT JOIN comment_votes cv ON cv.comment_id = c.id AND cv.user_id = ?
		WHERE ` + whereClause + suffix
}

func scanCommentWithAuthor(scanner rowScanner) (*models.CommentWithAuthor, error) {
	var (
		commentID       string
		postID          string
		userID          sql.NullString
		parentCommentID sql.NullString
		content         string
		imageURL        sql.NullString
		createdAt       string
		deletedAt       sql.NullString
		updatedAt       sql.NullString
		authorID        sql.NullString
		firstName       sql.NullString
		lastName        sql.NullString
		nickname        sql.NullString
		avatar          sql.NullString
		viewerVote      string
		repliesCount    int
		comment         models.Comment
	)

	err := scanner.Scan(
		&commentID,
		&postID,
		&userID,
		&parentCommentID,
		&content,
		&imageURL,
		&comment.LikeCount,
		&comment.DislikeCount,
		&createdAt,
		&deletedAt,
		&updatedAt,
		&authorID,
		&firstName,
		&lastName,
		&nickname,
		&avatar,
		&viewerVote,
		&repliesCount,
	)
	if err != nil {
		return nil, err
	}

	id, err := uuid.FromString(commentID)
	if err != nil {
		return nil, err
	}
	parsedPostID, err := uuid.FromString(postID)
	if err != nil {
		return nil, err
	}
	parsedUserID, err := nullableUUID(userID)
	if err != nil {
		return nil, err
	}
	parsedParentID, err := nullableUUID(parentCommentID)
	if err != nil {
		return nil, err
	}
	parsedCreatedAt, err := parseSQLiteTime(createdAt)
	if err != nil {
		return nil, err
	}
	parsedDeletedAt, err := nullableTime(deletedAt)
	if err != nil {
		return nil, err
	}
	parsedUpdatedAt, err := nullableTime(updatedAt)
	if err != nil {
		return nil, err
	}

	comment.ID = id
	comment.PostID = parsedPostID
	comment.UserID = parsedUserID
	comment.ParentCommentID = parsedParentID
	comment.Content = content
	comment.ImageURL = nullableString(imageURL)
	comment.CreatedAt = parsedCreatedAt
	comment.DeletedAt = parsedDeletedAt
	comment.UpdatedAt = parsedUpdatedAt

	return &models.CommentWithAuthor{
		Comment:      comment,
		Author:       scanPublicUser(authorID, firstName, lastName, nickname, avatar),
		ViewerVote:   models.ViewerVote(viewerVote),
		RepliesCount: repliesCount,
	}, nil
}

func nullableUUIDArg(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func nullableStringArg(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func nullableTimeArg(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}
