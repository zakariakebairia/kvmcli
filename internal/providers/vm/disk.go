package vm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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
		return fmt.Errorf("create overlay failed: %s: %w", output, err)
	}
	return nil
}

func resizeOverlay(ctx context.Context, dest string) error {
	// args := []string{}
	return nil
}

func deleteOverlay(ctx context.Context, dest string) error {
	// if file exist but remove process returns error, return that error
	if err := os.Remove(dest); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete overlay %q: %w", dest, err)
	}
	// otherwise, return nil
	return nil
}
