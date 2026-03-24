package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/zakariakebairia/kvmcli/internal/config"
	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/engine"
	"github.com/zakariakebairia/kvmcli/internal/registry"

	// Blank imports so provider init() functions register resource types
	_ "github.com/zakariakebairia/kvmcli/internal/providers/network"
	_ "github.com/zakariakebairia/kvmcli/internal/providers/store"
)

func CreateFromManifest(manifestPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	session, cleanup, err := NewSession(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to create context: %w", err)
	}
	defer cleanup()

	cfg, err := config.LoadConfig(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load config %q: %w", manifestPath, err)
	}

	dbHandler := database.NewDBHandler(session.DB)
	if err := dbHandler.EnsureTable(ctx); err != nil {
		return fmt.Errorf("ensure state table: %w", err)
	}

	eng := engine.New(session, dbHandler)

	var desired []registry.Object
	desired = append(desired, config.BuildNetworkObjects(cfg)...)
	desired = append(desired, config.BuildStoreObjects(cfg)...)
	return eng.Apply(desired)
}
