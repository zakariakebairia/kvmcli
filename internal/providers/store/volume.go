package store

import (
	"encoding/xml"
	"fmt"

	"github.com/digitalocean/go-libvirt"
)

type VolumeFormatType string

const (
	QCOW2Format VolumeFormatType = "qcow2"
	RAWFormat   VolumeFormatType = "raw"
)

//
// type Volume struct {
// 	XMLName      xml.Name           `xml:"volume"`
// 	Type         string             `xml:"type,attr"`
// 	Name         string             `xml:"name"`
// 	Capacity     Capacity           `xml:"capacity"`
// 	VolumeTarget VolumeTarget       `xml:"target"`
// 	BackingStore VolumeBackingStore `xml:"backingStore"`
// }
// type Capacity struct {
// 	Value uint64 `xml:",chardata"`
// 	Unit  string `xml:"unit,attr,omitempty"`
// }
// type VolumeTarget struct {
// 	Format VolumeFormat `xml:"format,attr"`
// }
// type VolumeBackingStore struct {
// 	Path   string       `xml:"path"`
// 	Format VolumeFormat `xml:"format"`
// }
//
// type VolumeFormat struct {
// 	Type string `xml:"type,attr"`
// }

type Volume struct {
	XMLName      xml.Name         `xml:"volume"`
	Type         VolumeFormatType `xml:"type,attr"`
	Name         string           `xml:"name"`
	Capacity     VolumeCapacity   `xml:"capacity"`
	Target       VolumeTarget     `xml:"target"` // this is where the actual disk for the vm
	BackingStore *BackingStore    `xml:"backingStore,omitempty"`
}

type VolumeCapacity struct {
	Unit  string `xml:"unit,attr,omitempty"`
	Value uint64 `xml:",chardata"`
}

type VolumeTarget struct {
	Format VolumeFormat `xml:"format"`
}

type VolumeFormat struct {
	Type string `xml:"type,attr"`
}

type BackingStore struct {
	Path   string       `xml:"path"`
	Format VolumeFormat `xml:"format"`
}

// GenerateXML to generate the xml content for a Volume
func (v *Volume) GenerateXML() (string, error) {
	// return xml.MarshalIndent(p, "", "  ")
	XML, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marsheling xml for %q: %w", v.Name, err)
	}
	return xml.Header + string(XML), nil
}

func NewVolume(tp VolumeFormatType, name string, capacity uint64, path string) Volume {
	return Volume{
		Type: tp,
		Name: name,
		Capacity: VolumeCapacity{
			Unit:  "bytes",
			Value: capacity,
		},
		Target: VolumeTarget{
			Format: VolumeFormat{Type: "qcow2"},
		},
		BackingStore: &BackingStore{
			Path:   path,
			Format: VolumeFormat{Type: "qcow2"},
		},
	}
}

func (v *Volume) Create(conn *libvirt.Libvirt, pool libvirt.StoragePool, XML string) error {
	// pool, err := conn.StoragePoolLookupByName("mypool")
	// if err != nil {
	// 	return fmt.Errorf("pool lookup: %w", err)
	// }

	volume, err := conn.StorageVolCreateXML(pool, XML, 0)
	if err != nil {
		return fmt.Errorf("volume xml creation: %w", err)
	}
	_ = volume
	return nil
}
func (v *Volume) Destroy(name string) error { return nil }
