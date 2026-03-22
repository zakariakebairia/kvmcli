package main

import (
	"github.com/zakariakebairia/kvmcli/cmd"
)

// IDEA: kvmcli <verb> <resource> [names…] [flags]
//  single VM
// kvmcli start vm    admin1
// kvmcli stop  vm    admin1
// kvmcli delete vm   admin1

// all VMs in namespace "admins"
// kvmcli start vms   --all -n admins
// kvmcli stop  vms   --all -n admins
// kvmcli delete vms  --all -n admins

//	networks
//
// kvmcli create network homelab --cidr=10.0.0.0/24
// kvmcli delete network homelab
//
// # stores
// kvmcli get    stores
// kvmcli delete store store-homelab
func main() {
	// Execute the CLI commands defined in the cmd package.
	cmd.Execute()
}
