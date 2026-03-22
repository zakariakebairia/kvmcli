package vms

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	log "github.com/zakariakebairia/kvmcli/internal/logger"
)

type DiskManager interface {
	CreateOverlay(ctx context.Context, src, dest string) error
	DeleteOverlay(ctx context.Context, dest string) error
	Paths() (baseImagesPath, destImagesPath string)
	// Size()
}

type QemuDiskManager struct {
	QemuImgPath    string
	Timeout        time.Duration
	BaseImagesPath string
	DestImagesPath string
	// Disk           []DiskManager
}

func (d *QemuDiskManager) Paths() (string, string) {
	return d.BaseImagesPath, d.DestImagesPath
}

func NewQemuDiskManager() (*QemuDiskManager, error) {
	return &QemuDiskManager{}, nil
}

func (d *QemuDiskManager) CreateOverlay(ctx context.Context, src, dest string) error {
	// Build a context with timeout
	timeout := d.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	args := []string{
		"create",
		"-f", "qcow2",
		"-o", fmt.Sprintf("backing_file=%s,backing_fmt=qcow2", src),
		dest,
	}
	cmdPath := d.QemuImgPath
	if cmdPath == "" {
		cmdPath = "qemu-img"
	}
	// Execute
	output, err := exec.CommandContext(ctx, cmdPath, args...).CombinedOutput()
	if err != nil {
		log.Errorf("qemu-img error: %s", output)
		return fmt.Errorf("create overlay failed: %w", err)
	}
	log.Debugf("overlay created at %s", dest)
	return nil
}

func (d *QemuDiskManager) DeleteOverlay(ctx context.Context, dest string) error {
	log.Debugf("deleting overlay at %s", dest)
	if err := os.Remove(dest); err != nil {
		return fmt.Errorf("delete overlay %q failed: %w", dest, err)
	}
	return nil
}

func (d *QemuDiskManager) GetPath() error {
	return nil
}
