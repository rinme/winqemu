package msdl

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"winqemu/pkg/catalog"
)

const (
	DefaultAPIBase = "https://api.msdl.tech-latest.com"
	DefaultTimeout = 12 * time.Second
)

// Product represents a Windows release catalogued on MSDL.
type Product struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Badge string   `json:"badge,omitempty"`
	Archs []string `json:"archs,omitempty"`
}

// LanguageSKU represents a localized edition available for a product.
type LanguageSKU struct {
	ID           string `json:"Id"`
	Language     string `json:"Language"`
	Localized    string `json:"LocalizedLanguage"`
	DisplayName  string `json:"ProductDisplayName"`
}

// DownloadOption represents a direct download link returned by MSDL proxy.
type DownloadOption struct {
	Name         string `json:"Name"`
	URI          string `json:"Uri"`
	Language     string `json:"Language"`
	DownloadType int    `json:"DownloadType"` // 0: x86, 1: x64, 2: ARM64
}

// EvalLink represents an enterprise or server evaluation download option.
type EvalLink struct {
	Arch string `json:"arch"`
	Lang string `json:"lang"`
	URL  string `json:"url"`
}

type skuInfoResponse struct {
	Skus []LanguageSKU `json:"Skus"`
}

type proxyResponse struct {
	ProductDownloadOptions []DownloadOption `json:"ProductDownloadOptions"`
	Error                  string           `json:"error,omitempty"`
}

type evalLinksResponse struct {
	Product string     `json:"product"`
	Name    string     `json:"name"`
	Links   []EvalLink `json:"links"`
	Error   string     `json:"error,omitempty"`
}

// Client interacts with the MSDL backend API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient returns a new MSDL API client.
func NewClient() *Client {
	base := DefaultAPIBase
	if env := os.Getenv("MSDL_API_URL"); env != "" {
		base = strings.TrimRight(env, "/")
	}
	return &Client{
		baseURL: base,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// GetProducts returns the catalog of Windows products from MSDL.
func (c *Client) GetProducts(ctx context.Context) ([]Product, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/data/products.json", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WinQEMU/1.0 (X11; Linux x86_64)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch products from MSDL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MSDL API returned HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var raw map[string]struct {
		Name  string   `json:"name"`
		Badge string   `json:"badge"`
		Archs []string `json:"archs"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse products JSON: %w", err)
	}

	var products []Product
	for id, p := range raw {
		products = append(products, Product{
			ID:    id,
			Name:  p.Name,
			Badge: p.Badge,
			Archs: p.Archs,
		})
	}
	return products, nil
}

// GetLanguageSKUs retrieves the list of language SKUs for a consumer product ID.
func (c *Client) GetLanguageSKUs(ctx context.Context, productID string) ([]LanguageSKU, error) {
	u := fmt.Sprintf("%s/skuinfo?product_id=%s", c.baseURL, productID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WinQEMU/1.0 (X11; Linux x86_64)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch SKUs: %w", err)
	}
	defer resp.Body.Close()

	var data skuInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode SKUs: %w", err)
	}
	return data.Skus, nil
}

// GetDownloadLinks queries the MSDL signed proxy for direct Microsoft CDN links.
func (c *Client) GetDownloadLinks(ctx context.Context, productID, skuID string) ([]DownloadOption, error) {
	u := fmt.Sprintf("%s/proxy?product_id=%s&sku_id=%s", c.baseURL, productID, skuID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WinQEMU/1.0 (X11; Linux x86_64)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("proxy request failed: %w", err)
	}
	defer resp.Body.Close()

	var data proxyResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode proxy response: %w", err)
	}

	if data.Error != "" {
		return nil, fmt.Errorf("MSDL proxy error: %s", data.Error)
	}
	if len(data.ProductDownloadOptions) == 0 {
		return nil, fmt.Errorf("no download options returned by Microsoft CDN")
	}

	return data.ProductDownloadOptions, nil
}

// GetEvalLinks queries MSDL for enterprise or server evaluation download options.
func (c *Client) GetEvalLinks(ctx context.Context, slug string) ([]EvalLink, error) {
	u := fmt.Sprintf("%s/evallinks?product=%s", c.baseURL, slug)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "WinQEMU/1.0 (X11; Linux x86_64)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch eval links: %w", err)
	}
	defer resp.Body.Close()

	var data evalLinksResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode eval links: %w", err)
	}

	if data.Error != "" {
		return nil, fmt.Errorf("MSDL eval error: %s", data.Error)
	}
	return data.Links, nil
}

// ResolveConsumerLink attempts to resolve a live signed Microsoft CDN link for a product ID and language.
func (c *Client) ResolveConsumerLink(ctx context.Context, productID, targetLang, targetArch string) (string, error) {
	skus, err := c.GetLanguageSKUs(ctx, productID)
	if err != nil {
		return "", err
	}

	var matchedSKU *LanguageSKU
	normTarget := strings.ToLower(targetLang)
	for i, s := range skus {
		norm := strings.ToLower(s.Language)
		if norm == normTarget || strings.Contains(norm, normTarget) {
			matchedSKU = &skus[i]
			break
		}
	}
	if matchedSKU == nil {
		// Fallback: look for "English"
		for i, s := range skus {
			if strings.EqualFold(s.Language, "English") {
				matchedSKU = &skus[i]
				break
			}
		}
	}
	if matchedSKU == nil && len(skus) > 0 {
		matchedSKU = &skus[0]
	}
	if matchedSKU == nil {
		return "", fmt.Errorf("no language SKU found for product %s", productID)
	}

	options, err := c.GetDownloadLinks(ctx, productID, matchedSKU.ID)
	if err != nil {
		return "", err
	}

	// Filter by architecture
	// DownloadType: 0 = x86, 1 = x64, 2 = ARM64
	for _, opt := range options {
		if strings.EqualFold(targetArch, "x64") && opt.DownloadType == 1 {
			return opt.URI, nil
		}
		if (strings.EqualFold(targetArch, "x86") || targetArch == "32") && opt.DownloadType == 0 {
			return opt.URI, nil
		}
		if strings.EqualFold(targetArch, "arm64") && opt.DownloadType == 2 {
			return opt.URI, nil
		}
	}

	// If no type matched, return the first one
	return options[0].URI, nil
}

// ResolveEvalLink attempts to resolve an Enterprise/Server evaluation link.
func (c *Client) ResolveEvalLink(ctx context.Context, slug, targetLang, targetArch string) (string, error) {
	links, err := c.GetEvalLinks(ctx, slug)
	if err != nil {
		return "", err
	}

	normLang := strings.ToLower(targetLang)
	normArch := strings.ToLower(targetArch)

	for _, l := range links {
		if (strings.Contains(strings.ToLower(l.Lang), normLang) || strings.EqualFold(l.Lang, "en-us")) &&
			strings.EqualFold(l.Arch, normArch) {
			return l.URL, nil
		}
	}

	if len(links) > 0 {
		return links[0].URL, nil
	}
	return "", fmt.Errorf("no evaluation link found for slug %s", slug)
}

// ResolveLiveDownloadURL determines the fastest active download URL for an edition.
// If the edition has MSDL identifiers, it queries https://api.msdl.tech-latest.com
// for fresh signed Microsoft Azure CDN URLs. If MSDL is unreachable or locked,
// it gracefully falls back to the static verified direct mirror URL.
func ResolveLiveDownloadURL(ctx context.Context, ed *catalog.Edition) (string, string) {
	if ed == nil {
		return "", "unknown"
	}

	client := NewClient()

	// 1. Consumer product via MSDL proxy
	if ed.MSDLProductID != "" {
		resolveCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
		defer cancel()

		url, err := client.ResolveConsumerLink(resolveCtx, ed.MSDLProductID, ed.Language, ed.Architecture)
		if err == nil && url != "" {
			return url, "msdl-microsoft-cdn"
		}
		fmt.Fprintf(os.Stderr, "Notice: MSDL live link resolution (%s): %v. Using verified mirror fallback.\n", ed.MSDLProductID, err)
	}

	// 2. Enterprise / Server Evaluation via MSDL
	if ed.MSDLEvalSlug != "" {
		resolveCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
		defer cancel()

		url, err := client.ResolveEvalLink(resolveCtx, ed.MSDLEvalSlug, ed.Language, ed.Architecture)
		if err == nil && url != "" {
			return url, "msdl-microsoft-cdn"
		}
		fmt.Fprintf(os.Stderr, "Notice: MSDL eval link resolution (%s): %v. Using verified mirror fallback.\n", ed.MSDLEvalSlug, err)
	}

	// 3. Fallback to direct catalog download URL
	return ed.DownloadURL, "direct-mirror"
}
