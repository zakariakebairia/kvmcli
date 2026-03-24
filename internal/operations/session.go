package operations

import (
	"context"
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal"
	"github.com/zakariakebairia/kvmcli/internal/config"
	db "github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

// NewSession initialises the shared dependencies (libvirt, DB) and returns
// a registry.Session ready for use, plus a closer function to release them.
func NewSession(ctx context.Context, configPath string) (registry.Session, func(), error) {
	if ctx == nil {
		ctx = context.Background()
	}

	// Load global config to get the DB path
	cfg, err := config.LoadGlobal(configPath)
	if err != nil {
		return registry.Session{}, nil, fmt.Errorf("load global config: %w", err)
	}

	// Connect to libvirt
	conn, err := internal.InitConnection()
	if err != nil {
		return registry.Session{}, nil, fmt.Errorf("init libvirt: %w", err)
	}

	// Open the SQLite database
	database, err := db.InitDB(ctx, cfg.Paths.DB)
	if err != nil {
		_ = conn.Disconnect()
		return registry.Session{}, nil, fmt.Errorf("init database: %w", err)
	}

	regCtx := registry.Session{
		Ctx:  ctx,
		DB:   database,
		Conn: conn,
	}

	closer := func() {
		if conn != nil {
			_ = conn.Disconnect()
		}
		if database != nil {
			_ = database.Close()
		}
	}

	return regCtx, closer, nil
}
