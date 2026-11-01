package tuning

// ProfileType represents tuning preset profiles.
type ProfileType string

const (
	ProfileBalanced ProfileType = "balanced"
	ProfileGaming   ProfileType = "gaming"
	ProfileCustom   ProfileType = "custom"
)

// HyperVFlags holds Windows Hyper-V enlightenment options.
type HyperVFlags struct {
	Relaxed         bool `json:"relaxed"`
	VAPIC           bool `json:"vapic"`
	Spinlocks       bool `json:"spinlocks"`
	SpinlockRetries int  `json:"spinlock_retries"`
	Time            bool `json:"time"`
	Synic           bool `json:"synic"`
	STimer          bool `json:"stimer"`
	Reset           bool `json:"reset"`
	Frequencies     bool `json:"frequencies"`
	Reenlightenment bool `json:"reenlightenment"`
}

// OptimizationConfig contains performance and quality of life tweaks.
type OptimizationConfig struct {
	Profile             ProfileType `json:"profile"`
	HyperV              HyperVFlags `json:"hyperv"`
	HostCPUPassthrough  bool        `json:"host_cpu_passthrough"`
	CPUTopologySmtMatch bool        `json:"cpu_topology_smt_match"`
	Hugepages           bool        `json:"hugepages"`
	VirtIOSCSI          bool        `json:"virtio_scsi"`
	VirtIONetwork       bool        `json:"virtio_network"`
	VirtIOBalloon       bool        `json:"virtio_balloon"`
	SpiceGuestTools     bool        `json:"spice_guest_tools"`
	AudioBackend        string      `json:"audio_backend"` // "ich9", "hda", "scream"
	TPM2Emulation       bool        `json:"tpm2_emulation"`
	SecureBoot          bool        `json:"secure_boot"`
	BypassTPMAndOOBE    bool        `json:"bypass_tpm_and_oobe"`
	MASActivationPrompt bool        `json:"mas_activation_prompt"`
}
