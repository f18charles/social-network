package repositories

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type sqliteMessageRepository struct {
	db *sql.DB
}

func NewMessageRepository(db *sql.DB) MessageRepository {
	return &sqliteMessageRepository{db: db}
}

func (r *sqliteMessageRepository) CreateMessage(message *models.Message) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `INSERT INTO messages (id, sender_id, dm_thread_id, group_id, content, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(
		query,
		message.ID.String(),
		message.SenderID.String(),
		nullableUUIDArg(message.DMThreadID),
		nullableUUIDArg(message.GroupID),
		message.Content,
		message.CreatedAt.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}

	if message.DMThreadID != nil {
		updateQuery := `UPDATE dm_threads SET last_message_at = ? WHERE id = ?`
		_, err = tx.Exec(updateQuery, message.CreatedAt.Format(time.RFC3339), message.DMThreadID.String())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqliteMessageRepository) GetMessageByID(id uuid.UUID) (*models.Message, error) {
	query := `SELECT id, sender_id, dm_thread_id, group_id, content, created_at FROM messages WHERE id = ?`
	row := r.db.QueryRow(query, id.String())

	var (
		rawID, rawSenderID        string
		rawDMThreadID, rawGroupID sql.NullString
		m                         models.Message
		createdAt                 string
	)

	err := row.Scan(&rawID, &rawSenderID, &rawDMThreadID, &rawGroupID, &m.Content, &createdAt)
	if err != nil {
		return nil, err
	}

	m.ID, _ = uuid.FromString(rawID)
	m.SenderID, _ = uuid.FromString(rawSenderID)
	m.DMThreadID, _ = nullableUUID(rawDMThreadID)
	m.GroupID, _ = nullableUUID(rawGroupID)

	parsedCreatedAt, err := parseSQLiteTime(createdAt)
	if err != nil {
		return nil, err
	}
	m.CreatedAt = parsedCreatedAt

	return &m, nil
}

func (r *sqliteMessageRepository) ListMessagesByGroup(groupID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `SELECT id, sender_id, dm_thread_id, group_id, content, created_at FROM messages WHERE group_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(query, groupID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

func (r *sqliteMessageRepository) ListMessagesByThread(threadID uuid.UUID, limit, offset int) ([]*models.Message, error) {
	query := `SELECT id, sender_id, dm_thread_id, group_id, content, created_at FROM messages WHERE dm_thread_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(query, threadID.String(), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanMessages(rows)
}

func (r *sqliteMessageRepository) GetOrCreateDMThread(user1ID, user2ID uuid.UUID) (*models.DMThread, error) {
	// Order IDs to ensure uniqueness constraint user1_id < user2_id is preserved if necessary
	id1, id2 := user1ID, user2ID
	if id1.String() > id2.String() {
		id1, id2 = id2, id1
	}

	query := `SELECT id, user1_id, user2_id, last_message_at FROM dm_threads WHERE user1_id = ? AND user2_id = ?`
	row := r.db.QueryRow(query, id1.String(), id2.String())

	var (
		rawID, rawUser1ID, rawUser2ID string
		t                             models.DMThread
		lastMsgAt                     string
	)

	err := row.Scan(&rawID, &rawUser1ID, &rawUser2ID, &lastMsgAt)
	if err == nil {
		t.ID, _ = uuid.FromString(rawID)
		t.User1ID, _ = uuid.FromString(rawUser1ID)
		t.User2ID, _ = uuid.FromString(rawUser2ID)
		t.LastMessageAt, _ = parseSQLiteTime(lastMsgAt)
		return &t, nil
	}

	if err != sql.ErrNoRows {
		return nil, err
	}

	// Create new thread
	newID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	insertQuery := `INSERT INTO dm_threads (id, user1_id, user2_id, last_message_at) VALUES (?, ?, ?, ?)`
	_, err = r.db.Exec(insertQuery, newID.String(), id1.String(), id2.String(), now.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	return &models.DMThread{
		ID:            newID,
		User1ID:       id1,
		User2ID:       id2,
		LastMessageAt: now,
	}, nil
}

func (r *sqliteMessageRepository) GetDMThreadByID(id uuid.UUID) (*models.DMThread, error) {
	query := `SELECT id, user1_id, user2_id, last_message_at FROM dm_threads WHERE id = ?`
	row := r.db.QueryRow(query, id.String())

	var (
		rawID, rawUser1ID, rawUser2ID string
		t                             models.DMThread
		lastMsgAt                     string
	)

	err := row.Scan(&rawID, &rawUser1ID, &rawUser2ID, &lastMsgAt)
	if err != nil {
		return nil, err
	}

	t.ID, _ = uuid.FromString(rawID)
	t.User1ID, _ = uuid.FromString(rawUser1ID)
	t.User2ID, _ = uuid.FromString(rawUser2ID)
	t.LastMessageAt, _ = parseSQLiteTime(lastMsgAt)
	return &t, nil
}

func (r *sqliteMessageRepository) ListConversations(userID uuid.UUID) ([]*models.Conversation, error) {
	query := `
		SELECT 
			t.id AS thread_id, 
			NULL AS group_id, 
			'dm' AS type, 
			u.first_name || ' ' || u.last_name AS target_name, 
			COALESCE(u.avatar, '') AS target_avatar, 
			COALESCE(m.content, '') AS last_message, 
			t.last_message_at AS last_message_at
		FROM dm_threads t
		JOIN users u ON u.id = CASE WHEN t.user1_id = ? THEN t.user2_id ELSE t.user1_id END
		LEFT JOIN (
			SELECT dm_thread_id, content, MAX(created_at) FROM messages WHERE dm_thread_id IS NOT NULL GROUP BY dm_thread_id
		) m ON m.dm_thread_id = t.id
		WHERE t.user1_id = ? OR t.user2_id = ?

		UNION ALL

		SELECT 
			NULL AS thread_id, 
			g.id AS group_id, 
			'group' AS type, 
			g.title AS target_name, 
			'' AS target_avatar, 
			COALESCE(m.content, '') AS last_message, 
			COALESCE(m.created_at, g.created_at) AS last_message_at
		FROM groups g
		JOIN group_members gm ON gm.group_id = g.id AND gm.user_id = ? AND gm.status = 'accepted'
		LEFT JOIN (
			SELECT group_id, content, created_at, MAX(created_at) FROM messages WHERE group_id IS NOT NULL GROUP BY group_id
		) m ON m.group_id = g.id

		ORDER BY last_message_at DESC
	`

	rows, err := r.db.Query(query, userID.String(), userID.String(), userID.String(), userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*models.Conversation
	for rows.Next() {
		var (
			rawThreadID, rawGroupID sql.NullString
			c                       models.Conversation
			lastMsgAt               string
		)

		err := rows.Scan(&rawThreadID, &rawGroupID, &c.Type, &c.TargetName, &c.TargetAvatar, &c.LastMessage, &lastMsgAt)
		if err != nil {
			return nil, err
		}

		if rawThreadID.Valid && rawThreadID.String != "" {
			parsedID, _ := uuid.FromString(rawThreadID.String)
			c.ThreadID = &parsedID
		}
		if rawGroupID.Valid && rawGroupID.String != "" {
			parsedID, _ := uuid.FromString(rawGroupID.String)
			c.GroupID = &parsedID
		}

		c.LastMessageAt, _ = parseSQLiteTime(lastMsgAt)
		conversations = append(conversations, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *sqliteMessageRepository) scanMessages(rows *sql.Rows) ([]*models.Message, error) {
	var messages []*models.Message
	for rows.Next() {
		var (
			rawID, rawSenderID        string
			rawDMThreadID, rawGroupID sql.NullString
			m                         models.Message
			createdAt                 string
		)

		if err := rows.Scan(&rawID, &rawSenderID, &rawDMThreadID, &rawGroupID, &m.Content, &createdAt); err != nil {
			return nil, err
		}

		m.ID, _ = uuid.FromString(rawID)
		m.SenderID, _ = uuid.FromString(rawSenderID)
		m.DMThreadID, _ = nullableUUID(rawDMThreadID)
		m.GroupID, _ = nullableUUID(rawGroupID)

		parsedCreatedAt, err := parseSQLiteTime(createdAt)
		if err != nil {
			return nil, err
		}
		m.CreatedAt = parsedCreatedAt

		messages = append(messages, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
