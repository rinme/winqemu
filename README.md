# WinQEMU ⚡

> **Modern cross-platform GUI and TUI for creating high-performance Windows virtual machines on QEMU / KVM / virt-manager with WispyDocs curated genuine ISOs, automated VirtIO driver injection, Windows 11 TPM/OOBE bypass, and Looking Glass support.**

Built with Go, [Go-Gui](https://github.com/go-gui-org/go-gui) (native GPU-rendered desktop interface), Charm [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

---

## ✨ Features

- **Dual-Mode Interface (GUI & TUI)**:
  - **Desktop GUI**: Native immediate-mode GPU-accelerated window (Metal on macOS, OpenGL on Linux/Windows, zero cgo or web runtimes). Auto-detected in desktop sessions or launched with `--gui`.
  - **Terminal TUI**: Full-screen stepped wizard built with Charm Bubble Tea & Lip Gloss. Auto-detected in terminal/headless environments or forced with `--tui`.
- **Cross-Platform Host Detection**:
  - **Linux**: Auto-detects Arch Linux, Debian, Ubuntu, Fedora, openSUSE, Void, and Alpine. Verifies `qemu`, `libvirt`, `virt-manager`, `swtpm`, and `ovmf`. Offers interactive package installation.
  - **macOS**: Native Hypervisor Framework (`-accel hvf`) with Homebrew integration.
  - **Windows**: Native Windows Hypervisor Platform (`-accel whpx`) with Winget integration.
- **Genuine Windows Media & WispyDocs Recommendations**:
  - Embedded verified catalog of Windows 11, 10, 8.1, and 7 direct from [Massgrave Genuine Installation Media](https://massgrave.dev/genuine-installation-media).
  - Highlights WispyDocs top picks (`⭐ Recommended by WispyDocs`):
    - **Windows 11 IoT Enterprise LTSC 2024 (24H2)**: Zero consumer bloat, no AI telemetry overhead, 10-year support lifecycle.
    - **Windows 11 Enterprise LTSC 2024 (24H2)** & **Windows 11 Pro 24H2**.
    - **Windows 10 IoT Enterprise LTSC 2021 (21H2)**: Supported until 2032 with ultra-low DPC latency.
- **In-App Resumable Downloader & ISO Cache**:
  - Resumable multi-gigabyte HTTP engine with live progress bars and SHA-256 verification.
  - Automatically recognizes cached ISOs (`[Cached - Ready]`) to skip redundant downloads.
  - Option to supply an existing local ISO file.
- **Automated Driver & Win11 TPM / OOBE Bypass**:
  - Pure Go FAT12 floppy driver image builder (`unattend.vfd`).
  - Seamlessly injects VirtIO storage/network drivers and applies registry bypasses for **TPM 2.0, Secure Boot, RAM, and CPU checks**, plus offline local account creation.
  - Optional automated first-logon Microsoft Activation Scripts (MAS).
- **Performance Tuning & Looking Glass**:
  - **Maximum Gaming & Low-Latency Profile**:
    - Full Hyper-V enlightenments (`hv_relaxed`, `hv_vapic`, `hv_spinlocks=0x1fff`, `hv_time`, `hv_synic`, etc.).
    - Host CPU passthrough with exact socket/core/thread topology matching host CCX.
    - **Looking Glass IVSHMEM Support**: Shared memory device (`/dev/shm/looking-glass`) for ultra-low latency, high-refresh display streaming.
    - VirtIO SCSI controller with high-throughput native async I/O.
    - Choice of audio backends: **Scream** (network low-latency gaming audio), **ICH9**, or **Intel HDA**.
  - **Balanced Profile**: Standard virtualization with memory ballooning.
- **VM Manager & Custom Storage**:
  - Built-in VM Manager to scan, list, and directly launch previously created WinQEMU virtual machines.
  - Customizable storage directory to locate virtual machines on fast secondary NVMe drives.
- **Libvirt & Virt-Manager Integration**:
  - On Linux, automatically writes a Libvirt Domain XML definition and registers it via `virsh define`, making the VM instantly manageable in `virt-manager`.
  - Also outputs standalone, portable `start-vm.sh` / `start-vm.bat` scripts.

---

## 🚀 Quick Start

### Prerequisites
- Go 1.22 or newer

### Building and Running
```bash
# Clone the repository
git clone https://github.com/rinme/winqemu.git
cd winqemu

# Build the executable
make build

# Launch (auto-detects GUI in desktop environments, TUI in terminal)
./bin/winqemu

# Or force a specific mode:
./bin/winqemu --gui    # Launch native desktop GUI
./bin/winqemu --tui    # Launch terminal wizard
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
