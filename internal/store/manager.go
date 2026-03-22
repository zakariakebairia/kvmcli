package store

import (
	"context"
	"database/sql"

	db "github.com/zakariakebairia/kvmcli/internal/database"
)

// StoreManager defines the interface for managing image stores.
type StoreManager interface {
	Create(ctx context.Context, spec Config) error
	Delete(ctx context.Context, name, namespace string) error
	Get(ctx context.Context, name string) (*db.Store, error)
}

// DBStoreManager implements StoreManager using a SQL database.
type DBStoreManager struct {
	db *sql.DB
}

// NewDBStoreManager creates a new DBStoreManager.
func NewDBStoreManager(db *sql.DB) *DBStoreManager {
	return &DBStoreManager{
		db: db,
	}
}

// Create and Delete are implemented in create.go and delete.go

// Get retrieves the store record.
func (m *DBStoreManager) Get(ctx context.Context, name string) (*db.Store, error) {
	s, err := db.GetStoreByName(ctx, m.db, name)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
