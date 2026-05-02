package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

type StoreAttrs struct {
	Backend       string   `json:"backend"`
	ArtifactsPath string   `json:"artifacts_path"`
	Images        []*Image `json:"images"`
}
type Image struct {
	Checksum  string `json:"checksum"`
	Display   string `json:"display"`
	File      string `json:"file"`
	Name      string `json:"name"`
	OsProfile string `json:"os_profile"`
}

func (s *StoreAttrs) FromObject(object *registry.Object) error {
	data, err := json.Marshal(object.Attrs)
	if err != nil {
		return fmt.Errorf("failed to marshal attrs: %w", err)
	}
	if err := json.Unmarshal(data, s); err != nil {
		return fmt.Errorf("failed to unmarshal attrs: %w", err)
	}
	return nil
}

func (s *StoreAttrs) Validate() error {
	var errs []string
	if s.Backend == "" {
		errs = append(errs, "store.backend is required")
	}
	if s.ArtifactsPath == "" {
		errs = append(errs, "store.artifacts path is required")
	}
	if s.Images == nil {
		errs = append(errs, "store.images is required")
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("store: %s", strings.Join(errs, ", "))
}
