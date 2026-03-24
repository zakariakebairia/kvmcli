package operations

import (
	"context"
	"fmt"
	"time"

	"github.com/zakariakebairia/kvmcli/internal/config"
	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/engine"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

func DeleteFromManifest(manifestPath string) error {
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
	eng := engine.New(session, dbHandler)

	var targets []registry.Object
	targets = append(targets, config.BuildNetworkObjects(cfg)...)
	targets = append(targets, config.BuildStoreObjects(cfg)...)
	return eng.Destroy(targets)
}
