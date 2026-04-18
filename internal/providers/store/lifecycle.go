package store

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

func init() {
	registry.Register(&registry.ResourceType{
		Name:      "store",
		DependsOn: []string{}, // stores have no dependencies
		Lifecycle: &StoreLifecycle{},
		Columns:   []string{"NAME", "NAMESPACE", "BACKEND", "ARTIFACTS", "IMAGES", "STATUS"},
		Format: func(s registry.Object) []string {
			return []string{
				s.Name,
				s.Namespace,
				s.GetString("backend"),
				s.GetString("artifacts_path"),
				s.GetString("images_path"),
				s.Status,
			}
		},
	})
}

// I might change this into a Pool and Volumes
// StoreLifecycle implements registry.ResourceLifecycle.
type StoreLifecycle struct{}

func (l *StoreLifecycle) Plan(desired, current *registry.Object) (registry.Action, error) {
	if current == nil && desired != nil {
		return registry.ActionCreate, nil
	}
	if current != nil && desired == nil {
		return registry.ActionDelete, nil
	}
	// Could add update detection here later
	return registry.ActionNone, nil
}

func (l *StoreLifecycle) Apply(session registry.Session, change registry.Change) error {
	spec := change.Desired

	// This is the ONLY thing store-specific: insert images into images table
	images, ok := spec.Attrs["images"].([]map[string]any)
	if ok {
		if err := insertImages(session, spec.Name, spec.Namespace, images); err != nil {
			return err
		}
	}
	return nil
}

func (l *StoreLifecycle) Destroy(session registry.Session, change registry.Change) error {
	// Images are cleaned up by ON DELETE CASCADE in the images table FK.
	// The engine handles removing the state from the resources table.
	return nil
}

func insertImages(
	session registry.Session,
	storeName string,
	namespace string,
	images []map[string]any,
) error {
	// Get the repository's ID from the resources table
	// (or from the images table's foreign key — we'll figure out the exact approach
	// when we wire up the state store)

	// const imgInsert = `
	//          INSERT INTO images (
	//              store_id, name, display, version, os_profile,
	//              file, checksum, size
	//          ) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	//      `
	if err := ensureImagesTable(session.Ctx, session.DB); err != nil {
		return err
	}

	const imgInsert = `
          INSERT INTO images (
              store_name, store_ns, name, display, version, os_profile,
              file, checksum, size
          ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
      `

	tx, err := session.DB.BeginTx(session.Ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, img := range images {
		_, err := tx.ExecContext(session.Ctx, imgInsert,
			// /* store_id */ ...,
			storeName, namespace,
			img["name"], img["display"], img["version"],
			img["os_profile"], img["file"], img["checksum"], img["size"],
		)
		if err != nil {
			return fmt.Errorf("insert image %v: %w", img["name"], err)
		}
	}

	return tx.Commit()
}
