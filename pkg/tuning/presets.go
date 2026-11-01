package tuning

// GetPreset returns pre-configured optimization settings for a given profile.
func GetPreset(profile ProfileType, isWin11 bool) OptimizationConfig {
	switch profile {
	case ProfileGaming:
		return OptimizationConfig{
			Profile: ProfileGaming,
			HyperV: HyperVFlags{
				Relaxed:         true,
				VAPIC:           true,
				Spinlocks:       true,
				SpinlockRetries: 8191,
				Time:            true,
				Synic:           true,
				STimer:          true,
				Reset:           true,
				Frequencies:     true,
				Reenlightenment: true,
			},
			HostCPUPassthrough:  true,
			CPUTopologySmtMatch: true,
			Hugepages:           false, // Optional to avoid failing if host hugepages are not configured
			VirtIOSCSI:          true,
			VirtIONetwork:       true,
			VirtIOBalloon:       false, // Ballooning adds latency during heavy gaming
			SpiceGuestTools:     true,
			AudioBackend:        "scream",
			LookingGlassEnabled: true,
			LookingGlassMemMB:   64,
			TPM2Emulation:       isWin11,
			SecureBoot:          isWin11,
			BypassTPMAndOOBE:    isWin11,
			MASActivationPrompt: true,
		}
	case ProfileBalanced:
		fallthrough
	default:
		return OptimizationConfig{
			Profile: ProfileBalanced,
			HyperV: HyperVFlags{
				Relaxed:         true,
				VAPIC:           true,
				Spinlocks:       true,
				SpinlockRetries: 8191,
				Time:            true,
				Synic:           false,
				STimer:          false,
				Reset:           false,
				Frequencies:     false,
				Reenlightenment: false,
			},
			HostCPUPassthrough:  true,
			CPUTopologySmtMatch: true,
			Hugepages:           false,
			VirtIOSCSI:          true,
			VirtIONetwork:       true,
			VirtIOBalloon:       true,
			SpiceGuestTools:     true,
			AudioBackend:        "ich9",
			LookingGlassEnabled: false,
			LookingGlassMemMB:   32,
			TPM2Emulation:       isWin11,
			SecureBoot:          isWin11,
			BypassTPMAndOOBE:    isWin11,
			MASActivationPrompt: false,
		}
	}
}
