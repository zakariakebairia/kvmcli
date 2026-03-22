package operations

import "github.com/zakariakebairia/kvmcli/internal/resources"

func (o *Operator) Insert(r resources.Record) error {
	return r.Insert(o.ctx, o.db)
}
