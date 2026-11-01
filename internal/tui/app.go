package tui

import (
	"context"
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"winqemu/pkg/catalog"
	"winqemu/pkg/downloader"
	"winqemu/pkg/sysinfo"
	"winqemu/pkg/tuning"
	"winqemu/pkg/vm"
)

type WizardStep int

const (
	StepWelcome WizardStep = iota
	StepDepCheck
	StepFamilySelect
	StepEditionSelect
	StepISOSource
	StepDownloadProgress
	StepHardware
	StepTuning
	StepExecution
	StepDone
	StepVMManager
)

type DownloadMsg downloader.Progress
type DownloadDoneMsg struct {
	Err error
}
type VMCreateDoneMsg struct {
	Result *vm.CreationResult
	Err    error
}
type InstallFinishedMsg struct {
	Err error
}

// AppModel is the primary state machine for the Charm Bubble Tea application.
type AppModel struct {
	Step            WizardStep
	Host            *sysinfo.HostInfo
	Catalog         *catalog.Catalog
	SelectedFamily  *catalog.OSFamily
	SelectedEdition *catalog.Edition
	TuningConfig    tuning.OptimizationConfig

	// Form & Navigation State
	Cursor       int
	SubCursor    int
	WindowWidth  int
	WindowHeight int
	StatusMsg    string
	IsLoading    bool
	Err          error

	// Local ISO / Download
	UseLocalISO   bool
	LocalISOInput textinput.Model
	DownloadedISO string
	VirtIOISO     string
	ProgressBar   progress.Model

	// Hardware Inputs
	VMNameInput     textinput.Model
	RAMInput        textinput.Model
	CPUInput        textinput.Model
	DiskInput       textinput.Model
	StorageDirInput textinput.Model

	// Existing VMs
	ExistingVMs []vm.ExistingVM

	// VM Creation Result
	Result *vm.CreationResult

	// Cancel context for downloads
	cancelDownload context.CancelFunc
}

// NewAppModel initializes the TUI with host diagnostics, catalog, and reasonable defaults.
func NewAppModel() (*AppModel, error) {
	host := sysinfo.DetectHost()
	cat, err := catalog.LoadCatalog()
	if err != nil {
		return nil, fmt.Errorf("could not load catalog: %w", err)
	}

	// Initialize inputs
	isoInput := textinput.New()
	isoInput.Placeholder = "/path/to/windows.iso"
	isoInput.CharLimit = 256
	isoInput.Width = 50

	vmNameInput := textinput.New()
	vmNameInput.Placeholder = "Win11-VM"
	vmNameInput.SetValue("Windows-VM")
	vmNameInput.Width = 30

	// Default RAM: half of total host RAM (at least 4096MB)
	defRAM := host.TotalRAMMB / 2
	if defRAM < 4096 {
		defRAM = 4096
	}
	ramInput := textinput.New()
	ramInput.Placeholder = "8192"
	ramInput.SetValue(strconv.Itoa(defRAM))
	ramInput.Width = 10

	// Default CPU cores: half of logical host CPUs (at least 2, max 8)
	defCPU := host.LogicalCPUs / 2
	if defCPU < 2 {
		defCPU = 2
	}
	if defCPU > 8 {
		defCPU = 8
	}
	cpuInput := textinput.New()
	cpuInput.Placeholder = "4"
	cpuInput.SetValue(strconv.Itoa(defCPU))
	cpuInput.Width = 10

	diskInput := textinput.New()
	diskInput.Placeholder = "64"
	diskInput.SetValue("64")
	diskInput.Width = 10

	storageDirInput := textinput.New()
	storageDirInput.Placeholder = vm.GetDefaultStorageDir()
	storageDirInput.SetValue(vm.GetDefaultStorageDir())
	storageDirInput.Width = 40

	prog := progress.New(progress.WithDefaultGradient())

	// Default family: Windows 11
	fam := cat.GetFamily("win11")
	if fam == nil && len(cat.Families) > 0 {
		fam = &cat.Families[0]
	}

	existing, _ := vm.ScanExistingVMs("")

	return &AppModel{
		Step:            StepWelcome,
		Host:            host,
		Catalog:         cat,
		SelectedFamily:  fam,
		TuningConfig:    tuning.GetPreset(tuning.ProfileGaming, true),
		LocalISOInput:   isoInput,
		VMNameInput:     vmNameInput,
		RAMInput:        ramInput,
		CPUInput:        cpuInput,
		DiskInput:       diskInput,
		StorageDirInput: storageDirInput,
		ProgressBar:     prog,
		ExistingVMs:     existing,
	}, nil
}

func (m *AppModel) Init() tea.Cmd {
	return nil
}

func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.WindowWidth = msg.Width
		m.WindowHeight = msg.Height
		m.ProgressBar.Width = msg.Width - 10
		if m.ProgressBar.Width > 70 {
			m.ProgressBar.Width = 70
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.cancelDownload != nil {
				m.cancelDownload()
			}
			return m, tea.Quit
		}

	case DownloadMsg:
		cmd := m.ProgressBar.SetPercent(msg.Percent / 100.0)
		m.StatusMsg = fmt.Sprintf("Downloading... %.1f%% (%.1f MB/s)", msg.Percent, msg.SpeedBytesSec/(1024*1024))
		return m, cmd

	case DownloadDoneMsg:
		m.IsLoading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Step = StepISOSource
			return m, nil
		}
		m.Step = StepHardware
		m.Cursor = 0
		m.VMNameInput.Focus()
		return m, nil

	case VMCreateDoneMsg:
		m.IsLoading = false
		if msg.Err != nil {
			m.Err = msg.Err
		} else {
			m.Result = msg.Result
			m.Step = StepDone
		}
		return m, nil

	case InstallFinishedMsg:
		m.IsLoading = false
		m.Host = sysinfo.DetectHost() // Re-check all dependencies
		if msg.Err != nil {
			m.Err = fmt.Errorf("package manager exited with error: %w", msg.Err)
			m.StatusMsg = ""
		} else {
			m.Err = nil
			m.StatusMsg = "✔ Package installation completed! System dependencies re-checked."
		}
		return m, nil
	}

	// Route based on current step
	switch m.Step {
	case StepWelcome:
		return m.updateWelcome(msg)
	case StepDepCheck:
		return m.updateDepCheck(msg)
	case StepFamilySelect:
		return m.updateFamilySelect(msg)
	case StepEditionSelect:
		return m.updateEditionSelect(msg)
	case StepISOSource:
		return m.updateISOSource(msg)
	case StepDownloadProgress:
		return m.updateDownloadProgress(msg)
	case StepHardware:
		return m.updateHardware(msg)
	case StepTuning:
		return m.updateTuning(msg)
	case StepExecution:
		return m.updateExecution(msg)
	case StepDone:
		return m.updateDone(msg)
	case StepVMManager:
		return m.updateVMManager(msg)
	}

	return m, nil
}

func (m *AppModel) View() string {
	var body string
	switch m.Step {
	case StepWelcome:
		body = m.viewWelcome()
	case StepDepCheck:
		body = m.viewDepCheck()
	case StepFamilySelect:
		body = m.viewFamilySelect()
	case StepEditionSelect:
		body = m.viewEditionSelect()
	case StepISOSource:
		body = m.viewISOSource()
	case StepDownloadProgress:
		body = m.viewDownloadProgress()
	case StepHardware:
		body = m.viewHardware()
	case StepTuning:
		body = m.viewTuning()
	case StepExecution:
		body = m.viewExecution()
	case StepDone:
		body = m.viewDone()
	case StepVMManager:
		body = m.viewVMManager()
	}

	header := HeaderBox.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			TitleStyle.Render("⚡ WinQEMU • Next-Gen Virtual Machine Creator"),
			StepIndicator.Render(fmt.Sprintf("Step %d/8: %s", m.Step+1, m.stepTitle())),
		),
	)

	return lipgloss.JoinVertical(lipgloss.Left, header, body)
}

func (m *AppModel) stepTitle() string {
	switch m.Step {
	case StepWelcome:
		return "Environment & Host Detection"
	case StepDepCheck:
		return "Dependency & Service Check"
	case StepFamilySelect:
		return "Select Windows Generation"
	case StepEditionSelect:
		return "Select Edition (WispyDocs Curated)"
	case StepISOSource:
		return "ISO Acquisition (Download or Local)"
	case StepDownloadProgress:
		return "Downloading Genuine Installation Media"
	case StepHardware:
		return "Configure Hardware Resources"
	case StepTuning:
		return "Performance & QoL Tuning Presets"
	case StepExecution:
		return "Building VM Artifacts & Configuration"
	case StepDone:
		return "VM Ready"
	case StepVMManager:
		return "Existing Virtual Machines Manager"
	default:
		return "Setup"
	}
}
