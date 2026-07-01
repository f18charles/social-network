package repositories

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type sqliteNotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) NotificationRepository {
	return &sqliteNotificationRepository{db: db}
}

func (r *sqliteNotificationRepository) CreateNotification(n *models.Notification) error {
	query := `INSERT INTO notifications (id, user_id, type, source_id, group_id, is_read, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	isReadInt := 0
	if n.IsRead {
		isReadInt = 1
	}

	var groupID sql.NullString
	if n.GroupID != nil {
		groupID = sql.NullString{String: n.GroupID.String(), Valid: true}
	}

	_, err := r.db.Exec(query, n.ID.String(), n.UserID.String(), n.Type, n.SourceID.String(), groupID, isReadInt, n.CreatedAt.Format(time.RFC3339))
	return err
}

func (r *sqliteNotificationRepository) GetNotificationByID(id uuid.UUID) (*models.Notification, error) {
	query := `SELECT id, user_id, type, source_id, group_id, is_read, created_at FROM notifications WHERE id = ?`
	row := r.db.QueryRow(query, id.String())

	var (
		rawID, rawUserID, rawSourceID string
		rawGroupID                    sql.NullString
		n                             models.Notification
		isReadInt                     int
		createdAt                     string
	)

	err := row.Scan(&rawID, &rawUserID, &n.Type, &rawSourceID, &rawGroupID, &isReadInt, &createdAt)
	if err != nil {
		return nil, err
	}

	n.ID, _ = uuid.FromString(rawID)
	n.UserID, _ = uuid.FromString(rawUserID)
	n.SourceID, _ = uuid.FromString(rawSourceID)
	n.GroupID = parseNullableUUID(rawGroupID)
	n.IsRead = isReadInt == 1

	parsedCreatedAt, err := parseSQLiteTime(createdAt)
	if err != nil {
		return nil, err
	}
	n.CreatedAt = parsedCreatedAt

	return &n, nil
}

func (r *sqliteNotificationRepository) ListNotificationsByUser(userID uuid.UUID) ([]*models.Notification, error) {
	query := `SELECT id, user_id, type, source_id, group_id, is_read, created_at FROM notifications WHERE user_id = ? ORDER BY created_at DESC`
	rows, err := r.db.Query(query, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*models.Notification
	for rows.Next() {
		var (
			rawID, rawUserID, rawSourceID string
			rawGroupID                    sql.NullString
			n                             models.Notification
			isReadInt                     int
			createdAt                     string
		)

		if err := rows.Scan(&rawID, &rawUserID, &n.Type, &rawSourceID, &rawGroupID, &isReadInt, &createdAt); err != nil {
			return nil, err
		}

		n.ID, _ = uuid.FromString(rawID)
		n.UserID, _ = uuid.FromString(rawUserID)
		n.SourceID, _ = uuid.FromString(rawSourceID)
		n.GroupID = parseNullableUUID(rawGroupID)
		n.IsRead = isReadInt == 1

		parsedCreatedAt, err := parseSQLiteTime(createdAt)
		if err != nil {
			return nil, err
		}
		n.CreatedAt = parsedCreatedAt

		notifications = append(notifications, &n)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *sqliteNotificationRepository) MarkAsRead(id uuid.UUID) error {
	query := `UPDATE notifications SET is_read = 1 WHERE id = ?`
	_, err := r.db.Exec(query, id.String())
	return err
}

func (r *sqliteNotificationRepository) MarkAllAsRead(userID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = 1 WHERE user_id = ?`
	_, err := r.db.Exec(query, userID.String())
	return err
}

// parseNullableUUID converts a nullable SQL string column into a *uuid.UUID,
// returning nil when the column was NULL or fails to parse.
func parseNullableUUID(raw sql.NullString) *uuid.UUID {
	if !raw.Valid {
		return nil
	}
	id, err := uuid.FromString(raw.String)
	if err != nil {
		return nil
	}
	return &id
}
