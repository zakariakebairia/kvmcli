package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

// DBHandler handles all DB operations for any resource type.
// One struct replaces vm_record.go + network_record.go + store_record.go.
type DBHandler struct {
	db *sql.DB
}

func NewDBHandler(db *sql.DB) *DBHandler {
	return &DBHandler{db: db}
}

// EnsureTable creates the unified resources table.
// Instead of three separate tables (vms, networks, stores), we use ONE table.
func (s *DBHandler) EnsureTable(ctx context.Context) error {
	const schema = `
    CREATE TABLE IF NOT EXISTS resources (
        id         INTEGER PRIMARY KEY AUTOINCREMENT,
        type       TEXT NOT NULL,
        name       TEXT NOT NULL,
        namespace  TEXT NOT NULL DEFAULT '',
        labels     TEXT DEFAULT '{}',
        attrs      TEXT DEFAULT '{}',
        status     TEXT DEFAULT '',
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_type_name_ns
        ON resources(type, name, namespace);
    `
	_, err := s.db.ExecContext(ctx, schema)
	return err
}

func (s *DBHandler) Get(
	ctx context.Context,
	typeName, name, namespace string,
) (*registry.Object, error) {
	const query = `
    SELECT type, name, namespace, labels, attrs, status
    FROM resources
    WHERE type = ? AND name = ? AND namespace = ?
    `
	var labelsRaw, attrsRaw string
	state := &registry.Object{}
	err := s.db.QueryRowContext(ctx, query, typeName, name, namespace).Scan(
		&state.TypeName, &state.Name, &state.Namespace, &labelsRaw, &attrsRaw, &state.Status,
	)
	if err == sql.ErrNoRows {
		return nil, nil // not found = no current state
	}
	if err != nil {
		return nil, fmt.Errorf("get state: %w", err)
	}

	json.Unmarshal([]byte(labelsRaw), &state.Labels)
	json.Unmarshal([]byte(attrsRaw), &state.Attrs)
	return state, nil
}

// List retrieves all resources of a given type (or all types if typeName is "").
func (s *DBHandler) List(ctx context.Context, typeName string) ([]registry.Object, error) {
	// ...
	return nil, nil
}

// Put inserts or updates a resource state.
func (s *DBHandler) Put(ctx context.Context, state *registry.Object) error {
	labelsJSON, _ := json.Marshal(state.Labels)
	attrsJSON, _ := json.Marshal(state.Attrs)

	const query = `
        INSERT INTO resources (type, name, namespace, labels, attrs, status)
        VALUES (?, ?, ?, ?, ?, ?)
        ON CONFLICT(type, name, namespace) DO UPDATE SET
            labels = excluded.labels,
            attrs = excluded.attrs,
            status = excluded.status,
            updated_at = CURRENT_TIMESTAMP
    `
	_, err := s.db.ExecContext(ctx, query,
		state.TypeName, state.Name, state.Namespace,
		string(labelsJSON), string(attrsJSON), state.Status,
	)
	return err
}

// Remove deletes a resource state.
func (s *DBHandler) Remove(ctx context.Context, typeName, name, namespace string) error {
	const query = `DELETE FROM resources WHERE type = ? AND name = ? AND namespace = ?`
	_, err := s.db.ExecContext(ctx, query, typeName, name, namespace)
	return err
}
