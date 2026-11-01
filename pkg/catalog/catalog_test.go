package catalog

import (
	"strings"
	"testing"
)

func TestLoadEmbeddedCatalog(t *testing.T) {
	cat, err := LoadCatalog()
	if err != nil {
		t.Fatalf("failed to load embedded catalog: %v", err)
	}

	if len(cat.Families) < 4 {
		t.Errorf("expected at least 4 families, got %d", len(cat.Families))
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

	if !strings.HasPrefix(ed.Version, "21H2") {
		t.Errorf("expected version starting with 21H2, got %s", ed.Version)
	}

	if !ed.IsRecommended {
		t.Errorf("expected Win10 IoT LTSC to be recommended by WispyDocs")
	}
}
