package registry

import (
	"context"
	"database/sql"

	"github.com/digitalocean/go-libvirt"
)

// Session carries the shared dependencies for one command invocation.
type Session struct {
	Ctx  context.Context
	DB   *sql.DB
	Conn *libvirt.Libvirt
}

type Action int

const (
	ActionNone Action = iota
	ActionCreate
	ActionUpdate
	ActionDelete
)

// State is the unified struct for any resource (vm, network, store ...etc)
// IDEA: I will change its name  to Object or something similar

// type Object stcut {}
type Object struct {
	TypeName  string
	Name      string
	Namespace string
	Labels    map[string]string
	Attrs     map[string]any
	Status    string
}

// Change is per resource, each resource have a current state and a desirred state.
// and an Action defined upon the difference between the desired and the current state.
type Change struct {
	Action  Action
	Desired *Object // nil for Delete, which means I want something to become nil
	Current *Object // nil for Create, because we don't have any current Object
}

// Plan is a list of changes, because I have multiple resource per manifest
// this plan group them all
type Plan struct {
	Changes []Change
}
