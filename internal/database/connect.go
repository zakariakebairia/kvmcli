package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

const (
	databaseName   = "kvmcli"
	storesTable    = "stores"
	imagesTable    = "images"
	vmsTable       = "vms"
	networksTable  = "networks"
	snapshotsTable = "snapshots"
)

// InitDB opens a database handle and verifies the connection using context.
// It returns a ready-to-use *sql.DB or an error if the connection fails.
func InitDB(ctx context.Context, dbPath string) (*sql.DB, error) {
	// Define a 5-second timeout context for DB operations
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	if dbPath == "" {
		return nil, fmt.Errorf("database path is empty")
	}

	// Open a handle to the SQLite database (does not connect yet)
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB handle: %w", err)
	}

	// Verify the connection with PingContext (this actually connects)
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to DB at %s: %w", dbPath, err)
	}

	return db, nil
}
