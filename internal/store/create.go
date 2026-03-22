package store

import (
	// "context"
	// "database/sql"
	"context"
	"fmt"
	// db "github.com/zakariakebairia/kvmcli/internal/database"
)

// Create persists the store to the database.
func (m *DBStoreManager) Create(ctx context.Context, spec Config) error {
	record := NewStoreRecord(spec)
	err := record.Insert(ctx, m.db)
	if err != nil {
		return fmt.Errorf("failed to insert new store record: %w", err)
	}
	fmt.Printf("store/%s created\n", spec.Name)
	return nil
}
