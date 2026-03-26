package config

import "github.com/zakariakebairia/kvmcli/internal/registry"

// buildObjects converts all HCL resource configs into registry Objects.
// Order: networks and stores first (no deps), then VMs (depend on both).
func buildObjects(cfg *hclConfig) []registry.Object {
	var objects []registry.Object

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
		objects = append(objects, registry.Object{
			TypeName:  "network",
			Name:      n.Name,
			Namespace: n.Namespace,
			Labels:    n.Labels,
			Attrs:     attrs,
		})
	}

	for _, s := range cfg.Stores {
		images := make([]map[string]any, 0, len(s.Images))
		for _, image := range s.Images {
			images = append(images, map[string]any{
				"name":       image.Name,
				"display":    image.Display,
				"version":    image.Version,
				"os_profile": image.OSProfile,
				"file":       image.File,
				"checksum":   image.Checksum,
				"size":       image.Size,
			})
		}
		objects = append(objects, registry.Object{
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

	for _, v := range cfg.VMs {
		objects = append(objects, registry.Object{
			TypeName:  "vm",
			Name:      v.Name,
			Namespace: v.Namespace,
			Labels:    v.Labels,
			Attrs: map[string]any{
				"cpu":     v.CPU,
				"memory":  v.Memory,
				"disk":    v.Disk,
				"image":   v.Image,
				"network": v.NetName,
				"store":   v.Store,
				"ip":      v.IP,
				"mac":     v.MAC,
			},
		})
	}

	return objects
}
