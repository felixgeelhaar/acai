// Package localstore provides SQLite-backed local persistence for write operations.
// This is the write-side store for the CQRS split: reads go through the Granola API
// decorator chain, writes go directly to local SQLite.
package localstore

import "database/sql"

// InitSchema creates the local store tables if they don't exist.
func InitSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS agent_notes (
			id         TEXT PRIMARY KEY,
			meeting_id TEXT NOT NULL,
			author     TEXT NOT NULL,
			content    TEXT NOT NULL,
			created_at DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_agent_notes_meeting ON agent_notes(meeting_id);

		CREATE TABLE IF NOT EXISTS action_item_overrides (
			action_item_id TEXT PRIMARY KEY,
			meeting_id     TEXT NOT NULL,
			text           TEXT,
			completed      INTEGER,
			updated_at     DATETIME NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_action_item_overrides_meeting ON action_item_overrides(meeting_id);

		CREATE TABLE IF NOT EXISTS outbox_entries (
			id         TEXT PRIMARY KEY,
			event_type TEXT NOT NULL,
			payload    BLOB NOT NULL,
			status     TEXT NOT NULL DEFAULT 'pending',
			created_at DATETIME NOT NULL,
			synced_at  DATETIME,
			attempts   INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_outbox_status ON outbox_entries(status);
	`)
	return err
}
