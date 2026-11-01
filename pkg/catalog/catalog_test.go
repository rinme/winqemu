package catalog

import (
	"testing"
)

func TestLoadEmbeddedCatalog(t *testing.T) {
	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("failed to load embedded catalog: %v", err)
	}

	if len(cat.Families) != 4 {
		t.Errorf("expected 4 families (win11, win10, win81, win7), got %d", len(cat.Families))
	}

	// Verify Windows 11 family has WispyDocs recommended editions
	win11 := cat.GetFamily("win11")
	if win11 == nil {
		t.Fatalf("win11 family not found")
	}

	var foundIoTRec bool
	for _, ed := range win11.Editions {
		if ed.ID == "win11-iot-enterprise-ltsc-2024" && ed.IsRecommended {
			foundIoTRec = true
		}
	}

	if !foundIoTRec {
		t.Errorf("expected win11-iot-enterprise-ltsc-2024 to be marked as recommended")
	}
}

func TestGetEdition(t *testing.T) {
	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("failed to load catalog: %v", err)
	}

	ed := cat.GetEdition("win10-iot-enterprise-ltsc-2021")
	if ed == nil {
		t.Fatalf("expected edition win10-iot-enterprise-ltsc-2021 to be found")
	}

	if ed.Version != "21H2" {
		t.Errorf("expected version 21H2, got %s", ed.Version)
	}

	if !ed.IsRecommended {
		t.Errorf("expected Win10 IoT LTSC to be recommended by WispyDocs")
	}
}
