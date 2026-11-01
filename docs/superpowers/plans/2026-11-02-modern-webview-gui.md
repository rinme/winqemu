# WinQEMU Modern Webview GUI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the legacy GUI with a modern, interactive, Dark Tech glassmorphic desktop interface powered by native Webview (`webview_go`) and embedded Tailwind/Lucide SPA, with real-time Go IPC bridge for downloads, VM management, and virtualization configuration.

**Architecture:**
- `internal/gui/assets/`: Embedded static single-page application (Dark Tech styling, glassmorphism, responsive split-view, Lucide vector icons).
- `internal/gui/app.go`: Webview window manager with two-way Go ⇄ JavaScript IPC binding to domain packages (`pkg/sysinfo`, `pkg/catalog`, `pkg/downloader`, `pkg/vm`, `pkg/tuning`).
- `internal/gui/terminal.go`: Helper for spawning terminal emulator during interactive package installation.
- `cmd/winqemu/main.go`: Seamless launcher supporting `--gui`, `--tui`, and desktop auto-detection.

---

### Task 1: Modern Webview Frontend Assets (`internal/gui/assets`)

**Files:**
- Create: `internal/gui/assets/index.html`
- Create: `internal/gui/assets/style.css`
- Create: `internal/gui/assets/app.js`

- [ ] **Step 1: Build semantic HTML structure with sidebar, top telemetry bar, and 7 interactive panels**
- [ ] **Step 2: Implement Dark Tech Cyberpunk Minimalist CSS styling (deep slate `#090d16`, glassmorphic cards, glowing borders, smooth animations)**
- [ ] **Step 3: Implement interactive JavaScript logic with Go IPC event hooks (tabs, inputs, download progress, presets, VM actions)**
- [ ] **Step 4: Commit**

---

### Task 2: Go Webview Window & Native IPC Bridge (`internal/gui`)

**Files:**
- Create: `internal/gui/app.go`
- Create: `internal/gui/bridge.go`
- Remove: `internal/gui/state.go`
- Remove: `internal/gui/views.go`

- [ ] **Step 1: Embed `assets/*` via `//go:embed` and configure local asset server/injection**
- [ ] **Step 2: Implement IPC bridge exposing `getHostInfo`, `getCatalog`, `startDownload`, `deployVM`, `launchVM`, and `installPackages`**
- [ ] **Step 3: Implement real-time download and deployment progress streaming to Webview (`w.Eval`)**
- [ ] **Step 4: Test compilation and verify IPC endpoints**
- [ ] **Step 5: Commit**

---

### Task 3: Dependency Cleanup & CLI Integration

**Files:**
- Modify: `go.mod`
- Modify: `cmd/winqemu/main.go`

- [ ] **Step 1: Run `go mod tidy` to clean up unneeded `go-gui` dependencies**
- [ ] **Step 2: Verify `cmd/winqemu/main.go` calls `gui.RunGUI()`**
- [ ] **Step 3: Build `bin/winqemu`**
- [ ] **Step 4: Commit**

---

### Task 4: End-to-End Verification, Documentation & Push

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Run all unit tests (`make test`)**
- [ ] **Step 2: Update README with modern Webview GUI details**
- [ ] **Step 3: Push commits to GitHub repository**
