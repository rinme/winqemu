package catalog

// Edition represents a specific Windows edition.
type Edition struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Version         string   `json:"version"`          // e.g. "24H2", "23H2", "21H2", "SP1"
	Release         string   `json:"release"`          // e.g. "2024", "2021"
	Architecture    string   `json:"arch"`             // "x64", "x86", "arm64"
	Language        string   `json:"lang"`             // "en-us"
	DownloadURL     string   `json:"download_url"`
	SHA256          string   `json:"sha256"`
	SizeBytes       int64    `json:"size_bytes"`
	IsRecommended   bool     `json:"is_recommended"`   // WispyDocs recommended badge
	Recommendation  string   `json:"recommendation"`   // Reasoning note from WispyDocs
	ProductKey      string   `json:"product_key"`      // Generic installation key / KMS client key
	IsWin11         bool     `json:"is_win11"`
	MSDLProductID   string   `json:"msdl_product_id,omitempty"` // MSDL consumer product ID (e.g. "3813" for Win11 26H2, "2618" for Win10)
	MSDLEvalSlug    string   `json:"msdl_eval_slug,omitempty"`  // MSDL evaluation/enterprise slug (e.g. "win11-ent", "server-2025")
	Badge           string   `json:"badge,omitempty"`           // Badge: "LATEST", "STABLE", "EOL", "WISPY"
}

// OSFamily represents a Windows generation (e.g., Windows 11, Windows 10, Windows 8.1, Windows 7).
type OSFamily struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	Description string    `json:"description"`
	Editions    []Edition `json:"editions"`
}

// Catalog represents the entire collection of Windows versions and editions.
type Catalog struct {
	Families   []OSFamily `json:"families"`
	LastUpdate string     `json:"last_update"`
}
