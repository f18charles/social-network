package repositories

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (ur *userRepository) CreateUser(u *models.User) error {
	query := `INSERT INTO users (id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public, follower_count, following_count, created_at) 
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := ur.db.Exec(query, u.ID, u.Email, u.PassHash, u.FirstName, u.LastName, u.DOB, u.Avatar, u.Nickname, u.AboutMe, u.IsPublic, u.FollowerCount, u.FollowingCount, u.CreatedAt)
	return err
}

func (ur *userRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public, follower_count, following_count, created_at FROM users WHERE id = ?`
	u := &models.User{}
	var avatar, nickname, aboutMe sql.NullString
	var dob, createdAt string
	err := ur.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.PassHash, &u.FirstName, &u.LastName, &dob, &avatar, &nickname, &aboutMe, &u.IsPublic, &u.FollowerCount, &u.FollowingCount, &createdAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	parsedDOB, err := parseSQLiteTime(dob)
	if err != nil {
		return nil, err
	}
	parsedCreatedAt, err := parseSQLiteTime(createdAt)
	if err != nil {
		return nil, err
	}
	u.DOB = parsedDOB
	u.CreatedAt = parsedCreatedAt
	u.Avatar = avatar.String
	u.Nickname = nickname.String
	u.AboutMe = aboutMe.String
	return u, nil
}

func (ur *userRepository) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public, follower_count, following_count, created_at FROM users WHERE email = ?`
	u := &models.User{}
	var avatar, nickname, aboutMe sql.NullString
	var dob, createdAt string
	err := ur.db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PassHash, &u.FirstName, &u.LastName, &dob, &avatar, &nickname, &aboutMe, &u.IsPublic, &u.FollowerCount, &u.FollowingCount, &createdAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	if err != nil {
		return nil, err
	}
	parsedDOB, err := parseSQLiteTime(dob)
	if err != nil {
		return nil, err
	}
	parsedCreatedAt, err := parseSQLiteTime(createdAt)
	if err != nil {
		return nil, err
	}
	u.DOB = parsedDOB
	u.CreatedAt = parsedCreatedAt
	u.Avatar = avatar.String
	u.Nickname = nickname.String
	u.AboutMe = aboutMe.String
	return u, nil
}

// UpdateUserProfile updates an existing user's profile information in the database.
func (ur *userRepository) UpdateUserProfile(u *models.User) error {
	query := `
		UPDATE users 
		SET email = ?, password_hash = ?, first_name = ?, last_name = ?, dob = ?, avatar = ?, nickname = ?, about_me = ?, is_public = ?
		WHERE id = ?
	`
	_, err := ur.db.Exec(query, u.Email, u.PassHash, u.FirstName, u.LastName, u.DOB, u.Avatar, u.Nickname, u.AboutMe, u.IsPublic, u.ID.String())
	return err
}

func (ur *userRepository) ListUsers(queryText string, excludeID uuid.UUID) ([]*models.User, error) {
	var rows *sql.Rows
	var err error
	if queryText == "" {
		rows, err = ur.db.Query(
			`SELECT id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public, follower_count, following_count, created_at
			 FROM users
			 WHERE is_public = 1 OR is_public = 0 AND id != ?
			 ORDER BY created_at DESC
			 LIMIT 50`, excludeID)
	} else {
		search := "%" + strings.ToLower(strings.TrimSpace(queryText)) + "%"
		rows, err = ur.db.Query(
			`SELECT id, email, password_hash, first_name, last_name, dob, avatar, nickname, about_me, is_public, follower_count, following_count, created_at
			 FROM users
			 WHERE is_public = 1 OR is_public = 0 AND id != ? AND (
				LOWER(nickname) LIKE ? OR LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ?
			 )
			 ORDER BY created_at DESC
			 LIMIT 50`, excludeID, search, search, search)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		u := &models.User{}
		var avatar, nickname, aboutMe sql.NullString
		var dob, createdAt string
		err := rows.Scan(&u.ID, &u.Email, &u.PassHash, &u.FirstName, &u.LastName, &dob, &avatar, &nickname, &aboutMe, &u.IsPublic, &u.FollowerCount, &u.FollowingCount, &createdAt)
		if err != nil {
			return nil, err
		}
		parsedDOB, err := parseSQLiteTime(dob)
		if err != nil {
			return nil, err
		}
		parsedCreatedAt, err := parseSQLiteTime(createdAt)
		if err != nil {
			return nil, err
		}
		u.DOB = parsedDOB
		u.CreatedAt = parsedCreatedAt
		u.Avatar = avatar.String
		u.Nickname = nickname.String
		u.AboutMe = aboutMe.String
		users = append(users, u)
	}
	return users, nil
}

// DeleteUser removes an account after turning authored posts and comments into tombstones.
func (ur *userRepository) DeleteUser(id uuid.UUID) error {
	tx, err := ur.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := tx.Exec(`CREATE TEMP TABLE IF NOT EXISTS deleted_user_affected_posts (id TEXT PRIMARY KEY)`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM deleted_user_affected_posts`); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		INSERT OR IGNORE INTO deleted_user_affected_posts (id)
		SELECT id FROM posts WHERE user_id = ?
		UNION
		SELECT post_id FROM comments WHERE user_id = ?`, id.String(), id.String()); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE posts
		SET deleted_at = COALESCE(deleted_at, ?), content = '', image_url = NULL
		WHERE user_id = ?`, now, id.String()); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE comments
		SET deleted_at = COALESCE(deleted_at, ?), content = '', image_url = NULL
		WHERE user_id = ?`, now, id.String()); err != nil {
		return err
	}
	if _, err := tx.Exec(`
		UPDATE posts
		SET comment_count = (
			SELECT COUNT(*)
			FROM comments c
			WHERE c.post_id = posts.id
				AND c.parent_comment_id IS NULL
				AND c.deleted_at IS NULL
		)
		WHERE id IN (SELECT id FROM deleted_user_affected_posts)`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM sessions WHERE user_id = ?`, id.String()); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM followers WHERE follower_id = ? OR followee_id = ?`, id.String(), id.String()); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM users WHERE id = ?`, id.String()); err != nil {
		return err
	}
	return tx.Commit()
}
