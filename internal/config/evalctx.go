package config

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/kebairia/kvmcli/internal/database"
	"github.com/zclconf/go-cty/cty"
)

func BuildEvalContext(
	cfg *Config,
	idx *Index,
	ctx context.Context,
	db *sql.DB,
) (*hcl.EvalContext, error) {
	vars := map[string]cty.Value{}
	locals := map[string]cty.Value{}

	// ------------------------------------------------------------
	// Variables (defaults only for now)
	// ------------------------------------------------------------
	for _, v := range cfg.Variables {
		if v.Default == nil {
			continue
		}
		val, diags := v.Default.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("var.%s: %s", v.Name, diags.Error())
		}
		vars[v.Name] = val
	}

	// ------------------------------------------------------------
	// Locals
	// ------------------------------------------------------------
	if cfg.Locals != nil {
		for name, expr := range cfg.Locals.Values {
			val, diags := expr.Value(&hcl.EvalContext{
				Variables: map[string]cty.Value{
					"var": cty.ObjectVal(vars),
				},
			})
			if diags.HasErrors() {
				return nil, fmt.Errorf("local.%s: %s", name, diags.Error())
			}
			locals[name] = val
		}
	}

	// ------------------------------------------------------------
	// network / store objects
	// ------------------------------------------------------------
	netMap := map[string]cty.Value{}
	for n := range idx.Networks {
		netMap[n] = cty.StringVal(n) // Changed from ObjectVal to StringVal to match parser.go logic which seemed simpler?
		// Wait, parser.go used cty.StringVal(n).
		// evalctx.go.old used cty.ObjectVal(map[string]cty.Value{"name": cty.StringVal(n)})
		// If I change this, I might break existing HCL if it relies on network.name.
		// parser.go says:
		// netMap[n] = cty.StringVal(n)
		// ...
		// vm.NetExpr.Value(evalCtx) -> Expects String.
		// So `network.default` should resolve to "default" (string).
		// If I use ObjectVal, `network.default` resolves to an object. Then the user has to write `network.default.name`?
		// parser.go `evalString` expects String.
		// I stick to parser.go implementation: StringVal.
	}

	storeMap := map[string]cty.Value{}
	for s := range idx.Stores {
		storeMap[s] = cty.StringVal(s)
	}

	// ------------------------------------------------------------
	// data sources (validated)
	// ------------------------------------------------------------
	dataNet := map[string]cty.Value{}
	dataStore := map[string]cty.Value{}

	for _, d := range cfg.Data {
		switch d.Type {
		case "network":
			if _, err := database.GetNetworkIDByName(ctx, db, d.Name); err != nil {
				return nil, fmt.Errorf("data.network.%s: %w", d.Name, err)
			}
			dataNet[d.Name] = cty.StringVal(d.Name)
		case "store":
			if _, err := database.GetStoreIDByName(ctx, db, d.Name); err != nil {
				return nil, fmt.Errorf("data.store.%s: %w", d.Name, err)
			}
			dataStore[d.Name] = cty.StringVal(d.Name)
		default:
			return nil, fmt.Errorf("unknown data type %q (supported: store, network)", d.Type)
		}
	}

	// Construct the data object
	dataObj := cty.ObjectVal(map[string]cty.Value{
		"network": cty.ObjectVal(dataNet),
		"store":   cty.ObjectVal(dataStore),
	})

	return &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"var":     cty.ObjectVal(vars),
			"local":   cty.ObjectVal(locals),
			"network": cty.ObjectVal(netMap),
			"store":   cty.ObjectVal(storeMap),
			"data":    dataObj,
		},
	}, nil
}
