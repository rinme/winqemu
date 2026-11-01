package vm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"winqemu/pkg/catalog"
	"winqemu/pkg/tuning"
)

func TestBuildQEMUArgs(t *testing.T) {
	cfg := &VMConfig{
		Name:          "TestWin11",
		RAMMB:         8192,
		CPUCores:      4,
		CPUThreads:    2,
		DiskPath:      "/tmp/test.qcow2",
		ISOPath:       "/tmp/win11.iso",
		VirtIOISOPath: "/tmp/virtio.iso",
		UnattendDisk:  "/tmp/unattend.vfd",
		Tuning:        tuning.GetPreset(tuning.ProfileGaming, true),
	}

	args := BuildQEMUArgs(cfg)
	argStr := strings.Join(args, " ")

	if !strings.Contains(argStr, "-m 8192M") {
		t.Errorf("expected -m 8192M, got: %s", argStr)
	}

	if !strings.Contains(argStr, "hv_relaxed") {
		t.Errorf("expected hv_relaxed in args, got: %s", argStr)
	}

	if !strings.Contains(argStr, "-fda /tmp/unattend.vfd") {
		t.Errorf("expected unattend floppy, got: %s", argStr)
	}

	if !strings.Contains(argStr, "virtio-scsi-pci") {
		t.Errorf("expected virtio-scsi-pci in args, got: %s", argStr)
	}
}

func TestGenerateLibvirtXML(t *testing.T) {
	cfg := &VMConfig{
		Name:            "Win11-Gaming",
		RAMMB:           16384,
		CPUCores:        8,
		CPUThreads:      2,
		DiskPath:        "/var/lib/libvirt/images/win11.qcow2",
		ISOPath:         "/var/lib/libvirt/images/win11.iso",
		VirtIOISOPath:   "/var/lib/libvirt/images/virtio.iso",
		UnattendDisk:    "/var/lib/libvirt/images/unattend.vfd",
		SelectedEdition: catalog.Edition{IsWin11: true},
		Tuning:          tuning.GetPreset(tuning.ProfileGaming, true),
	}

	xml, err := GenerateLibvirtXML(cfg)
	if err != nil {
		t.Fatalf("failed to generate Libvirt XML: %v", err)
	}

	if !strings.Contains(xml, "<name>Win11-Gaming</name>") {
		t.Errorf("expected domain name in XML, got: %s", xml)
	}

	if !strings.Contains(xml, "<memory unit='MiB'>16384</memory>") {
		t.Errorf("expected memory in XML, got: %s", xml)
	}

	if !strings.Contains(xml, "model='virtio-scsi'") {
		t.Errorf("expected virtio-scsi controller in XML, got: %s", xml)
	}

	if !strings.Contains(xml, "<tpm model='tpm-crb'>") {
		t.Errorf("expected tpm-crb in XML for Win11, got: %s", xml)
	}
}

func TestGenerateLaunchScript(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "winqemu-script-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &VMConfig{
		Name:       "TestVM",
		RAMMB:      4096,
		CPUCores:   2,
		CPUThreads: 1,
		DiskPath:   filepath.Join(tmpDir, "disk.qcow2"),
		OutputDir:  tmpDir,
		Tuning:     tuning.GetPreset(tuning.ProfileBalanced, false),
	}

	scriptPath, err := GenerateLaunchScript(cfg)
	if err != nil {
		t.Fatalf("failed to generate launch script: %v", err)
	}

	if _, err := os.Stat(scriptPath); err != nil {
		t.Fatalf("expected script to exist at %s", scriptPath)
	}
}
