package vm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

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

// diskManager wraps qemu-img operations.
type diskManager struct {
	qemuImgPath string
	timeout     time.Duration
}

func newDiskManager() *diskManager {
	return &diskManager{
		qemuImgPath: "qemu-img",
		timeout:     30 * time.Second,
	}
}

func (d *diskManager) CreateOverlay(ctx context.Context, src, dest string) error {
	args := []string{
		"create",
		"-f", "qcow2",
		"-o", fmt.Sprintf("backing_file=%s,backing_fmt=qcow2", src),
		dest,
	}
	output, err := exec.CommandContext(ctx, d.qemuImgPath, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("create overlay failed: %s: %w", output, err)
	}
	return nil
}

func (d *diskManager) DeleteOverlay(ctx context.Context, dest string) error {
	if err := os.Remove(dest); err != nil {
		return fmt.Errorf("delete overlay %q: %w", dest, err)
	}
	return nil
}
