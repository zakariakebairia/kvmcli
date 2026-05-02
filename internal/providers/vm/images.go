package vm

import (
	"fmt"
	"path/filepath"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

//	type Image struct {
//		ArtifactsPath   string
//		ImagesPath      string
//		ImageFile       string
//		OsProfile       string
//		SourceImageFile string
//		DiskPath        string
//	}
type Image struct {
	File      string
	OsProfile string
	SrcPath   string // absolute path to the base image artifact
	DiskPath  string // directory where overlay disks should be placed
}

// FIX: What if the store and other object are in different namespaces
func getImage(session registry.Session, storeName, imageName, nameSpace string) (*Image, error) {
	dbHandler := database.NewDBHandler(session.DB)
	store, err := dbHandler.Get(session.Ctx, "store", storeName, nameSpace)
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
		return &Image{
			File:      file,
			OsProfile: osProfile,
			SrcPath:   filepath.Join(artifactsPath, file),
			DiskPath:  filepath.Join(imagesPath, imageName+".qcow2"),
		}, nil
	}
	return nil, nil
}
