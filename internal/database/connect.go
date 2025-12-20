package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Constants
// IDEA: the path of the db file muust be declared in a config file
//			 in /etc/kvmcli/kvmcli.toml for example
// IDEA: path for database file: /var/lib/kvmcli/kvmcli.db

const (
	DBFilePath     = "/home/zakaria/dox/homelab/kvmcli/kvmcli.db"
	databaseName   = "kvmcli"
	storesTable    = "stores"
	imagesTable    = "images"
	vmsTable       = "vms"
	networksTable  = "networks"
	snapshotsTable = "snapshots"
)

// InitDB opens a database handle and verifies the connection using context.
// It returns a ready-to-use *sql.DB or an error if the connection fails.
func InitDB(ctx context.Context) (*sql.DB, error) {
	// Define a 5-second timeout context for DB operations
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// Open a handle to the SQLite database (does not connect yet)
	db, err := sql.Open("sqlite3", DBFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DB handle: %w", err)
	}

	// Verify the connection with PingContext (this actually connects)
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	return db, nil
}
