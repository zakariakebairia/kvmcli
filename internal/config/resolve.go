package config

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/kebairia/kvmcli/internal/vms"
	"github.com/zclconf/go-cty/cty"
)

func ResolveAll(cfg *Config, idx *Index, evalCtx *hcl.EvalContext) error {
	// Build data maps for validation (we assume they are validated by BuildEvalContext)
	dataNetworks := make(map[string]struct{})
	dataStores := make(map[string]struct{})

	for _, d := range cfg.Data {
		switch d.Type {
		case "network":
			dataNetworks[d.Name] = struct{}{}
		case "store":
			dataStores[d.Name] = struct{}{}
		}
	}

	for i := range cfg.VMs {
		if err := resolveVM(&cfg.VMs[i], idx, dataNetworks, dataStores, evalCtx); err != nil {
			return err
		}
	}
	return nil
}

func resolveVM(
	vm *vms.Config,
	idx *Index,
	dataNetworks map[string]struct{},
	dataStores map[string]struct{},
	ctx *hcl.EvalContext,
) error {
	// Resolve Network
	if err := resolveVMNetwork(vm, idx.Networks, dataNetworks, ctx); err != nil {
		return err
	}

	// Resolve Store
	if err := resolveVMStore(vm, idx.Stores, dataStores, ctx); err != nil {
		return err
	}

	return nil
}

func resolveVMNetwork(
	vm *vms.Config,
	networks map[string]struct{},
	dataNetworks map[string]struct{},
	ctx *hcl.EvalContext,
) error {
	net, err := evalString(vm.NetExpr, ctx, "network", vm.Name)
	if err != nil {
		return err
	}

	_, local := networks[net]
	_, data := dataNetworks[net]

	if !local && !data {
		return fmt.Errorf("vm.%s.network: unknown network %q", vm.Name, net)
	}

	vm.NetName = net
	return nil
}

func resolveVMStore(
	vm *vms.Config,
	stores map[string]struct{},
	dataStores map[string]struct{},
	ctx *hcl.EvalContext,
) error {
	st, err := evalString(vm.StoreExpr, ctx, "store", vm.Name)
	if err != nil {
		return err
	}

	_, local := stores[st]
	_, data := dataStores[st]

	if !local && !data {
		return fmt.Errorf("vm.%s.store: unknown store %q", vm.Name, st)
	}

	vm.Store = st
	return nil
}

func evalString(expr hcl.Expression, ctx *hcl.EvalContext, field, name string) (string, error) {
	if expr == nil {
		return "", fmt.Errorf("vm.%s.%s: missing attribute", name, field)
	}

	val, diags := expr.Value(ctx)
	if diags.HasErrors() {
		return "", fmt.Errorf("vm.%s.%s: %s", name, field, diags.Error())
	}
	if val.Type() != cty.String {
		return "", fmt.Errorf("vm.%s.%s: must be string", name, field)
	}
	return val.AsString(), nil
}
