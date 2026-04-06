package config

import (
	"context"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zclconf/go-cty/cty"
)

// resolve evaluates all HCL expressions in the config.
// It validates names, checks data sources against the DB,
// and fills in each VM's NetName and Store fields.
func resolve(cfg *hclConfig, ctx context.Context, dbHandler *database.DBHandler) error {
	// Collect names defined in this file and check for duplicates
	networks, err := collectNames(
		"network",
		cfg.Networks,
		func(n networkDef) string { return n.Name },
	)
	if err != nil {
		return err
	}

	stores, err := collectNames("store", cfg.Stores, func(s storeDef) string { return s.Name })
	if err != nil {
		return err
	}

	// Build the HCL eval context: the symbol table that lets
	// expressions like network.test or data.store.homelab evaluate
	evalCtx, err := buildEvalContext(cfg, networks, stores, ctx, dbHandler)
	if err != nil {
		return err
	}

	// Resolve each VM's network and store expressions
	for index := range cfg.VMs {
		vm := &cfg.VMs[index]

		if err := resolveExpr(
			vm.NetExpr,
			evalCtx,
			&vm.NetName,
			"vm %q: network",
			vm.Name,
		); err != nil {
			return err
		}
		if err := resolveExpr(
			vm.StoreExpr,
			evalCtx,
			&vm.Store,
			"vm %q: store",
			vm.Name,
		); err != nil {
			return err
		}
	}

	return nil
}

// collectNames extracts names from a slice, validates they're non-empty
// and unique, and returns them as a set.
func collectNames[T any](
	kind string,
	items []T,
	getName func(T) string,
) (map[string]struct{}, error) {
	names := make(map[string]struct{}, len(items))
	for _, item := range items {
		name := getName(item)
		if name == "" {
			return nil, fmt.Errorf("%s with empty name", kind)
		}
		if _, exists := names[name]; exists {
			return nil, fmt.Errorf("duplicate %s %q", kind, name)
		}
		names[name] = struct{}{}
	}
	return names, nil
}

// resolveExpr evaluates an HCL expression to a string and stores
// the result in target.
func resolveExpr(
	expr hcl.Expression,
	evalCtx *hcl.EvalContext,
	target *string,
	errFmt string,
	errArgs ...any,
) error {
	val, diags := expr.Value(evalCtx)
	if diags.HasErrors() {
		return fmt.Errorf(errFmt+": %w", append(errArgs, diags)...)
	}
	if val.Type() != cty.String {
		return fmt.Errorf(
			errFmt+": expected string, got %s",
			append(errArgs, val.Type().FriendlyName())...)
	}
	*target = val.AsString()
	return nil
}

// buildEvalContext creates the HCL symbol table.
//
// It registers these namespaces so HCL expressions can reference them:
//   - local.X        → value from the locals block
//   - network.X      → name of a network defined in this file
//   - store.X        → name of a store defined in this file
//   - data.store.X   → name of a store that exists in the DB
//   - data.network.X → name of a network that exists in the DB
func buildEvalContext(
	cfg *hclConfig,
	networks map[string]struct{},
	stores map[string]struct{},
	ctx context.Context,
	dbHandler *database.DBHandler,
) (*hcl.EvalContext, error) {
	evalCtx := &hcl.EvalContext{
		Variables: map[string]cty.Value{},
	}

	// Locals: evaluate each expression (locals can't reference resources)
	if cfg.Locals != nil {
		localVals := map[string]cty.Value{}
		for key, expr := range cfg.Locals.Values {
			val, diags := expr.Value(nil)
			if diags.HasErrors() {
				return nil, fmt.Errorf("local.%s: %w", key, diags)
			}
			localVals[key] = val
		}
		evalCtx.Variables["local"] = cty.ObjectVal(localVals)
	}

	// Networks defined in this file
	netMap := map[string]cty.Value{}
	for name := range networks {
		netMap[name] = cty.StringVal(name)
	}

	// Stores defined in this file
	storeMap := map[string]cty.Value{}
	for name := range stores {
		storeMap[name] = cty.StringVal(name)
	}

	// Data sources: references to resources already in the DB
	dataNet := map[string]cty.Value{}
	dataStore := map[string]cty.Value{}
	for _, data := range cfg.Data {
		switch data.Type {
		case "network", "store":
			obj, err := dbHandler.Get(ctx, data.Type, data.Name, "default")
			if err != nil {
				return nil, fmt.Errorf("data.%s.%s: %w", data.Type, data.Name, err)
			}
			if obj == nil {
				return nil, fmt.Errorf("data.%s.%s: resource not found", data.Type, data.Name)
			}
			if data.Type == "network" {
				dataNet[data.Name] = cty.StringVal(data.Name)
			} else {
				dataStore[data.Name] = cty.StringVal(data.Name)
			}
		default:
			return nil, fmt.Errorf("unknown data type %q (supported: store, network)", data.Type)
		}
	}

	// Register resource namespaces
	if len(netMap) > 0 {
		evalCtx.Variables["network"] = cty.ObjectVal(netMap)
	}
	if len(storeMap) > 0 {
		evalCtx.Variables["store"] = cty.ObjectVal(storeMap)
	}

	dataVars := map[string]cty.Value{}
	if len(dataNet) > 0 {
		dataVars["network"] = cty.ObjectVal(dataNet)
	}
	if len(dataStore) > 0 {
		dataVars["store"] = cty.ObjectVal(dataStore)
	}
	if len(dataVars) > 0 {
		evalCtx.Variables["data"] = cty.ObjectVal(dataVars)
	}

	return evalCtx, nil
}
