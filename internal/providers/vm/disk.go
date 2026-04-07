package vm

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// TODO: qemuImgPath can be a variable specified at build time using makefile
var QemuImgBinary = "qemu-img"

// diskManager wraps qemu-img operations.
type diskManager struct {
	qemuImgPath string
	timeout     time.Duration
}

func newDiskManager() *diskManager {
	return &diskManager{
		qemuImgPath: QemuImgBinary,
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
