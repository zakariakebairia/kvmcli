package network

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/digitalocean/go-libvirt"
	"github.com/zakariakebairia/kvmcli/internal/common"
	db "github.com/zakariakebairia/kvmcli/internal/database"
	log "github.com/zakariakebairia/kvmcli/internal/logger"
)

const (
	// Network state constants as defined by libvirt.
	networkStateActive   = 1
	networkStateInactive = 0
	networkStateUnknown  = "unknown"
)

type VirtualNetworkInfo struct {
	Name      string
	State     string
	Bridge    string
	Subnet    string
	Gateway   string
	DHCPRange string
	Age       string
}

func (info *VirtualNetworkInfo) Header() *tabwriter.Writer {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "NAME\tSTATE\tBRIDGE\tSUBNET\tGATEWAY\tDHCP RANGE\tAGE")
	return w
}

func (info *VirtualNetworkInfo) PrintInfo(w *tabwriter.Writer) {
	fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
		info.Name,
		info.State,
		info.Bridge,
		info.Subnet,
		info.Gateway,
		info.DHCPRange,
		info.Age,
	)
}

// NewVirtualMachineInfo constructs a VirtualMachineInfo by querying libvirt and the DB.
func NewVirtualNetworkInfo(
	ctx context.Context,
	database *sql.DB,
	conn *libvirt.Libvirt,
	record db.VirtualNetwork,
) (*VirtualNetworkInfo, error) {
	// Domain lookup
	net, err := conn.NetworkLookupByName(record.Name)
	if err != nil {
		return nil, fmt.Errorf("lookup network %q: %w", record.Name, err)
	}

	// State
	state, err := getState(conn, net)
	if err != nil {
		log.Errorf("cannot get state for %q: %v", record.Name, err)
		state = networkStateUnknown
	}

	// Format the DHCP range.
	var dhcpRange string
	if record.DHCP != nil {
		dhcpRange = record.DHCP["start"] + " → " + record.DHCP["end"]
	}

	return &VirtualNetworkInfo{
		Name:      record.Name,
		State:     state,
		Bridge:    record.Bridge,
		Subnet:    record.Netmask,
		Gateway:   record.NetAddress,
		DHCPRange: dhcpRange,
		Age:       common.FormatAge(record.CreatedAt),
	}, nil
}

func GetVirtualNetworks(
	ctx context.Context,
	database *sql.DB,
	conn *libvirt.Libvirt,
) ([]VirtualNetworkInfo, error) {
	records, err := db.GetNetworks(ctx, database, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get network records  %w", err)
	}

	networks := make([]VirtualNetworkInfo, 0, len(records))

	for _, rec := range records {
		netInfo, err := NewVirtualNetworkInfo(ctx, database, conn, rec)
		if err != nil {
			log.Errorf("could not build VM info for %q: %v", rec.Name, err)
			continue
		}
		networks = append(networks, *netInfo)
	}

	return networks, nil
}

// getState retrieves the active/inactive state of the network using libvirt.
func getState(conn *libvirt.Libvirt, network libvirt.Network) (string, error) {
	state, err := conn.NetworkIsActive(network)
	if err != nil {
		return "", fmt.Errorf("failed to get state for network %s: %w", network.Name, err)
	}

	switch int(state) {
	case networkStateActive:
		return "Active", nil
	case networkStateInactive:
		return "Inactive", nil
	default:
		return "Unknown", nil
	}
}
