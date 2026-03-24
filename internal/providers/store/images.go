package store

import (
	"context"
	"database/sql"
	"fmt"
)

// ensureImagesTable creates the images table if it doesn't exist.
// This is a store-specific concern — the images table is owned by the store provider,
// not by the generic state store.
func ensureImagesTable(ctx context.Context, db *sql.DB) error {
	const schema = `
	CREATE TABLE IF NOT EXISTS images (
		id         INTEGER PRIMARY KEY AUTOINCREMENT,
		store_name TEXT NOT NULL,
		store_ns   TEXT NOT NULL DEFAULT '',
		name       TEXT,
		display    TEXT,
		version    TEXT,
		os_profile TEXT,
		file       TEXT,
		checksum   TEXT,
		size       TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_images_store_name_ns
		ON images(store_name, store_ns);
	`
	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("ensure images table: %w", err)
	}
	return nil
}
