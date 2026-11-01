package unattend

import (
	"strings"
	"testing"
	"winqemu/pkg/catalog"
	"winqemu/pkg/tuning"
)

func TestGenerateAutounattendXML(t *testing.T) {
	ed := catalog.Edition{
		ID:         "win11-iot-enterprise-ltsc-2024",
		Name:       "Windows 11 IoT Enterprise LTSC 2024",
		Version:    "24H2",
		ProductKey: "KBN8V-HFGQ4-RRX47-PRTH7-36CJ3",
		IsWin11:    true,
	}

	opts := tuning.OptimizationConfig{
		BypassTPMAndOOBE:    true,
		MASActivationPrompt: true,
	}

	xml, err := GenerateAutounattendXML(ed, opts, "AdminUser", "")
	if err != nil {
		t.Fatalf("failed to generate autounattend XML: %v", err)
	}

	// Verify TPM and Secure Boot bypass registry keys are present
	if !strings.Contains(xml, "BypassTPMCheck") {
		t.Errorf("expected xml to contain BypassTPMCheck")
	}
	if !strings.Contains(xml, "BypassSecureBootCheck") {
		t.Errorf("expected xml to contain BypassSecureBootCheck")
	}

	// Verify local account setup
	if !strings.Contains(xml, "AdminUser") {
		t.Errorf("expected user AdminUser to be in xml")
	}

	// Verify VirtIO driver paths
	if !strings.Contains(xml, "PnpCustomizationsWinPE") {
		t.Errorf("expected PnpCustomizationsWinPE for virtio drivers")
	}
}

func TestCreateFATDiskImage(t *testing.T) {
	testXML := "<unattend>test</unattend>"
	diskBytes, err := BuildFAT12Image("autounattend.xml", []byte(testXML))
	if err != nil {
		t.Fatalf("failed to build FAT12 image: %v", err)
	}

	// 1.44MB standard floppy is exactly 1,474,560 bytes
	if len(diskBytes) != 1474560 {
		t.Errorf("expected 1474560 bytes, got %d", len(diskBytes))
	}

	// Check FAT12 boot signature 0x55AA at offset 510
	if diskBytes[510] != 0x55 || diskBytes[511] != 0xAA {
		t.Errorf("invalid boot sector signature")
	}
}
