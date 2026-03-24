package config

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/digitalocean/go-libvirt"
	"github.com/zakariakebairia/kvmcli/internal/network"
	"github.com/zakariakebairia/kvmcli/internal/registry"
	"github.com/zakariakebairia/kvmcli/internal/resources"
	"github.com/zakariakebairia/kvmcli/internal/store"
	"github.com/zakariakebairia/kvmcli/internal/vms"
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

// BuildStoreObjects converts store HCL configs into registry Objects.
// This is the new path for stores — the old store.NewStore path will be removed
// once the migration is complete.
func BuildStoreObjects(cfg *Config) []registry.Object {
	var out []registry.Object
	for _, s := range cfg.Stores {
		images := make([]map[string]any, 0, len(s.Images))
		for _, img := range s.Images {
			images = append(images, map[string]any{
				"name":       img.Name,
				"display":    img.Display,
				"version":    img.Version,
				"os_profile": img.OSProfile,
				"file":       img.File,
				"checksum":   img.Checksum,
				"size":       img.Size,
			})
		}
		out = append(out, registry.Object{
			TypeName:  "store",
			Name:      s.Name,
			Namespace: s.Namespace,
			Labels:    s.Labels,
			Attrs: map[string]any{
				"backend":        s.Backend,
				"artifacts_path": s.Paths.Artifacts,
				"images_path":    s.Paths.Images,
				"images":         images,
			},
		})
	}
	return out
}

// BuildNetworkObjects converts network HCL configs into registry Objects.
func BuildNetworkObjects(cfg *Config) []registry.Object {
	var out []registry.Object
	for _, n := range cfg.Networks {
		attrs := map[string]any{
			"bridge":      n.Bridge,
			"mode":        n.Mode,
			"net_address": n.NetAddress,
			"netmask":     n.NetMask,
			"autostart":   n.Autostart,
		}
		if n.DHCP != nil {
			attrs["dhcp"] = map[string]any{
				"start": n.DHCP.Start,
				"end":   n.DHCP.End,
			}
		}
		out = append(out, registry.Object{
			TypeName:  "network",
			Name:      n.Name,
			Namespace: n.Namespace,
			Labels:    n.Labels,
			Attrs:     attrs,
		})
	}
	return out
}
