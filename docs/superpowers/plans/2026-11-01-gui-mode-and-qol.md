# WinQEMU GUI Mode & QoL Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a GPU-accelerated desktop GUI mode (`--gui`) built with Go-Gui, auto-detect graphical sessions, and implement major QoL features (VM manager for existing VMs, Looking Glass IVSHMEM support, custom storage directories, audio backend selection, and ISO cache detection) across both TUI and GUI modes.

**Architecture:**
- `internal/gui`: Implements the Go-Gui immediate mode desktop dashboard with persistent sidebar and responsive panels.
- `pkg/vm`: Extended with `ScanExistingVMs`, Looking Glass `ivshmem-plain` QEMU flags, and Libvirt `<shmem>` elements.
- `pkg/downloader`: Extended with cache detection helpers.
- `internal/tui`: Updated with new QoL options (custom storage, Looking Glass, audio choice, VM manager).
- `cmd/winqemu`: CLI entry point supporting `--gui`, `--tui`, `--help`, and auto-detection.

**Tech Stack:** Go 1.22+, `github.com/go-gui-org/go-gui`, `github.com/charmbracelet/bubbletea`, `github.com/charmbracelet/lipgloss`.

---

### Task 1: Core QoL Backend Enhancements (`pkg/vm` & `pkg/downloader`)

**Files:**
- Modify: `pkg/tuning/types.go`
- Modify: `pkg/vm/types.go`
- Modify: `pkg/vm/qemu.go`
- Modify: `pkg/vm/libvirt.go`
- Modify: `pkg/vm/manager.go`
- Modify: `pkg/downloader/downloader.go`
- Modify: `pkg/vm/vm_test.go`

- [x] **Step 1: Add LookingGlassEnabled and LookingGlassMemMB to `tuning.OptimizationConfig`**
- [x] **Step 2: Add `ExistingVM` struct and `ScanExistingVMs(storageDir string)` in `pkg/vm/manager.go`**
- [x] **Step 3: Update `BuildQEMUArgs` and `GenerateLibvirtXML` to emit Looking Glass IVSHMEM configuration when enabled**
- [x] **Step 4: Add `IsISOCached(editionID, arch, cacheDir string) (bool, string)` in `pkg/downloader/downloader.go`**
- [x] **Step 5: Write unit tests verifying `ScanExistingVMs` and Looking Glass QEMU/Libvirt outputs**
- [x] **Step 6: Run tests and commit**

---

### Task 2: Go-Gui Desktop Application (`internal/gui`)

**Files:**
- Create: `internal/gui/state.go`
- Create: `internal/gui/terminal.go`
- Create: `internal/gui/views.go`
- Create: `internal/gui/app.go`

- [x] **Step 1: Define `AppState` holding host info, catalog, active tab/step, VM inputs, and download progress**
- [x] **Step 2: Implement `SpawnInteractiveTerminal` to launch terminal emulators (`alacritty`, `kitty`, `gnome-terminal`, `konsole`, `xterm`) for package installation with Polkit fallback**
- [x] **Step 3: Build Split-View views in `views.go`: Sidebar, System Health, WispyDocs OS & Edition selector, ISO Downloader, Hardware Config (with custom storage path), Tuning (Looking Glass & Audio), and VM Manager (listing existing VMs with launch actions)**
- [x] **Step 4: Implement window initialization and backend runner in `app.go` (`RunGUI()`)**
- [x] **Step 5: Test compilation of `internal/gui`**
- [x] **Step 6: Commit**

---

### Task 3: TUI QoL Enhancements (`internal/tui`)

**Files:**
- Modify: `internal/tui/styles.go`
- Modify: `internal/tui/app.go`
- Modify: `internal/tui/step_welcome.go`
- Modify: `internal/tui/step_hardware.go`
- Modify: `internal/tui/step_tuning.go`
- Modify: `internal/tui/step_edition.go`

- [x] **Step 1: Show existing VMs on the welcome screen with quick launch shortcut `[V]`**
- [x] **Step 2: Add custom storage directory path input to `step_hardware.go`**
- [x] **Step 3: Add Looking Glass toggle and Audio Backend toggle to `step_tuning.go`**
- [x] **Step 4: Add `[Cached - Ready]` indicator to `step_edition.go` when the genuine ISO is already on disk**
- [x] **Step 5: Run tests and verify TUI compilation**
- [x] **Step 6: Commit**

---

### Task 4: CLI Flags & Graphical Session Auto-Detection (`cmd/winqemu`)

**Files:**
- Modify: `cmd/winqemu/main.go`

- [x] **Step 1: Add command-line flag parser (`--gui`, `--tui`, `--help`, `-v`, `--version`)**
- [x] **Step 2: Implement auto-detection: check if running in graphical environment (`DISPLAY` or `WAYLAND_DISPLAY` on Linux, GUI session on macOS/Windows) when no flags are given**
- [x] **Step 3: Wire `--gui` to `gui.RunGUI()` and `--tui` to Bubble Tea runner**
- [x] **Step 4: Build binary and verify `--help` output**
- [x] **Step 5: Commit**

---

### Task 5: End-to-End Verification, Documentation & Push

**Files:**
- Modify: `README.md`
- Modify: `Makefile`

- [x] **Step 1: Run all unit tests across all packages (`make test`)**
- [x] **Step 2: Update README with `--gui` mode instructions and new QoL features**
- [x] **Step 3: Build release binary (`make build`)**
- [x] **Step 4: Push commits to GitHub**
