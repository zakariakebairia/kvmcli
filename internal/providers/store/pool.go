package store

import (
	"encoding/xml"
	"fmt"

	"github.com/digitalocean/go-libvirt"
)

// Pool struct
type PoolType string

const (
	PoolTypeDir PoolType = "dir"
)

type Pool struct {
	XMLName xml.Name   `xml:"pool"`
	Type    PoolType   `xml:"type,attr"`
	Name    string     `xml:"name"`
	Source  PoolSource `xml:"source"`
	Target  PoolTarget `xml:"target"`
}
type PoolTarget struct {
	Path string `xml:"path"`
}
type PoolSource struct {
	Dir string `xml:"dir,omitempty"`
}
type PoolSize struct {
	Value uint64 `xml:",chardata"`
	Unit  string `xml:"unit,attr,omitempty"`
}

// GenerateXML to generate the xml content for a Pool
// ps: without an xml header
func (p *Pool) GenerateXML() (string, error) {
	// return xml.MarshalIndent(p, "", "  ")
	XML, err := xml.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marsheling xml for %q: %w", p.Name, err)
	}
	return xml.Header + string(XML), nil
}

// A constructor for a Pool
func NewPool(tp PoolType, name, targetPath, dir string) Pool {
	return Pool{
		Type: tp,
		Name: name,
		Target: PoolTarget{
			Path: targetPath,
		},
		Source: PoolSource{
			Dir: dir,
		},
	}
}

func (p *Pool) Create(conn *libvirt.Libvirt, XML string) error {
	fmt.Println("-->   Hello from create pool")
	storagePool, err := conn.StoragePoolDefineXML(XML, 0)
	if err != nil {
		return fmt.Errorf("define pool %q: %w", p.Name, err)
	}
	if err := conn.StoragePoolBuild(storagePool, 0); err != nil {
		return fmt.Errorf("build pool %q: %w", p.Name, err)
	}
	if err := conn.StoragePoolCreate(storagePool, 0); err != nil {
		return fmt.Errorf("create pool %q: %w", p.Name, err)
	}
	return nil
}

func (p *Pool) Destroy(conn *libvirt.Libvirt, name string) error {
	storagePool, err := conn.StoragePoolLookupByName(name)
	if err != nil {
		return fmt.Errorf("lookup pool %q: %w", p.Name, err)
	}
	if err := conn.StoragePoolDestroy(storagePool); err != nil {
		return fmt.Errorf("destroy pool %q: %w", p.Name, err)
	}
	if err := conn.StoragePoolUndefine(storagePool); err != nil {
		return fmt.Errorf("undefine pool %q: %w", p.Name, err)
	}
	return nil
}
