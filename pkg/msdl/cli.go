package msdl

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"winqemu/pkg/downloader"
)

// KnownProducts list fallback when offline
var staticProducts = []Product{
	{ID: "3813", Name: "Windows 11 26H2 (26300.9457)", Badge: "LATEST", Archs: []string{"x64"}},
	{ID: "3816", Name: "Windows 11 Arm64 26H2 (26300.9457)", Badge: "LATEST", Archs: []string{"ARM64"}},
	{ID: "3262", Name: "Windows 11 25H2 (26200.6584)", Badge: "STABLE", Archs: []string{"x64"}},
	{ID: "3321", Name: "Windows 11 25H2 (V2)", Badge: "STABLE", Archs: []string{"x64"}},
	{ID: "3113", Name: "Windows 11 24H2 (26100.1742)", Badge: "STABLE", Archs: []string{"x64"}},
	{ID: "2618", Name: "Windows 10 22H2 v1 (19045.2965)", Badge: "EOL", Archs: []string{"x64", "x86"}},
	{ID: "52", Name: "Windows 8.1 (9600.17415)", Badge: "LEGACY", Archs: []string{"x64", "x86"}},
}

// KnownEvalProducts list
type EvalMeta struct {
	Slug string
	Name string
}

var staticEval = []EvalMeta{
	{Slug: "win11-ent", Name: "Windows 11 Enterprise (Evaluation)"},
	{Slug: "server-2025", Name: "Windows Server 2025 (Standard / Datacenter)"},
	{Slug: "server-2022", Name: "Windows Server 2022 (Standard / Datacenter)"},
	{Slug: "server-2019", Name: "Windows Server 2019 (Standard / Datacenter)"},
	{Slug: "server-2016", Name: "Windows Server 2016 (Standard / Datacenter)"},
}

// RunCLI handles CLI invocations of `winqemu msdl [args...]`.
func RunCLI(args []string) error {
	fs := flag.NewFlagSet("msdl", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, `winqemu msdl — Direct Windows ISO downloader via MSDL (https://msdl.tech-latest.com)

Usage:
  winqemu msdl                                  interactive: select product and language
  winqemu msdl [query]                          search products by keyword (e.g. "winqemu msdl 26H2")
  winqemu msdl --id 3813                        fetch Windows 11 26H2 (prompts for language)
  winqemu msdl --id 3813 --lang "English"       print direct signed Microsoft CDN URL to stdout
  winqemu msdl --eval win11-ent                 fetch Windows 11 Enterprise evaluation ISO
  winqemu msdl --eval server-2025               fetch Windows Server 2025 ISO
  winqemu msdl --list                           list all available products (consumer & enterprise)
  winqemu msdl --download                       immediately download the resolved ISO using WinQEMU downloader

Flags:`)
		fs.PrintDefaults()
	}

	productID := fs.String("id", "", "Consumer product ID (e.g. 3813 for Win11 26H2, 2618 for Win10)")
	langFlag := fs.String("lang", "", `Language name (e.g. "English", "German", "French")`)
	evalSlug := fs.String("eval", "", "Evaluation product slug (e.g. win11-ent, server-2025, server-2022)")
	listMode := fs.Bool("list", false, "List all available products and exit")
	downloadMode := fs.Bool("download", false, "Immediately download the resolved ISO to ~/.cache/winqemu/iso/")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	client := NewClient()
	ctx := context.Background()

	if *listMode {
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println(" MSDL Catalog (Microsoft Software Download Listing)")
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println("\n[Consumer Products]")
		products, err := client.GetProducts(ctx)
		if err != nil || len(products) == 0 {
			products = staticProducts
		}
		for _, p := range products {
			badge := p.Badge
			if badge == "" {
				badge = "STABLE"
			}
			fmt.Printf("  ID: %-6s  [%-7s]  %s\n", p.ID, badge, p.Name)
		}

		fmt.Println("\n[Evaluation & Enterprise Products]")
		for _, e := range staticEval {
			fmt.Printf("  Slug: %-14s  %s\n", e.Slug, e.Name)
		}
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		return nil
	}

	// 1. Eval Mode
	if *evalSlug != "" {
		fmt.Fprintf(os.Stderr, "Fetching enterprise/evaluation links for %s...\n", *evalSlug)
		links, err := client.GetEvalLinks(ctx, *evalSlug)
		if err != nil {
			return fmt.Errorf("failed to fetch eval links: %w", err)
		}

		var targetLink *EvalLink
		for i, l := range links {
			if *langFlag != "" && strings.EqualFold(l.Lang, *langFlag) {
				targetLink = &links[i]
				break
			}
			if targetLink == nil && (strings.EqualFold(l.Lang, "en-us") || strings.EqualFold(l.Arch, "x64")) {
				targetLink = &links[i]
			}
		}
		if targetLink == nil && len(links) > 0 {
			targetLink = &links[0]
		}
		if targetLink == nil {
			return fmt.Errorf("no links found for eval slug %s", *evalSlug)
		}

		fmt.Printf("%s\n", targetLink.URL)
		if *downloadMode {
			return executeDownload(targetLink.URL, fmt.Sprintf("%s-%s.iso", *evalSlug, targetLink.Arch))
		}
		return nil
	}

	// 2. Direct ID or Search/Interactive Mode
	selectedID := *productID
	query := strings.Join(fs.Args(), " ")

	if selectedID == "" {
		products, err := client.GetProducts(ctx)
		if err != nil || len(products) == 0 {
			products = staticProducts
		}

		var matches []Product
		if query != "" {
			for _, p := range products {
				if strings.Contains(strings.ToLower(p.Name), strings.ToLower(query)) || p.ID == query {
					matches = append(matches, p)
				}
			}
		} else {
			matches = products
		}

		if len(matches) == 0 {
			return fmt.Errorf("no products found matching query %q", query)
		}

		if len(matches) == 1 {
			selectedID = matches[0].ID
			fmt.Fprintf(os.Stderr, "Selected: %s (ID: %s)\n", matches[0].Name, matches[0].ID)
		} else {
			fmt.Fprintln(os.Stderr, "\nSelect a Windows Product:")
			for i, p := range matches {
				fmt.Fprintf(os.Stderr, "  [%d] %s (%s)\n", i+1, p.Name, p.ID)
			}
			fmt.Fprintf(os.Stderr, "Enter choice [1-%d]: ", len(matches))
			scanner := bufio.NewScanner(os.Stdin)
			if !scanner.Scan() {
				return fmt.Errorf("no input provided")
			}
			idx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
			if err != nil || idx < 1 || idx > len(matches) {
				return fmt.Errorf("invalid selection")
			}
			selectedID = matches[idx-1].ID
		}
	}

	// 3. Resolve Language SKU
	skus, err := client.GetLanguageSKUs(ctx, selectedID)
	if err != nil {
		return fmt.Errorf("failed to fetch language options for product %s: %w", selectedID, err)
	}

	var selectedSKU *LanguageSKU
	if *langFlag != "" {
		for i, s := range skus {
			if strings.EqualFold(s.Language, *langFlag) || strings.Contains(strings.ToLower(s.Language), strings.ToLower(*langFlag)) {
				selectedSKU = &skus[i]
				break
			}
		}
	}

	if selectedSKU == nil {
		// If only 1 SKU or non-interactive default English
		for i, s := range skus {
			if strings.EqualFold(s.Language, "English") {
				selectedSKU = &skus[i]
				break
			}
		}
		if selectedSKU == nil && len(skus) > 0 {
			selectedSKU = &skus[0]
		}
	}

	if selectedSKU == nil {
		return fmt.Errorf("no language SKU found for product %s", selectedID)
	}

	fmt.Fprintf(os.Stderr, "Requesting signed Microsoft CDN link for %s (%s)...\n", selectedID, selectedSKU.Language)
	options, err := client.GetDownloadLinks(ctx, selectedID, selectedSKU.ID)
	if err != nil {
		return fmt.Errorf("failed to retrieve direct Microsoft CDN link: %w", err)
	}

	var targetURI string
	for _, opt := range options {
		if opt.DownloadType == 1 { // x64
			targetURI = opt.URI
			break
		}
	}
	if targetURI == "" && len(options) > 0 {
		targetURI = options[0].URI
	}

	fmt.Printf("%s\n", targetURI)
	if *downloadMode {
		return executeDownload(targetURI, fmt.Sprintf("msdl-product-%s-x64.iso", selectedID))
	}

	return nil
}

func executeDownload(url, filename string) error {
	dest := fmt.Sprintf("%s/%s", downloader.GetDefaultCacheDir(), filename)
	fmt.Fprintf(os.Stderr, "\nDownloading ISO to: %s\n", dest)
	dl := downloader.NewDownloader()
	dl.MinBytes = 100 * 1024 * 1024

	lastPercent := -1.0
	return dl.Download(context.Background(), url, dest, "", func(p downloader.Progress) {
		if int(p.Percent) != int(lastPercent) {
			lastPercent = p.Percent
			fmt.Fprintf(os.Stderr, "\rProgress: %.1f%% (%.2f / %.2f GB) • Speed: %.1f MB/s",
				p.Percent,
				float64(p.BytesDownloaded)/(1024*1024*1024),
				float64(p.TotalBytes)/(1024*1024*1024),
				p.SpeedBytesSec/(1024*1024),
			)
		}
		if p.Done {
			fmt.Fprintln(os.Stderr, "\n✔ Download complete!")
		}
	})
}
