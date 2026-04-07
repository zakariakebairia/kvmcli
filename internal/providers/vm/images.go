package vm

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

type Image struct {
	ArtifactsPath string
	ImagesPath    string
	ImageFile     string
	OsProfile     string
}

// IDEA: I will add store to the arguments of the function
// func getImage(session registry.Session, store, imageName, string) (*Image, error)

func getImage(session registry.Session, storeName, imageName string) (*Image, error) {
	dbHandler := database.NewDBHandler(session.DB)
	store, err := dbHandler.Get(session.Ctx, "store", storeName, "homelab")
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}

	images := store.Attrs["images"].([]any)
	for _, raw := range images {

		img, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if img["name"] == imageName {
			return &Image{
				ArtifactsPath: store.Attrs["artifacts_path"].(string),
				ImagesPath:    store.Attrs["images_path"].(string),
				ImageFile:     img["file"].(string),
				OsProfile:     img["os_profile"].(string),
			}, nil
		}
	}

	return nil, fmt.Errorf("image %q not found in any store", imageName)
}
