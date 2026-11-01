# WinQEMU TUI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a cross-platform Go TUI application that automates QEMU/KVM/virt-manager Windows VM creation, featuring WispyDocs-curated genuine Windows ISO selection (Windows 7 to 11), automated dependency detection/installation across distros, VirtIO & autounattend.xml driver injection/Win11 bypass, and performance/QoL presets.

**Architecture:** A modular Go application built with Charm Bubble Tea (`tea.Model`), Lip Gloss, and Bubbles. Divided into distinct packages: `sysinfo` (OS/distro detection & package management), `catalog` (embedded + remote WispyDocs/Massgrave ISO index), `downloader` (resumable multi-threaded HTTP downloader with SHA256 checking), `unattend` (autounattend.xml generator & FAT/driver disk builder), `tuning` (performance & QoL flags, Hyper-V enlightenments, topology), `vm` (qcow2 creation, libvirt XML & virt-install orchestration for Linux/virt-manager, and native QEMU scripts for macOS/Windows), and `tui` (Bubble Tea wizard steps).

**Tech Stack:** Go 1.22+, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`, `github.com/charmbracelet/bubbles`, standard library (`os`, `os/exec`, `crypto/sha256`, `archive/tar`, `embed`).

**Spec:** Agreed requirements from the `/grill-me` session:
- Hybrid VM orchestration: Libvirt/virt-manager on Linux, native QEMU scripts on macOS/Windows.
- Curated genuine ISO catalog: WispyDocs recommendations (IoT Enterprise LTSC, Enterprise LTSC, Pro), Massgrave source, Windows 7 to 11.
- In-app downloader with resume, progress bar, and SHA-256 verification + local file selector.
- Automated dependency installer for pacman, apt, dnf, zypper, brew, winget.
- Presets: "Balanced" and "Maximum Gaming / Low-Latency" (Hyper-V flags, CPU pinning/topology, VirtIO, Scream audio).
- Unattended driver disk with `autounattend.xml` bypassing Win 11 TPM 2.0 / Secure Boot / MS Account OOBE.

## Global Constraints
- Must be buildable and testable on modern Go without external CGO dependencies if possible (pure Go for core logic).
- Support Linux (Arch, Ubuntu/Debian, Fedora, openSUSE), macOS, and Windows.
- Keep terminal UI clean, responsive, and styled with Lip Gloss.

---

### Task 1: Project Initialization & Data Models

**Files:**
- Create: `go.mod`
- Create: `pkg/catalog/types.go`
- Create: `pkg/sysinfo/types.go`
- Create: `pkg/tuning/types.go`
- Create: `pkg/vm/types.go`

- [ ] **Step 1: Initialize go.mod and install dependencies**
- [ ] **Step 2: Define Catalog, System, Tuning, and VM configuration structs**
- [ ] **Step 3: Verify module compiles with `go test ./...`**
- [ ] **Step 4: Commit**

---

### Task 2: OS & Distro Detection and Dependency Manager (`pkg/sysinfo`)

**Files:**
- Create: `pkg/sysinfo/detector.go`
- Create: `pkg/sysinfo/packages.go`
- Create: `pkg/sysinfo/detector_test.go`

- [ ] **Step 1: Write tests for OS and Linux distro detection (`/etc/os-release` parsing)**
- [ ] **Step 2: Implement OS detection (Arch, Debian, Ubuntu, Fedora, openSUSE, macOS, Windows)**
- [ ] **Step 3: Implement dependency checking (qemu-system-x86_64, virsh, virt-install, swtpm, ovmf) and install command generation for pacman, apt, dnf, zypper, brew, winget**
- [ ] **Step 4: Run tests and verify with `go test ./pkg/sysinfo/...`**
- [ ] **Step 5: Commit**

---

### Task 3: WispyDocs & Massgrave ISO Catalog (`pkg/catalog`)

**Files:**
- Create: `pkg/catalog/catalog.json`
- Create: `pkg/catalog/catalog.go`
- Create: `pkg/catalog/catalog_test.go`

- [ ] **Step 1: Create catalog data containing Windows 11 (IoT Enterprise LTSC, Enterprise LTSC, Pro, Home with WispyDocs recommendation badges), Windows 10, Windows 8.1, and Windows 7, with official Massgrave links and SHA256 hashes**
- [ ] **Step 2: Embed catalog via `//go:embed` and implement query/filter functions (by OS family, edition, recommendation status)**
- [ ] **Step 3: Implement optional dynamic fetcher to check for updated Massgrave links**
- [ ] **Step 4: Write unit tests verifying catalog validity, lookups, and recommendations**
- [ ] **Step 5: Run tests and commit**

---

### Task 4: In-App Resumable Downloader (`pkg/downloader`)

**Files:**
- Create: `pkg/downloader/downloader.go`
- Create: `pkg/downloader/downloader_test.go`

- [ ] **Step 1: Write unit tests for HTTP downloader with mock server, testing resume (`Range` header) and SHA-256 verification**
- [ ] **Step 2: Implement downloader emitting progress updates (bytes transferred, total bytes, speed, percentage)**
- [ ] **Step 3: Add VirtIO ISO automated fetcher (`virtio-win.iso` latest release from Fedora)**
- [ ] **Step 4: Run tests and commit**

---

### Task 5: Driver Disk & Autounattend.xml Generator (`pkg/unattend`)

**Files:**
- Create: `pkg/unattend/generator.go`
- Create: `pkg/unattend/fatdisk.go`
- Create: `pkg/unattend/generator_test.go`

- [ ] **Step 1: Write unit tests verifying `autounattend.xml` template generation for Windows 7, 10, and 11 (including TPM/RAM bypass registry keys and VirtIO driver search paths)**
- [ ] **Step 2: Implement `autounattend.xml` generator supporting edition selection and bypass flags**
- [ ] **Step 3: Implement pure Go FAT32/ISO secondary disk image creator containing `autounattend.xml`**
- [ ] **Step 4: Run tests and commit**

---

### Task 6: Tuning Presets & QEMU / Libvirt Config Generator (`pkg/tuning` & `pkg/vm`)

**Files:**
- Create: `pkg/tuning/presets.go`
- Create: `pkg/vm/qemu.go`
- Create: `pkg/vm/libvirt.go`
- Create: `pkg/vm/vm_test.go`

- [ ] **Step 1: Implement "Balanced" and "Maximum Gaming / Low-Latency" presets (Hyper-V flags, CPU pinning/topology, hugepages, Scream audio)**
- [ ] **Step 2: Implement Libvirt Domain XML generation and `virt-install` invocation generator for Linux/virt-manager**
- [ ] **Step 3: Implement native standalone QEMU launcher script generation (`run.sh` / `run.bat`) for macOS (`-accel hvf`) and Windows (`-accel whpx`)**
- [ ] **Step 4: Implement `qemu-img` disk creator for qcow2 sparse disks**
- [ ] **Step 5: Write unit tests verifying generated command line arguments and XML correctness**
- [ ] **Step 6: Run tests and commit**

---

### Task 7: Charm Bubble Tea TUI Wizard (`internal/tui`)

**Files:**
- Create: `internal/tui/styles.go`
- Create: `internal/tui/app.go`
- Create: `internal/tui/step_welcome.go`
- Create: `internal/tui/step_depcheck.go`
- Create: `internal/tui/step_edition.go`
- Create: `internal/tui/step_download.go`
- Create: `internal/tui/step_hardware.go`
- Create: `internal/tui/step_tuning.go`
- Create: `internal/tui/step_summary.go`
- Create: `cmd/winqemu/main.go`

- [ ] **Step 1: Implement styling and wizard navigation state machine**
- [ ] **Step 2: Implement Step 1 (Welcome & OS detection) and Step 2 (Dependency check & auto-installer)**
- [ ] **Step 3: Implement Step 3 (WispyDocs Windows ISO selector with badges) and Step 4 (Downloader with live progress bar & local file picker)**
- [ ] **Step 4: Implement Step 5 (Hardware resource review) and Step 6 (Tuning & QoL presets)**
- [ ] **Step 5: Implement Step 7 (Summary, VM generation, and Launch / virt-manager trigger)**
- [ ] **Step 6: Create `main.go` entry point**
- [ ] **Step 7: Build binary with `go build -o bin/winqemu ./cmd/winqemu` and test execution**
- [ ] **Step 8: Commit**

---

### Task 8: End-to-End Verification and Documentation

**Files:**
- Create: `README.md`
- Create: `Makefile`

- [ ] **Step 1: Build release binary and verify `--help` and interactive flags**
- [ ] **Step 2: Run all unit and integration tests across packages**
- [ ] **Step 3: Write comprehensive README explaining cross-platform support, WispyDocs recommendations, and virt-manager integration**
- [ ] **Step 4: Commit and present to user**
