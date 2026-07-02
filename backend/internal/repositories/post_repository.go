package repositories

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type sqlitePostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a SQLite-backed post repository.
func NewPostRepository(db *sql.DB) PostRepository {
	return &sqlitePostRepository{db: db}
}

// NewPostAudienceRepository creates a SQLite-backed selected post audience repository.
func NewPostAudienceRepository(db *sql.DB) PostAudienceRepository {
	return &sqlitePostRepository{db: db}
}

func (r *sqlitePostRepository) CreatePost(post *models.Post) error {
	query := `
		INSERT INTO posts (
			id, user_id, group_id, content, image_url, privacy,
			comment_count, like_count, dislike_count, created_at, updated_at, deleted_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(
		query,
		post.ID.String(),
		nullableUUIDArg(post.UserID),
		nullableUUIDArg(post.GroupID),
		post.Content,
		nullableStringArg(post.ImageURL),
		post.Privacy,
		post.CommentCount,
		post.LikeCount,
		post.DislikeCount,
		post.CreatedAt,
		nullableTimeArg(post.UpdatedAt),
		nullableTimeArg(post.DeletedAt),
	)
	return err
}

func (r *sqlitePostRepository) CreatePostWithAudience(post *models.Post, audience []uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO posts (
			id, user_id, group_id, content, image_url, privacy,
			comment_count, like_count, dislike_count, created_at, updated_at, deleted_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(
		query,
		post.ID.String(),
		nullableUUIDArg(post.UserID),
		nullableUUIDArg(post.GroupID),
		post.Content,
		nullableStringArg(post.ImageURL),
		post.Privacy,
		post.CommentCount,
		post.LikeCount,
		post.DislikeCount,
		post.CreatedAt,
		nullableTimeArg(post.UpdatedAt),
		nullableTimeArg(post.DeletedAt),
	)
	if err != nil {
		return err
	}

	for _, userID := range audience {
		_, err = tx.Exec(
			`INSERT INTO post_audiences (post_id, user_id) VALUES (?, ?)`,
			post.ID.String(),
			userID.String(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqlitePostRepository) GetPostByID(id, viewerID uuid.UUID) (*models.PostWithAuthor, error) {
	row := r.db.QueryRow(postSelectSQL(`p.id = ?`, ""), viewerID.String(), id.String())
	post, err := scanPostWithAuthor(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("post not found")
	}
	return post, err
}

func (r *sqlitePostRepository) ListPosts(query models.PostQuery, viewerID uuid.UUID) ([]*models.PostWithAuthor, error) {
	conditions := []string{"1 = 1"}
	args := []any{viewerID.String()}

	if query.AuthorID != nil {
		conditions = append(conditions, "p.user_id = ?")
		args = append(args, query.AuthorID.String())
	}
	if query.GroupID != nil {
		conditions = append(conditions, "p.group_id = ?")
		args = append(args, query.GroupID.String())
	}

	limitClause := ""
	if query.Limit > 0 {
		limitClause = " LIMIT ? OFFSET ?"
		args = append(args, query.Limit, maxInt(query.Offset, 0))
	}

	rows, err := r.db.Query(postSelectSQL(strings.Join(conditions, " AND "), " ORDER BY p.created_at DESC"+limitClause), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.PostWithAuthor{}
	for rows.Next() {
		post, err := scanPostWithAuthor(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *sqlitePostRepository) ListHomeFeed(viewerID uuid.UUID, limit, offset int) ([]*models.PostWithAuthor, error) {
	whereClause := `
		p.group_id IS NULL
		AND p.deleted_at IS NULL
		AND (
			p.user_id = ?
			OR p.privacy = 'public'
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
		)`
	args := []any{viewerID.String(), viewerID.String(), viewerID.String(), viewerID.String(), limit, maxInt(offset, 0)}
	return r.listFeed(whereClause, args)
}

func (r *sqlitePostRepository) ListProfilePosts(profileUserID, viewerID uuid.UUID, limit, offset int) ([]*models.PostWithAuthor, error) {
	whereClause := `
		p.group_id IS NULL
		AND p.deleted_at IS NULL
		AND p.user_id = ?
		AND (
			p.user_id = ?
			OR p.privacy = 'public'
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
		)`
	args := []any{viewerID.String(), profileUserID.String(), viewerID.String(), viewerID.String(), viewerID.String(), limit, maxInt(offset, 0)}
	return r.listFeed(whereClause, args)
}

func (r *sqlitePostRepository) ListGroupFeed(groupID, viewerID uuid.UUID, limit, offset int) ([]*models.PostWithAuthor, error) {
	whereClause := `p.group_id = ?`
	args := []any{viewerID.String(), groupID.String(), limit, maxInt(offset, 0)}
	return r.listFeed(whereClause, args)
}

func (r *sqlitePostRepository) listFeed(whereClause string, args []any) ([]*models.PostWithAuthor, error) {
	rows, err := r.db.Query(postSelectSQL(whereClause, " ORDER BY p.created_at DESC, p.id DESC LIMIT ? OFFSET ?"), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := []*models.PostWithAuthor{}
	for rows.Next() {
		post, err := scanPostWithAuthor(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *sqlitePostRepository) ReplacePostAudience(postID uuid.UUID, userIDs []uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM post_audiences WHERE post_id = ?`, postID.String()); err != nil {
		return err
	}
	for _, userID := range userIDs {
		if _, err := tx.Exec(`INSERT INTO post_audiences (post_id, user_id) VALUES (?, ?)`, postID.String(), userID.String()); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *sqlitePostRepository) ListPostAudience(postID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.db.Query(`SELECT user_id FROM post_audiences WHERE post_id = ? ORDER BY user_id`, postID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	userIDs := []uuid.UUID{}
	for rows.Next() {
		var rawID string
		if err := rows.Scan(&rawID); err != nil {
			return nil, err
		}
		userID, err := uuid.FromString(rawID)
		if err != nil {
			return nil, err
		}
		userIDs = append(userIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *sqlitePostRepository) IsPostAudienceMember(postID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM post_audiences WHERE post_id = ? AND user_id = ?)`,
		postID.String(),
		userID.String(),
	).Scan(&exists)
	return exists, err
}

func postSelectSQL(whereClause, suffix string) string {
	return `
		SELECT
			p.id,
			p.user_id,
			p.group_id,
			p.content,
			p.image_url,
			p.privacy,
			(
				SELECT COUNT(*)
				FROM comments c
				WHERE c.post_id = p.id
					AND c.parent_comment_id IS NULL
					AND c.deleted_at IS NULL
			) AS comment_count,
			p.like_count,
			p.dislike_count,
			p.created_at,
			p.updated_at,
			p.deleted_at,
			u.id,
			u.first_name,
			u.last_name,
			u.nickname,
			u.avatar,
			COALESCE(pv.vote, 'none') AS viewer_vote
		FROM posts p
		LEFT JOIN users u ON u.id = p.user_id
		LEFT JOIN post_votes pv ON pv.post_id = p.id AND pv.user_id = ?
		WHERE ` + whereClause + suffix
}

func scanPostWithAuthor(scanner rowScanner) (*models.PostWithAuthor, error) {
	var (
		postID      string
		userID      sql.NullString
		groupID     sql.NullString
		imageURL    sql.NullString
		privacy     string
		createdAt   string
		updatedAt   sql.NullString
		deletedAt   sql.NullString
		authorID    sql.NullString
		firstName   sql.NullString
		lastName    sql.NullString
		nickname    sql.NullString
		avatar      sql.NullString
		viewerVote  string
		postContent string
		post        models.Post
	)

	err := scanner.Scan(
		&postID,
		&userID,
		&groupID,
		&postContent,
		&imageURL,
		&privacy,
		&post.CommentCount,
		&post.LikeCount,
		&post.DislikeCount,
		&createdAt,
		&updatedAt,
		&deletedAt,
		&authorID,
		&firstName,
		&lastName,
		&nickname,
		&avatar,
		&viewerVote,
	)
	if err != nil {
		return nil, err
	}

	id, err := uuid.FromString(postID)
	if err != nil {
		return nil, err
	}
	parsedUserID, err := nullableUUID(userID)
	if err != nil {
		return nil, err
	}
	parsedGroupID, err := nullableUUID(groupID)
	if err != nil {
		return nil, err
	}
	parsedCreatedAt, err := parseSQLiteTime(createdAt)
	if err != nil {
		return nil, err
	}
	parsedUpdatedAt, err := nullableTime(updatedAt)
	if err != nil {
		return nil, err
	}
	parsedDeletedAt, err := nullableTime(deletedAt)
	if err != nil {
		return nil, err
	}

	post.ID = id
	post.UserID = parsedUserID
	post.GroupID = parsedGroupID
	post.Content = postContent
	post.ImageURL = nullableString(imageURL)
	post.Privacy = models.PostPrivacy(privacy)
	post.CreatedAt = parsedCreatedAt
	post.UpdatedAt = parsedUpdatedAt
	post.DeletedAt = parsedDeletedAt

	return &models.PostWithAuthor{
		Post:       post,
		Author:     scanPublicUser(authorID, firstName, lastName, nickname, avatar),
		ViewerVote: models.ViewerVote(viewerVote),
	}, nil
}

func scanPublicUser(id, firstName, lastName, nickname, avatar sql.NullString) *models.PublicUser {
	if !id.Valid || id.String == "" {
		return nil
	}
	userID, err := uuid.FromString(id.String)
	if err != nil {
		return nil
	}
	return &models.PublicUser{
		ID:        userID,
		FirstName: firstName.String,
		LastName:  lastName.String,
		Nickname:  nullableString(nickname),
		Avatar:    nullableString(avatar),
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (r *sqlitePostRepository) UpdatePostWithAudience(post *models.Post, audience []uuid.UUID) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		UPDATE posts
		SET content = ?, privacy = ?, image_url = ?, updated_at = ?
		WHERE id = ?`
	_, err = tx.Exec(
		query,
		post.Content,
		post.Privacy,
		nullableStringArg(post.ImageURL),
		nullableTimeArg(post.UpdatedAt),
		post.ID.String(),
	)
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM post_audiences WHERE post_id = ?`, post.ID.String()); err != nil {
		return err
	}

	for _, userID := range audience {
		_, err = tx.Exec(
			`INSERT INTO post_audiences (post_id, user_id) VALUES (?, ?)`,
			post.ID.String(),
			userID.String(),
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqlitePostRepository) DeletePost(id uuid.UUID) error {
	now := time.Now()
	_, err := r.db.Exec(
		`UPDATE posts SET deleted_at = ?, content = '', image_url = NULL WHERE id = ?`,
		now,
		id.String(),
	)
	return err
}
