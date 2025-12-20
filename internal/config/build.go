package config

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/digitalocean/go-libvirt"
	"github.com/kebairia/kvmcli/internal/network"
	"github.com/kebairia/kvmcli/internal/resources"
	"github.com/kebairia/kvmcli/internal/store"
	"github.com/kebairia/kvmcli/internal/vms"
)

func BuildResources(
	cfg *Config,
	ctx context.Context,
	db *sql.DB,
	conn *libvirt.Libvirt,
) ([]resources.Resource, error) {
	var out []resources.Resource
	if cfg == nil {
		return nil, fmt.Errorf("BuildResources: nil config")
	}

	netMgr := network.NewLibvirtNetworkManager(conn, db)
	storeMgr := store.NewDBStoreManager(db)

	for _, n := range cfg.Networks {
		out = append(out, network.NewNetwork(n, netMgr, ctx))
	}

	for _, s := range cfg.Stores {
		out = append(out, store.NewStore(s, storeMgr, ctx))
	}

	for _, v := range cfg.VMs {
		vm, err := vms.NewVirtualMachine(
			v,
			vms.WithContext(ctx),
			vms.WithDatabaseConnection(db),
			vms.WithLibvirtConnection(conn),
		)
		if err != nil {
			return nil, err
		}
		out = append(out, vm)
	}

	return out, nil
}
