package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"winqemu/internal/gui"
	"winqemu/internal/tui"
	"winqemu/pkg/msdl"
)

const version = "1.0.0"

func main() {
	// 1. Check for `winqemu msdl ...` subcommand or `--msdl` flag
	if len(os.Args) > 1 && (os.Args[1] == "msdl" || os.Args[1] == "--msdl") {
		if err := msdl.RunCLI(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	guiFlag := flag.Bool("gui", false, "Force GUI mode (desktop window)")
	tuiFlag := flag.Bool("tui", false, "Force TUI mode (terminal wizard)")
	verFlag := flag.Bool("version", false, "Print version information and exit")
	vShortFlag := flag.Bool("v", false, "Print version information and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "WinQEMU ⚡ Next-Gen Windows VM Creator on QEMU / virt-manager\n\n")
		fmt.Fprintf(os.Stderr, "Usage: winqemu [options] [command]\n\n")
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  msdl [flags]   Direct Microsoft ISO downloader CLI (msdl.tech-latest.com)\n")
		fmt.Fprintf(os.Stderr, "                 Examples:\n")
		fmt.Fprintf(os.Stderr, "                   winqemu msdl --list\n")
		fmt.Fprintf(os.Stderr, "                   winqemu msdl --id 3813 --lang \"English\"\n")
		fmt.Fprintf(os.Stderr, "                   winqemu msdl --eval win11-ent\n")
		fmt.Fprintf(os.Stderr, "                   winqemu msdl --eval server-2025\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  --gui          Launch desktop graphical interface (Native Webview)\n")
		fmt.Fprintf(os.Stderr, "  --tui          Launch terminal user interface (Bubble Tea)\n")
		fmt.Fprintf(os.Stderr, "  -v, --version  Print version information and exit\n")
		fmt.Fprintf(os.Stderr, "  -h, --help     Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "By default, WinQEMU auto-detects if you are in a graphical desktop environment\n")
		fmt.Fprintf(os.Stderr, "(X11/Wayland/macOS/Windows) and launches the GUI, falling back to TUI if headless.\n")
	}

	flag.Parse()

	if *verFlag || *vShortFlag {
		fmt.Printf("WinQEMU v%s (%s/%s)\n", version, runtime.GOOS, runtime.GOARCH)
		return
	}

	// Decision: GUI or TUI
	shouldRunGUI := false

	if *guiFlag {
		shouldRunGUI = true
	} else if *tuiFlag {
		shouldRunGUI = false
	} else {
		// Auto-detect graphical environment
		shouldRunGUI = isGraphicalSession()
	}

	if shouldRunGUI {
		if err := runGUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Notice: GUI failed to start (%v). Falling back to TUI mode...\n", err)
			runTUI()
		}
		return
	}

	runTUI()
}

func isGraphicalSession() bool {
	switch runtime.GOOS {
	case "linux":
		// Check for X11 or Wayland display server environment variables
		return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	case "darwin", "windows":
		// Desktop OS default
		return true
	default:
		return false
	}
}

func runGUI() error {
	return gui.RunGUI()
}

func runTUI() {
	app, err := tui.NewAppModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing WinQEMU TUI: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	app.SetProgram(p)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI program: %v\n", err)
		os.Exit(1)
	}
}
