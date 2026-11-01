package gui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"winqemu/pkg/catalog"
	"winqemu/pkg/downloader"
	"winqemu/pkg/sysinfo"
	"winqemu/pkg/tuning"
	"winqemu/pkg/unattend"
	"winqemu/pkg/vm"
)

type GuiTab int

const (
	TabOverview GuiTab = iota
	TabOSSelection
	TabMedia
	TabHardware
	TabTuning
	TabDeploy
	TabExistingVMs
)

// AppState maintains the state for the Go-Gui immediate-mode window.
type AppState struct {
	ActiveTab       GuiTab
	Host            *sysinfo.HostInfo
	Catalog         *catalog.Catalog
	SelectedFamily  *catalog.OSFamily
	SelectedEdition *catalog.Edition
	ExistingVMs     []vm.ExistingVM

	// Form Inputs
	VMName           string
	RAMMB            string
	CPUCores         string
	DiskGB           string
	CustomStorageDir string
	LocalISOPath     string
	UseLocalISO      bool

	// Tuning Configuration
	TuningConfig tuning.OptimizationConfig

	// Download progress
	IsDownloading    bool
	DownloadProgress float64
	DownloadSpeedStr string
	DownloadedISO    string
	DownloadErr      string
	cancelDownload   context.CancelFunc

	// Deployment progress
	IsDeploying  bool
	DeployResult *vm.CreationResult
	DeployErr    string
	StatusMsg    string
}

// NewAppState initializes the default application state with host telemetry and presets.
func NewAppState() *AppState {
	host := sysinfo.DetectHost()
	cat, _ := catalog.LoadCatalog()

	defRAM := host.TotalRAMMB / 2
	if defRAM < 4096 {
		defRAM = 4096
	}
	defCPU := host.LogicalCPUs / 2
	if defCPU < 2 {
		defCPU = 2
	}
	if defCPU > 8 {
		defCPU = 8
	}

	var fam *catalog.OSFamily
	var ed *catalog.Edition
	if cat != nil && len(cat.Families) > 0 {
		fam = &cat.Families[0]
		if len(fam.Editions) > 0 {
			ed = &fam.Editions[0]
		}
	}

	existing, _ := vm.ScanExistingVMs("")

	state := &AppState{
		ActiveTab:        TabOverview,
		Host:             host,
		Catalog:          cat,
		SelectedFamily:   fam,
		SelectedEdition:  ed,
		ExistingVMs:      existing,
		VMName:           "Windows-11-VM",
		RAMMB:            strconv.Itoa(defRAM),
		CPUCores:         strconv.Itoa(defCPU),
		DiskGB:           "64",
		CustomStorageDir: vm.GetDefaultStorageDir(),
		TuningConfig:     tuning.GetPreset(tuning.ProfileGaming, true),
	}

	if ed != nil {
		state.VMName = strings.ReplaceAll(ed.Name, " ", "-")
	}

	return state
}

// RefreshExistingVMs re-scans the storage directory.
func (s *AppState) RefreshExistingVMs() {
	storage := s.CustomStorageDir
	if storage == "" {
		storage = vm.GetDefaultStorageDir()
	}
	vms, _ := vm.ScanExistingVMs(storage)
	s.ExistingVMs = vms
}

// StartDownload initiates media download in a background goroutine.
func (s *AppState) StartDownload(onUpdate func()) {
	if s.SelectedEdition == nil || s.IsDownloading {
		return
	}

	s.IsDownloading = true
	s.DownloadErr = ""
	s.DownloadProgress = 0

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelDownload = cancel

	home, _ := os.UserHomeDir()
	cacheDir := filepath.Join(home, ".cache", "winqemu", "iso")
	_ = os.MkdirAll(cacheDir, 0755)
	dest := filepath.Join(cacheDir, fmt.Sprintf("%s-%s.iso", s.SelectedEdition.ID, s.SelectedEdition.Architecture))
	s.DownloadedISO = dest

	go func() {
		defer func() {
			s.IsDownloading = false
			if onUpdate != nil {
				onUpdate()
			}
		}()

		// If cached already
		if fi, err := os.Stat(dest); err == nil && fi.Size() > 100*1024*1024 {
			s.DownloadProgress = 1.0
			return
		}

		dl := downloader.NewDownloader()
		err := dl.Download(ctx, s.SelectedEdition.DownloadURL, dest, "", func(p downloader.Progress) {
			s.DownloadProgress = p.Percent / 100.0
			s.DownloadSpeedStr = fmt.Sprintf("%.1f MB/s", p.SpeedBytesSec/(1024*1024))
			if onUpdate != nil {
				onUpdate()
			}
		})

		if err != nil && err != context.Canceled {
			s.DownloadErr = err.Error()
		}
	}()
}

// DeployVM creates disk, generates configuration files, and defines the VM.
func (s *AppState) DeployVM(onDone func()) {
	s.IsDeploying = true
	s.DeployErr = ""
	s.DeployResult = nil

	go func() {
		defer func() {
			s.IsDeploying = false
			s.RefreshExistingVMs()
			if onDone != nil {
				onDone()
			}
		}()

		storageDir := s.CustomStorageDir
		if storageDir == "" {
			storageDir = vm.GetDefaultStorageDir()
		}

		vmName := strings.TrimSpace(s.VMName)
		if vmName == "" {
			vmName = "Windows-VM"
		}

		vmDir := filepath.Join(storageDir, vmName)
		_ = os.MkdirAll(vmDir, 0755)

		ramMB, _ := strconv.Atoi(s.RAMMB)
		if ramMB <= 0 {
			ramMB = 8192
		}
		cpuCores, _ := strconv.Atoi(s.CPUCores)
		if cpuCores <= 0 {
			cpuCores = 4
		}
		diskGB, _ := strconv.Atoi(s.DiskGB)
		if diskGB <= 0 {
			diskGB = 64
		}

		// 1. autounattend.xml & floppy
		xmlContent, err := unattend.GenerateAutounattendXML(*s.SelectedEdition, s.TuningConfig, "User", "")
		if err != nil {
			s.DeployErr = fmt.Sprintf("unattend generation failed: %v", err)
			return
		}

		_, floppyPath, err := unattend.WriteUnattendFiles(vmDir, xmlContent)
		if err != nil {
			s.DeployErr = fmt.Sprintf("floppy disk creation failed: %v", err)
			return
		}

		// 2. qcow2 disk
		diskPath := filepath.Join(vmDir, fmt.Sprintf("%s.qcow2", vmName))
		if _, err := os.Stat(diskPath); os.IsNotExist(err) {
			if err := vm.CreateQCOW2(diskPath, diskGB); err != nil {
				s.DeployErr = fmt.Sprintf("qcow2 disk creation failed: %v", err)
				return
			}
		}

		// 3. VM config
		isoPath := s.DownloadedISO
		if s.UseLocalISO && s.LocalISOPath != "" {
			isoPath = s.LocalISOPath
		}

		cfg := &vm.VMConfig{
			Name:            vmName,
			RAMMB:           ramMB,
			CPUCores:        cpuCores,
			CPUThreads:      1,
			DiskSizeGB:      diskGB,
			DiskPath:        diskPath,
			ISOPath:         isoPath,
			UnattendDisk:    floppyPath,
			OutputDir:       vmDir,
			SelectedEdition: *s.SelectedEdition,
			Tuning:          s.TuningConfig,
			UseLibvirt:      s.Host.OS == "linux",
		}

		scriptPath, err := vm.GenerateLaunchScript(cfg)
		if err != nil {
			s.DeployErr = fmt.Sprintf("launch script failed: %v", err)
			return
		}

		res := &vm.CreationResult{
			VMName:           vmName,
			DiskPath:         diskPath,
			DriverDiskPath:   floppyPath,
			LaunchScriptPath: scriptPath,
		}

		if s.Host.OS == "linux" {
			xmlPath, err := vm.WriteLibvirtXML(cfg)
			if err == nil {
				res.LibvirtXMLPath = xmlPath
				if err := vm.RegisterLibvirtDomain(xmlPath); err == nil {
					res.IsLibvirtReady = true
				}
			}
		}

		s.DeployResult = res
	}()
}
