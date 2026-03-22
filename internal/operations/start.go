package operations

import "github.com/zakariakebairia/kvmcli/internal/resources"

func (o *Operator) Start(r resources.Resource) error {
	return r.Start()
}
