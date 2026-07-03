package repositories

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type sqliteMessageReactionRepository struct {
	db *sql.DB
}

func NewMessageReactionRepository(db *sql.DB) MessageReactionRepository {
	return &sqliteMessageReactionRepository{db: db}
}

func (r *sqliteMessageReactionRepository) SetMessageReaction(messageID, userID uuid.UUID, emoji string) error {
	query := `
		INSERT INTO message_reactions (message_id, user_id, emoji, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(message_id, user_id, emoji) DO NOTHING
	`
	_, err := r.db.Exec(query, messageID.String(), userID.String(), emoji, time.Now().UTC())
	return err
}

func (r *sqliteMessageReactionRepository) DeleteMessageReaction(messageID, userID uuid.UUID, emoji string) error {
	query := `
		DELETE FROM message_reactions
		WHERE message_id = ? AND user_id = ? AND emoji = ?
	`
	_, err := r.db.Exec(query, messageID.String(), userID.String(), emoji)
	return err
}

func (r *sqliteMessageReactionRepository) GetMessageReactionSummary(messageID, viewerID uuid.UUID) ([]*models.MessageReactionSummary, error) {
	query := `
		SELECT 
			emoji, 
			COUNT(*), 
			MAX(CASE WHEN user_id = ? THEN 1 ELSE 0 END)
		FROM message_reactions
		WHERE message_id = ?
		GROUP BY emoji
	`
	rows, err := r.db.Query(query, viewerID.String(), messageID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.MessageReactionSummary
	for rows.Next() {
		var summary models.MessageReactionSummary
		var userReactedInt int
		if err := rows.Scan(&summary.Emoji, &summary.Count, &userReactedInt); err != nil {
			return nil, err
		}
		summary.UserReacted = userReactedInt == 1
		summaries = append(summaries, &summary)
	}
	if summaries == nil {
		return []*models.MessageReactionSummary{}, nil
	}
	return summaries, nil
}
