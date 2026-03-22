package store

import (
	"context"
	"fmt"

	db "github.com/zakariakebairia/kvmcli/internal/database"
)

// Delete removes the store from the database.
func (m *DBStoreManager) Delete(ctx context.Context, name, namespace string) error {
	// Construct a partial record or struct that supports Delete
	// db.Store has Delete method, we need Name and Namespace.
	// Assuming NewStoreRecord creates a full record but we only need name/namespace for delete.
	// Or we can construct &db.Store{Name: name, Namespace: namespace}.
	// But NewStoreRecord expects Config. We can make a dummy config.
	// Better: &db.Store{...}

	// Check if db.Store has Delete method. Yes.
	record := &db.Store{
		Name:      name,
		Namespace: namespace,
	}

	// Delete the store record from the database.
	if err := record.Delete(ctx, m.db); err != nil {
		return fmt.Errorf("failed to delete record for store %q: %v", name, err)
	}

	fmt.Printf("store/%s deleted\n", name)
	return nil
}
