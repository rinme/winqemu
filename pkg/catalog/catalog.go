package catalog

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

//go:embed catalog.json
var embeddedCatalogJSON []byte

// LoadCatalog unmarshals the embedded catalog dataset.
func LoadCatalog() (*Catalog, error) {
	var cat Catalog
	if err := json.Unmarshal(embeddedCatalogJSON, &cat); err != nil {
		return nil, fmt.Errorf("failed to parse embedded catalog: %w", err)
	}
	return &cat, nil
}

// FetchRemoteCatalog attempts to retrieve the latest catalog from a remote mirror URL.
func FetchRemoteCatalog(mirrorURL string) (*Catalog, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(mirrorURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mirror returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var cat Catalog
	if err := json.Unmarshal(body, &cat); err != nil {
		return nil, err
	}

	return &cat, nil
}

// GetFamily finds an OSFamily by its identifier ("win11", "win10", "win81", "win7").
func (c *Catalog) GetFamily(id string) *OSFamily {
	for i := range c.Families {
		if c.Families[i].ID == id {
			return &c.Families[i]
		}
	}
	return nil
}

// GetEdition searches all families for a specific edition by ID.
func (c *Catalog) GetEdition(id string) *Edition {
	for _, fam := range c.Families {
		for _, ed := range fam.Editions {
			if ed.ID == id {
				return &ed
			}
		}
	}
	return nil
}

// RecommendedEditions returns all editions marked as recommended across all families.
func (c *Catalog) RecommendedEditions() []Edition {
	var list []Edition
	for _, fam := range c.Families {
		for _, ed := range fam.Editions {
			if ed.IsRecommended {
				list = append(list, ed)
			}
		}
	}
	return list
}
