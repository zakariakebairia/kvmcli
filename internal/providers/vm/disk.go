package vm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/zakariakebairia/kvmcli/internal/registry"
)

var QemuImgBinary = "qemu-img"

func createOverlay(ctx context.Context, src, dest string) error {
	args := []string{
		"create",
		"-f", "qcow2",
		"-o", fmt.Sprintf("backing_file=%s,backing_fmt=qcow2", src),
		dest,
	}
	output, err := exec.CommandContext(ctx, QemuImgBinary, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("create overlay failed: %w, %s ", err, output)
	}
	return nil
}

func resizeOverlay(ctx context.Context, dest string) error {
	// args := []string{}
	return nil
}

func deleteOverlay(dest string) error {
	// if file exist but remove process returns error, return that error
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete overlay %q: %w", dest, err)
	}
	// otherwise, return nil
	return nil
}

func provisionDisk(session registry.Session, spec *registry.Object) (string, error) {
	image, err := getImage(session, spec.GetString("store"), spec.GetString("image"))
	if err != nil {
		return "", fmt.Errorf("lookup image: %w", err)
	}

	src := filepath.Join(image.ArtifactsPath, image.ImageFile)
	diskPath := filepath.Join(image.ImagesPath, spec.Name+".qcow2")

	if err = createOverlay(session.Ctx, src, diskPath); err != nil {
		return "", fmt.Errorf("create disk overlay: %w", err)
	}

	return diskPath, nil
}
