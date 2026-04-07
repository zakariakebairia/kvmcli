package vm

import (
	"fmt"

	"github.com/zakariakebairia/kvmcli/internal/database"
	"github.com/zakariakebairia/kvmcli/internal/registry"
)

func attrStr(s registry.Object, key string) string {
	if value, ok := s.Attrs[key].(string); ok {
		return value
	}
	return ""
}

func attrInt(s registry.Object, key string) int {
	// JSON numbers from the DB come back as float64
	switch value := s.Attrs[key].(type) {
	case int:
		return value
	case float64:
		return int(value)
	}
	return 0
}

// lookupImage finds an image by name across all stores in the DB.
// Returns the artifacts path, images path, and image file name.
func lookupImage(
	session registry.Session,
	imageName string,
) (artifactsPath, imagesPath, imageFile, osProfile string, err error) {
	dbHandler := database.NewDBHandler(session.DB)

	stores, err := dbHandler.List(session.Ctx, "store")
	if err != nil {
		return "", "", "", "", fmt.Errorf("list stores: %w", err)
	}

	for _, store := range stores {
		images, ok := store.Attrs["images"].([]any)
		if !ok {
			continue
		}
		for _, raw := range images {
			img, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if img["name"] == imageName {
				artifacts, _ := store.Attrs["artifacts_path"].(string)
				imagesDir, _ := store.Attrs["images_path"].(string)
				file, _ := img["file"].(string)
				profile, _ := img["os_profile"].(string)
				return artifacts, imagesDir, file, profile, nil
			}
		}
	}

	return "", "", "", "", fmt.Errorf("image %q not found in any store", imageName)
}
