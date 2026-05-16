package vm

import (
	"fmt"
	"path/filepath"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

type Image struct {
	File      string
	Display   string
	OsProfile string
	SrcPath   string // absolute path to the base image artifact
	DiskPath  string // directory where overlay disks should be placed
}

// FIX: What if the store and other object are in different namespaces
func getImage(
	session registry.Session,
	object *registry.Object,
) (*Image, error) {
	// Get the store name
	storeName := object.GetString("store")
	// Get the image name
	imageName := object.GetString("image")
	dbHandler := database.NewDBHandler(session.DB)
	store, err := dbHandler.Get(session.Ctx, "store", storeName, object.Namespace)
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}
	artifactsPath, ok := store.Attrs["artifacts_path"].(string)
	if !ok {
		return nil, fmt.Errorf("store %q: missing or invalid artifacts_path", storeName)
	}
	imagesPath, ok := store.Attrs["images_path"].(string)
	if !ok {
		return nil, fmt.Errorf("store %q: missing or invalid images_path", storeName)
	}

	images, ok := store.Attrs["images"].([]any)
	if !ok {
		return nil, fmt.Errorf("store %q: missing or invalid images list", storeName)
	}
	for _, raw := range images {
		image, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if image["name"] != imageName {
			continue
		}
		file, ok := image["file"].(string)
		if !ok {
			return nil, fmt.Errorf("image %q: missing or invalid file field", imageName)
		}
		osProfile, ok := image["os_profile"].(string)
		if !ok {
			return nil, fmt.Errorf("image %q: missing or invalid os_profile field", imageName)
		}
		imageDisplay, ok := image["display"].(string)
		if !ok {
			return nil, fmt.Errorf("image %q: missing or invalid display", storeName)
		}
		return &Image{
			File:      file,
			OsProfile: osProfile,
			SrcPath:   filepath.Join(artifactsPath, file),
			DiskPath:  filepath.Join(imagesPath, object.Name+".qcow2"),
			Display:   imageDisplay,
		}, nil
	}
	return nil, fmt.Errorf("image %q not found in store %q", imageName, storeName)
}
