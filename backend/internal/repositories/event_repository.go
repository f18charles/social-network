package repositories

import (
	"database/sql"
	"time"

	"github.com/gofrs/uuid/v5"
	"learn.zone01kisumu.ke/git/qquinton/social-network/internal/models"
)

type sqliteEventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return &sqliteEventRepository{db: db}
}

func (r *sqliteEventRepository) CreateEvent(event *models.Event) error {
	query := `INSERT INTO events (id, group_id, creator_id, title, description, event_date, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := r.db.Exec(query, event.ID.String(), event.GroupID.String(), event.CreatorID.String(), event.Title, event.Description, event.EventDate.Format(time.RFC3339), event.CreatedAt.Format(time.RFC3339))
	return err
}

func (r *sqliteEventRepository) GetEventByID(id uuid.UUID) (*models.Event, error) {
	query := `SELECT id, group_id, creator_id, title, description, event_date, created_at FROM events WHERE id = ?`
	row := r.db.QueryRow(query, id.String())

	var (
		rawID, rawGroupID    string
		rawCreatorID         sql.NullString
		e                    models.Event
		desc                 sql.NullString
		eventDate, createdAt string
	)

	err := row.Scan(&rawID, &rawGroupID, &rawCreatorID, &e.Title, &desc, &eventDate, &createdAt)
	if err != nil {
		return nil, err
	}

	e.ID, _ = uuid.FromString(rawID)
	e.GroupID, _ = uuid.FromString(rawGroupID)
	if rawCreatorID.Valid && rawCreatorID.String != "" {
		e.CreatorID, _ = uuid.FromString(rawCreatorID.String)
	} else {
		e.CreatorID = uuid.Nil
	}
	e.Description = desc.String

	parsedEventDate, err := parseSQLiteTime(eventDate)
	if err != nil {
		return nil, err
	}
	e.EventDate = parsedEventDate

	parsedCreatedAt, err := parseSQLiteTime(createdAt)
	if err != nil {
		return nil, err
	}
	e.CreatedAt = parsedCreatedAt

	return &e, nil
}

func (r *sqliteEventRepository) ListEventsByGroup(groupID uuid.UUID) ([]*models.Event, error) {
	query := `SELECT id, group_id, creator_id, title, description, event_date, created_at FROM events WHERE group_id = ? ORDER BY event_date ASC`
	rows, err := r.db.Query(query, groupID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*models.Event
	for rows.Next() {
		var (
			rawID, rawGroupID    string
			rawCreatorID         sql.NullString
			e                    models.Event
			desc                 sql.NullString
			eventDate, createdAt string
		)

		if err := rows.Scan(&rawID, &rawGroupID, &rawCreatorID, &e.Title, &desc, &eventDate, &createdAt); err != nil {
			return nil, err
		}

		e.ID, _ = uuid.FromString(rawID)
		e.GroupID, _ = uuid.FromString(rawGroupID)
		if rawCreatorID.Valid && rawCreatorID.String != "" {
			e.CreatorID, _ = uuid.FromString(rawCreatorID.String)
		} else {
			e.CreatorID = uuid.Nil
		}
		e.Description = desc.String

		parsedEventDate, err := parseSQLiteTime(eventDate)
		if err != nil {
			return nil, err
		}
		e.EventDate = parsedEventDate

		parsedCreatedAt, err := parseSQLiteTime(createdAt)
		if err != nil {
			return nil, err
		}
		e.CreatedAt = parsedCreatedAt

		events = append(events, &e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *sqliteEventRepository) GetRSVP(eventID, userID uuid.UUID) (string, error) {
	var status string
	err := r.db.QueryRow(`SELECT status FROM event_rsvps WHERE event_id = ? AND user_id = ?`, eventID.String(), userID.String()).Scan(&status)
	if err == sql.ErrNoRows {
		return "none", nil
	}
	return status, err
}

func (r *sqliteEventRepository) SetRSVP(eventID, userID uuid.UUID, status string) error {
	query := `INSERT INTO event_rsvps (event_id, user_id, status) VALUES (?, ?, ?)
		ON CONFLICT(event_id, user_id) DO UPDATE SET status = excluded.status`
	_, err := r.db.Exec(query, eventID.String(), userID.String(), status)
	return err
}

func (r *sqliteEventRepository) GetRSVPSummaries(eventID uuid.UUID) (going int, notGoing int, err error) {
	err = r.db.QueryRow(`SELECT COUNT(*) FROM event_rsvps WHERE event_id = ? AND status = 'going'`, eventID.String()).Scan(&going)
	if err != nil {
		return 0, 0, err
	}
	err = r.db.QueryRow(`SELECT COUNT(*) FROM event_rsvps WHERE event_id = ? AND status = 'not_going'`, eventID.String()).Scan(&notGoing)
	if err != nil {
		return 0, 0, err
	}
	return going, notGoing, nil
}
