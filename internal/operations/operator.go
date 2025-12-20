package operations

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/digitalocean/go-libvirt"
	"github.com/kebairia/kvmcli/internal"
	"github.com/kebairia/kvmcli/internal/config"
	db "github.com/kebairia/kvmcli/internal/database"
	"github.com/kebairia/kvmcli/internal/resources"
)

// Operator orchestrates lifecycle actions for any kvmcli Resource.
// It owns a database handle and a libvirt connection, plus context & logger.
type Operator struct {
	ctx  context.Context
	db   *sql.DB
	conn *libvirt.Libvirt
	cfg  *config.GlobalConfig
	// log  logger.Logger
}

// NewOperator initialises all shared dependencies.
//
// A nil ctx is promoted to context.Background().
// If any step fails, resources created earlier are released before the error
// is returned.
// NewOperator initialises all shared dependencies.
//
// A nil ctx is promoted to context.Background().
// If any step fails, resources created earlier are released before the error
// is returned.
// func NewOperator(ctx context.Context, log logger.Logger) (*Operator, error) {
func NewOperator(ctx context.Context, configPath string) (*Operator, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	// if log == nil {
	// 	log = logger.Log
	// }

	// Load global config
	cfg, err := config.LoadGlobal(configPath)
	if err != nil {
		return nil, fmt.Errorf("load global config: %w", err)
	}

	conn, err := internal.InitConnection()
	if err != nil {
		return nil, fmt.Errorf("init libvirt: %w", err)
	}

	// TODO:  sync.Once -> to ensure the database is initialized only once
	database, err := db.InitDB(ctx, cfg.Paths.DB)
	if err != nil {
		_ = conn.Disconnect()
		return nil, fmt.Errorf("init database: %w", err)
	}

	return &Operator{
		ctx:  ctx,
		db:   database,
		conn: conn,
		cfg:  cfg,
		// log:  log,
	}, nil
}

// Close releases all shared resources.
// It is safe to call multiple times.
func (o *Operator) Close() {
	if o.conn != nil {
		_ = o.conn.Disconnect()
	}
	if o.db != nil {
		_ = o.db.Close()
	}
}

// func (o *Operator) Get(resource resources.Resource) error {
// 	// If the resource requires a libvirt connection, inject it.
// 	if cs, ok := resource.(resources.ClientSetter); ok {
// 		cs.SetConnection(o.conn)
// 	}
// 	// get resource
// 	err := resource.Get()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// -- Helpers ------------------------------------------------------------
// setConnection provides libvirt connectivity to Resources that require it.
func (o *Operator) SetConnection(r resources.Resource) {
	if s, ok := r.(resources.ClientSetter); ok {
		s.SetConnection(o.ctx, o.db, o.conn)
	}
}
