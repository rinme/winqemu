package vm

import (
	"winqemu/pkg/catalog"
	"winqemu/pkg/tuning"
)

// VMConfig contains everything needed to define and boot a Windows VM.
type VMConfig struct {
	Name            string                    `json:"name"`
	RAMMB           int                       `json:"ram_mb"`
	CPUCores        int                       `json:"cpu_cores"`
	CPUThreads      int                       `json:"cpu_threads"`
	DiskSizeGB      int                       `json:"disk_size_gb"`
	DiskPath        string                    `json:"disk_path"`
	ISOPath         string                    `json:"iso_path"`
	VirtIOISOPath   string                    `json:"virtio_iso_path"`
	UnattendDisk    string                    `json:"unattend_disk"`
	OutputDir       string                    `json:"output_dir"`
	SelectedEdition catalog.Edition           `json:"edition"`
	Tuning          tuning.OptimizationConfig `json:"tuning"`
	UseLibvirt      bool                      `json:"use_libvirt"`
}

// CreationResult summarizes the created VM artifacts.
type CreationResult struct {
	VMName           string   `json:"vm_name"`
	DiskPath         string   `json:"disk_path"`
	DriverDiskPath   string   `json:"driver_disk_path"`
	LibvirtXMLPath   string   `json:"libvirt_xml_path"`
	LaunchScriptPath string   `json:"launch_script_path"`
	CommandsExecuted []string `json:"commands_executed"`
	IsLibvirtReady   bool     `json:"is_libvirt_ready"`
}
