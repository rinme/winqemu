package gui

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/webview/webview_go"

	"winqemu/pkg/catalog"
	"winqemu/pkg/downloader"
	"winqemu/pkg/msdl"
	"winqemu/pkg/sysinfo"
	"winqemu/pkg/tuning"
	"winqemu/pkg/unattend"
	"winqemu/pkg/vm"
)

//go:embed assets/*
var embeddedAssets embed.FS

type DeployConfigInput struct {
	Name         string `json:"name"`
	RAMMB        int    `json:"ram_mb"`
	CPUCores     int    `json:"cpu_cores"`
	DiskSizeGB   int    `json:"disk_size_gb"`
	StorageDir   string `json:"storage_dir"`
	UseLocalISO  bool   `json:"use_local_iso"`
	LocalISOPath string `json:"local_iso_path"`
	EditionID    string `json:"edition_id"`
	EditionArch  string `json:"edition_arch"`
	BypassTPM    bool   `json:"bypass_tpm"`
	LookingGlass bool   `json:"looking_glass"`
	VirtIOSCSI   bool   `json:"virtio_scsi"`
	MASActivate  bool   `json:"mas_activate"`
	AudioBackend string `json:"audio_backend"`
}

// RunGUI initializes the embedded HTTP asset server and launches the modern desktop Webview window.
func RunGUI() error {
	// 1. Strip the "assets" prefix so root matches "/"
	assetsSub, err := fs.Sub(embeddedAssets, "assets")
	if err != nil {
		return fmt.Errorf("failed to locate embedded assets: %w", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to bind loopback listener: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	server := &http.Server{
		Handler: http.FileServer(http.FS(assetsSub)),
	}
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Close()

	// 2. Initialize Native Webview
	w := webview.New(false)
	if w == nil {
		return fmt.Errorf("could not initialize native webview engine")
	}
	defer w.Destroy()

	w.SetTitle("WinQEMU ⚡ Virtual Machine Studio")
	w.SetSize(1160, 780, webview.HintNone)

	cat, _ := catalog.LoadCatalog()
	host := sysinfo.DetectHost()

	// 3. Bind Go Functions to JavaScript window.*
	w.Bind("getHostInfo", func() string {
		h := sysinfo.DetectHost()
		b, _ := json.Marshal(h)
		return string(b)
	})

	w.Bind("getCatalog", func() string {
		if cat == nil {
			cat, _ = catalog.LoadCatalog()
		}
		b, _ := json.Marshal(cat)
		return string(b)
	})

	w.Bind("getExistingVMs", func(storageDir string) string {
		vms, _ := vm.ScanExistingVMs(storageDir)
		b, _ := json.Marshal(vms)
		return string(b)
	})

	w.Bind("checkISOCache", func(editionID, arch string) string {
		cached, path, size := downloader.IsISOCached(editionID, arch, "")
		res := map[string]interface{}{
			"cached": cached,
			"path":   path,
			"size":   size,
		}
		b, _ := json.Marshal(res)
		return string(b)
	})

	w.Bind("installPackages", func() {
		_ = LaunchTerminalInstaller(host)
	})

	w.Bind("launchVM", func(vmName string) {
		_ = vm.LaunchVirtManager(vmName)
	})

	w.Bind("startDownload", func(editionID, arch string) {
		go func() {
			ed := cat.GetEdition(editionID)
			if ed == nil {
				w.Dispatch(func() {
					w.Eval("window.onDownloadProgress(0, '0 MB/s', false, 'Edition not found')")
				})
				return
			}

			home, _ := os.UserHomeDir()
			cacheDir := filepath.Join(home, ".cache", "winqemu", "iso")
			_ = os.MkdirAll(cacheDir, 0755)
			dest := filepath.Join(cacheDir, fmt.Sprintf("%s-%s.iso", ed.ID, ed.Architecture))

			// Check if already cached genuinely (> 100MB)
			if cached, _, _ := downloader.IsISOCached(ed.ID, ed.Architecture, cacheDir); cached {
				w.Dispatch(func() {
					w.Eval("window.onDownloadProgress(100.0, 'Cached', true, '')")
				})
				return
			}

			activeURL, provider := msdl.ResolveLiveDownloadURL(context.Background(), ed)
			if activeURL == "" {
				w.Dispatch(func() {
					w.Eval("window.onDownloadProgress(0, '0 MB/s', false, 'No download URL configured or resolved')")
				})
				return
			}

			dl := downloader.NewDownloader()
			dl.MinBytes = 100 * 1024 * 1024
			err := dl.Download(context.Background(), activeURL, dest, "", func(p downloader.Progress) {
				w.Dispatch(func() {
					w.Eval(fmt.Sprintf("window.onDownloadProgress(%.1f, '%.1f MB/s', false, '', %.2f, %.2f, %q)",
						p.Percent, p.SpeedBytesSec/(1024*1024),
						float64(p.BytesDownloaded)/(1024*1024*1024),
						float64(p.TotalBytes)/(1024*1024*1024),
						provider))
				})
			})

			w.Dispatch(func() {
				if err != nil {
					w.Eval(fmt.Sprintf("window.onDownloadProgress(0, '0 MB/s', false, %q)", err.Error()))
				} else {
					w.Eval("window.onDownloadProgress(100.0, 'Done', true, '')")
				}
			})
		}()
	})

	w.Bind("deployVM", func(cfgJSON string) {
		go func() {
			var in DeployConfigInput
			if err := json.Unmarshal([]byte(cfgJSON), &in); err != nil {
				w.Dispatch(func() {
					w.Eval(fmt.Sprintf("window.onDeployProgress('', 0, true, null, 'Invalid configuration: %s')", err.Error()))
				})
				return
			}

			w.Dispatch(func() {
				w.Eval("window.onDeployProgress('Initializing VM storage directory...', 15, false, null, '')")
			})

			storageDir := in.StorageDir
			if storageDir == "" {
				storageDir = vm.GetDefaultStorageDir()
			}
			vmName := strings.TrimSpace(in.Name)
			if vmName == "" {
				vmName = "Windows-VM"
			}
			vmDir := filepath.Join(storageDir, vmName)
			_ = os.MkdirAll(vmDir, 0755)

			ed := cat.GetEdition(in.EditionID)
			if ed == nil && len(cat.Families) > 0 && len(cat.Families[0].Editions) > 0 {
				ed = &cat.Families[0].Editions[0]
			}

			isWin11 := true
			if ed != nil {
				isWin11 = ed.IsWin11
			}

			tuningCfg := tuning.GetPreset(tuning.ProfileGaming, isWin11)
			tuningCfg.BypassTPMAndOOBE = in.BypassTPM
			tuningCfg.LookingGlassEnabled = in.LookingGlass
			tuningCfg.VirtIOSCSI = in.VirtIOSCSI
			tuningCfg.MASActivationPrompt = in.MASActivate
			if in.AudioBackend != "" {
				tuningCfg.AudioBackend = in.AudioBackend
			}

			w.Dispatch(func() {
				w.Eval("window.onDeployProgress('Generating autounattend.xml and FAT driver disk...', 35, false, null, '')")
			})

			xmlContent, err := unattend.GenerateAutounattendXML(*ed, tuningCfg, "User", "")
			if err != nil {
				w.Dispatch(func() {
					w.Eval(fmt.Sprintf("window.onDeployProgress('', 0, true, null, 'Failed to generate autounattend: %s')", err.Error()))
				})
				return
			}

			_, floppyPath, err := unattend.WriteUnattendFiles(vmDir, xmlContent)
			if err != nil {
				w.Dispatch(func() {
					w.Eval(fmt.Sprintf("window.onDeployProgress('', 0, true, null, 'Failed to write floppy disk: %s')", err.Error()))
				})
				return
			}

			w.Dispatch(func() {
				w.Eval("window.onDeployProgress('Creating sparse qcow2 virtual disk...', 60, false, null, '')")
			})

			diskPath := filepath.Join(vmDir, fmt.Sprintf("%s.qcow2", vmName))
			if _, err := os.Stat(diskPath); os.IsNotExist(err) {
				if err := vm.CreateQCOW2(diskPath, in.DiskSizeGB); err != nil {
					w.Dispatch(func() {
						w.Eval(fmt.Sprintf("window.onDeployProgress('', 0, true, null, 'Failed to create qcow2 disk: %s')", err.Error()))
					})
					return
				}
			}

			isoPath := in.LocalISOPath
			if !in.UseLocalISO {
				home, _ := os.UserHomeDir()
				isoPath = filepath.Join(home, ".cache", "winqemu", "iso", fmt.Sprintf("%s-%s.iso", ed.ID, ed.Architecture))
			}

			w.Dispatch(func() {
				w.Eval("window.onDeployProgress('Generating launcher script and Libvirt domain...', 85, false, null, '')")
			})

			cfg := &vm.VMConfig{
				Name:            vmName,
				RAMMB:           in.RAMMB,
				CPUCores:        in.CPUCores,
				CPUThreads:      1,
				DiskSizeGB:      in.DiskSizeGB,
				DiskPath:        diskPath,
				ISOPath:         isoPath,
				UnattendDisk:    floppyPath,
				OutputDir:       vmDir,
				SelectedEdition: *ed,
				Tuning:          tuningCfg,
				UseLibvirt:      host.OS == "linux",
			}

			scriptPath, err := vm.GenerateLaunchScript(cfg)
			if err != nil {
				w.Dispatch(func() {
					w.Eval(fmt.Sprintf("window.onDeployProgress('', 0, true, null, 'Failed to generate launch script: %s')", err.Error()))
				})
				return
			}

			res := &vm.CreationResult{
				VMName:           vmName,
				DiskPath:         diskPath,
				DriverDiskPath:   floppyPath,
				LaunchScriptPath: scriptPath,
			}

			if host.OS == "linux" {
				xmlPath, err := vm.WriteLibvirtXML(cfg)
				if err == nil {
					res.LibvirtXMLPath = xmlPath
					if err := vm.RegisterLibvirtDomain(xmlPath); err == nil {
						res.IsLibvirtReady = true
					}
				}
			}

			resJSON, _ := json.Marshal(res)
			w.Dispatch(func() {
				w.Eval(fmt.Sprintf("window.onDeployProgress('Completed!', 100, true, '%s', '')", string(resJSON)))
			})
		}()
	})

	// 4. Navigate Webview to embedded app URL
	w.Navigate(fmt.Sprintf("http://127.0.0.1:%d/index.html", port))
	w.Run()

	return nil
}
