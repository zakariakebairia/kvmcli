package internal

import (
	"fmt"
	"log"
	"net/url"

	"github.com/digitalocean/go-libvirt"
)

// ConnectLibvirt create and returns a libvirt connection
func ConnectLibvirt() (*libvirt.Libvirt, error) {
	uri, err := url.Parse(string(libvirt.QEMUSystem))
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	l, err := libvirt.ConnectToURI(uri)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	return l, nil
}
