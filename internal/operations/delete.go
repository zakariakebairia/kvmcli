package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/zakariakebairia/kvmcli/internal/config"
	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/engine"
)

func DeleteFromManifest(manifestPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, cleanup, err := NewSession(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to create context: %w", err)
	}
	defer cleanup()

	dbHandler := database.NewDBHandler(session.DB)
	if err := dbHandler.EnsureTable(ctx); err != nil {
		return fmt.Errorf("ensure state table: %w", err)
	}

	objects, err := config.Load(manifestPath, session.Ctx, dbHandler)
	if err != nil {
		return fmt.Errorf("load config %q: %w", manifestPath, err)
	}

	eng := engine.New(session, dbHandler)
	return eng.Destroy(objects)
}
