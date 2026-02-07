package localstore

import (
	"context"
	"database/sql"
	"time"

	domain "github.com/felixgeelhaar/acai/internal/domain/meeting"
)

// WriteRepository implements domain.WriteRepository using SQLite.
// It stores local overrides for action items (text, completion state).
type WriteRepository struct {
	db *sql.DB
}

// NewWriteRepository creates a new SQLite-backed write repository.
func NewWriteRepository(db *sql.DB) *WriteRepository {
	return &WriteRepository{db: db}
}

func (r *WriteRepository) SaveActionItemState(_ context.Context, item *domain.ActionItem) error {
	var completed int
	if item.IsCompleted() {
		completed = 1
	}
	_, err := r.db.Exec(
		`INSERT OR REPLACE INTO action_item_overrides
			(action_item_id, meeting_id, text, completed, updated_at)
		VALUES (?, ?, ?, ?, ?)`,
		string(item.ID()), string(item.MeetingID()), item.Text(), completed, time.Now().UTC(),
	)
	return err
}

func (r *WriteRepository) GetLocalActionItemState(_ context.Context, id domain.ActionItemID) (*domain.ActionItem, error) {
	var (
		actionItemID string
		meetingID    string
		text         sql.NullString
		completed    sql.NullInt64
	)
	err := r.db.QueryRow(
		"SELECT action_item_id, meeting_id, text, completed FROM action_item_overrides WHERE action_item_id = ?",
		string(id),
	).Scan(&actionItemID, &meetingID, &text, &completed)
	if err == sql.ErrNoRows {
		return nil, domain.ErrMeetingNotFound
	}
	if err != nil {
		return nil, err
	}

	itemText := ""
	if text.Valid {
		itemText = text.String
	}

	item, err := domain.NewActionItem(
		domain.ActionItemID(actionItemID),
		domain.MeetingID(meetingID),
		"", // owner not stored in overrides
		itemText,
		nil,
	)
	if err != nil {
		return nil, err
	}

	if completed.Valid && completed.Int64 == 1 {
		item.Complete()
	}

	return item, nil
}

var _ domain.WriteRepository = (*WriteRepository)(nil)
