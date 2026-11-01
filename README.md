# WinQEMU ⚡

> **Modern cross-platform TUI for setting up high-performance Windows virtual machines on QEMU / KVM / virt-manager with WispyDocs curated genuine ISOs, automated VirtIO driver injection, and Windows 11 TPM/OOBE bypass.**

Built with Go, Charm [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

---

## ✨ Features

- **Cross-Platform Host Detection**:
  - **Linux**: Auto-detects Arch Linux, Debian, Ubuntu, Fedora, openSUSE, Void, and Alpine. Verifies `qemu`, `libvirt`, `virt-manager`, `swtpm`, and `ovmf`. Offers 1-click automatic dependency installation via native package managers.
  - **macOS**: Native Hypervisor Framework (`-accel hvf`) with Homebrew integration.
  - **Windows**: Native Windows Hypervisor Platform (`-accel whpx`) with Winget integration.
- **Genuine Windows Media & WispyDocs Recommendations**:
  - Embedded verified catalog of Windows 11, 10, 8.1, and 7 direct from [Massgrave Genuine Installation Media](https://massgrave.dev/genuine-installation-media).
  - Highlights WispyDocs top picks (`⭐ Recommended by WispyDocs`) with detailed architectural notes:
    - **Windows 11 IoT Enterprise LTSC 2024 (24H2)**: Zero consumer bloat, no AI telemetry overhead, 10-year support lifecycle, optimal for gaming & virtualization.
    - **Windows 11 Enterprise LTSC 2024 (24H2)** & **Windows 11 Pro 24H2**.
    - **Windows 10 IoT Enterprise LTSC 2021 (21H2)**: Supported until 2032 with ultra-low DPC latency.
- **In-App Resumable Downloader**:
  - Built-in multi-gigabyte HTTP download engine supporting `Range` headers, real-time download speed, percentage progress bars, and SHA-256 integrity verification.
  - Option to supply an existing local ISO.
- **Automated Driver & Win11 TPM / OOBE Bypass**:
  - Generates `autounattend.xml` and formats a pure Go FAT12 floppy driver image (`unattend.vfd`).
  - Seamlessly injects VirtIO storage/network drivers and applies registry bypasses for **TPM 2.0, Secure Boot, RAM, and CPU checks**, plus offline local account creation (bypassing Microsoft Account login requirement).
  - Optional automated first-logon Microsoft Activation Scripts (MAS).
- **Performance & Tuning Presets**:
  - **Maximum Gaming & Low-Latency Profile**:
    - Full Hyper-V enlightenments: `hv_relaxed`, `hv_vapic`, `hv_spinlocks=0x1fff`, `hv_time`, `hv_synic`, `hv_stimer`, `hv_reset`, `hv_frequencies`, `hv_reenlightenment`.
    - Host CPU passthrough with exact socket/core/thread topology matching host CCX.
    - VirtIO SCSI controller with high-throughput native async I/O.
    - Low-latency Scream / ICH9 audio backend.
  - **Balanced Profile**: Standard virtualization with memory ballooning.
- **Libvirt & Virt-Manager First**:
  - On Linux, automatically writes a Libvirt Domain XML definition and registers it via `virsh define`, making the VM instantly manageable in `virt-manager`.
  - Also outputs standalone, portable `start-vm.sh` / `start-vm.bat` scripts.

---

## 🚀 Quick Start

### Prerequisites
- Go 1.22 or newer

### Building and Running
```bash
# Clone the repository
git clone https://github.com/winqemu/winqemu.git
cd winqemu

# Build the executable
make build

# Launch the interactive TUI
./bin/winqemu
```

---

## 🖥️ Wizard Workflow

1. **Host Diagnostic**: Checks OS, distribution, CPU topology, available RAM, and virtualization acceleration (`/dev/kvm`, `HVF`, or `WHPX`).
2. **Dependency Verifier**: Shows installed and missing tools (`qemu`, `virsh`, `virt-install`, `virt-manager`, `swtpm`). Press `[i]` to run the native package installation command automatically.
3. **Windows Generation**: Choose between Windows 11, Windows 10, Windows 8.1, or Windows 7.
4. **Edition Selection**: Select from the catalog featuring WispyDocs badges, insights, version tags, and KMS installation keys.
5. **Acquisition Mode**: Download directly via resumable progress bar or select a local ISO file.
6. **Hardware Configuration**: Review auto-detected resources (50% host RAM, host-matched vCPU cores, 64GB sparse qcow2 disk) and adjust as needed.
7. **Performance Tuning**: Toggle Gaming vs Balanced profiles, Win11 bypass, MAS activation, and VirtIO SCSI.
8. **Summary & Launch**: Review created artifacts and hit `[Enter]` to open directly in `virt-manager` or boot via QEMU.

---

## 📁 Project Architecture

```
.
├── cmd/
│   └── winqemu/          # Application entry point
├── internal/
│   └── tui/              # Charm Bubble Tea TUI state machine & steps
├── pkg/
│   ├── catalog/          # Embedded WispyDocs / Massgrave ISO index
│   ├── downloader/       # Resumable HTTP download & SHA-256 verification
│   ├── sysinfo/          # OS, distro, and package manager detector
│   ├── tuning/           # Performance presets & Hyper-V enlightenments
│   ├── unattend/         # Autounattend.xml & FAT12 driver disk builder
│   └── vm/               # QEMU arguments, Libvirt XML, and disk creator
├── docs/superpowers/     # Implementation plans and design specifications
├── Makefile
└── go.mod
```

---

## 📄 License
MIT License.
