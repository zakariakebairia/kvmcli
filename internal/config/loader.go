package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/digitalocean/go-libvirt"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/kebairia/kvmcli/internal/resources"
)

// Load parses and decodes the configuration file at the given path.
func Load(
	path string,
	ctx context.Context,
	db *sql.DB,
	conn *libvirt.Libvirt,
) ([]resources.Resource, error) {
	cfg, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}

	idx, err := BuildIndex(cfg)
	if err != nil {
		return nil, err
	}

	evalCtx, err := BuildEvalContext(cfg, idx, ctx, db)
	if err != nil {
		return nil, err
	}

	if err := ResolveAll(cfg, idx, evalCtx); err != nil {
		return nil, err
	}

	return BuildResources(cfg, ctx, db, conn)
}

func LoadConfig(path string) (*Config, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %q: %w", path, err)
	}

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCL(src, path)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse hcl %q: %w", path, diags)
	}

	var cfg Config
	if diags := gohcl.DecodeBody(file.Body, nil, &cfg); diags.HasErrors() {
		return nil, fmt.Errorf("decode hcl %q: %w", path, diags)
	}
	return &cfg, nil
}
