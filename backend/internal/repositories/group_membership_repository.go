package repositories

import (
	"database/sql"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type sqliteGroupMembershipRepository struct {
	db *sql.DB
}

// NewGroupMembershipRepository creates a SQLite-backed group membership reader and writer.
func NewGroupMembershipRepository(db *sql.DB) GroupMembershipRepository {
	return &sqliteGroupMembershipRepository{db: db}
}

func (r *sqliteGroupMembershipRepository) IsAcceptedGroupMember(groupID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM group_members
			WHERE group_id = ?
				AND user_id = ?
				AND status = 'accepted'
		)
	`, groupID.String(), userID.String()).Scan(&exists)
	return exists, err
}

func (r *sqliteGroupMembershipRepository) GetMembership(groupID, userID uuid.UUID) (string, error) {
	var status string
	err := r.db.QueryRow(`
		SELECT status
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID.String(), userID.String()).Scan(&status)
	if err == sql.ErrNoRows {
		return "none", nil
	}
	return status, err
}

func (r *sqliteGroupMembershipRepository) AddMembership(groupID, userID uuid.UUID, status string) error {
	query := `INSERT INTO group_members (group_id, user_id, status) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, groupID.String(), userID.String(), status)
	return err
}

func (r *sqliteGroupMembershipRepository) UpdateMembershipStatus(groupID, userID uuid.UUID, status string) error {
	query := `UPDATE group_members SET status = ? WHERE group_id = ? AND user_id = ?`
	_, err := r.db.Exec(query, status, groupID.String(), userID.String())
	return err
}

func (r *sqliteGroupMembershipRepository) RemoveMembership(groupID, userID uuid.UUID) error {
	query := `DELETE FROM group_members WHERE group_id = ? AND user_id = ?`
	_, err := r.db.Exec(query, groupID.String(), userID.String())
	return err
}

func (r *sqliteGroupMembershipRepository) ListGroupMembers(groupID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.first_name, u.last_name, u.dob, u.avatar, u.nickname, u.about_me, u.is_public, u.follower_count, u.following_count, u.created_at
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'accepted'
		ORDER BY CASE gm.role WHEN 'admin' THEN 0 ELSE 1 END, lower(u.first_name), lower(u.last_name), lower(u.email)
	`
	return r.scanUsers(query, groupID)
}

func (r *sqliteGroupMembershipRepository) ListPendingRequests(groupID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.first_name, u.last_name, u.dob, u.avatar, u.nickname, u.about_me, u.is_public, u.follower_count, u.following_count, u.created_at
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'pending_request'
	`
	return r.scanUsers(query, groupID)
}

func (r *sqliteGroupMembershipRepository) scanUsers(query string, groupID uuid.UUID) ([]*models.User, error) {
	rows, err := r.db.Query(query, groupID.String())
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
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *sqliteGroupMembershipRepository) IsGroupAdmin(groupID, userID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1
			FROM group_members
			WHERE group_id = ?
				AND user_id = ?
				AND status = 'accepted'
				AND role = 'admin'
		)
	`, groupID.String(), userID.String()).Scan(&exists)
	return exists, err
}

func (r *sqliteGroupMembershipRepository) CountGroupAdmins(groupID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ? AND status = 'accepted' AND role = 'admin'`, groupID.String()).Scan(&count)
	return count, err
}

func (r *sqliteGroupMembershipRepository) GetMembershipRole(groupID, userID uuid.UUID) (string, error) {
	var role string
	err := r.db.QueryRow(`SELECT role FROM group_members WHERE group_id = ? AND user_id = ?`, groupID.String(), userID.String()).Scan(&role)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return role, err
}

func (r *sqliteGroupMembershipRepository) UpdateMembershipRole(groupID, userID uuid.UUID, role string) error {
	_, err := r.db.Exec(`UPDATE group_members SET role = ? WHERE group_id = ? AND user_id = ? AND status = 'accepted'`, role, groupID.String(), userID.String())
	return err
}

func (r *sqliteGroupMembershipRepository) ListGroupMembersWithRoles(groupID uuid.UUID) ([]*models.GroupMemberUser, error) {
	rows, err := r.db.Query(`
		SELECT u.id, u.email, u.password_hash, u.first_name, u.last_name, u.dob, u.avatar, u.nickname, u.about_me, u.is_public, u.follower_count, u.following_count, u.created_at, gm.status, gm.role
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'accepted'
		ORDER BY CASE gm.role WHEN 'admin' THEN 0 ELSE 1 END, lower(u.first_name), lower(u.last_name), lower(u.email)
	`, groupID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := []*models.GroupMemberUser{}
	for rows.Next() {
		member, err := scanGroupMemberUser(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

func (r *sqliteGroupMembershipRepository) ListPendingInvitations(groupID uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT u.id, u.email, u.password_hash, u.first_name, u.last_name, u.dob, u.avatar, u.nickname, u.about_me, u.is_public, u.follower_count, u.following_count, u.created_at
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = ? AND gm.status = 'pending_invite'
		ORDER BY lower(u.first_name), lower(u.last_name), lower(u.email)
	`
	return r.scanUsers(query, groupID)
}

func scanGroupMemberUser(rows *sql.Rows) (*models.GroupMemberUser, error) {
	u := &models.User{}
	var avatar, nickname, aboutMe sql.NullString
	var dob, createdAt string
	member := &models.GroupMemberUser{User: *u}
	if err := rows.Scan(&u.ID, &u.Email, &u.PassHash, &u.FirstName, &u.LastName, &dob, &avatar, &nickname, &aboutMe, &u.IsPublic, &u.FollowerCount, &u.FollowingCount, &createdAt, &member.Status, &member.Role); err != nil {
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
	member.User = *u
	return member, nil
}
