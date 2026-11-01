package gui

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"winqemu/pkg/sysinfo"
)

// LaunchTerminalInstaller spawns an interactive terminal window to execute the dependency installer.
func LaunchTerminalInstaller(info *sysinfo.HostInfo) error {
	baseCmd := sysinfo.GetInstallCommand(info)

	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd.exe", "/c", "start", "powershell", "-NoProfile", "-Command",
			fmt.Sprintf(`Write-Host "==> WinQEMU Package Installer" -ForegroundColor Cyan; %s; Read-Host "Press [Enter] to exit..."`, baseCmd))
		return cmd.Start()

	case "darwin":
		appleScript := fmt.Sprintf(`tell application "Terminal" to do script "%s; echo ''; read -p 'Press [Enter] to exit...'"`, baseCmd)
		cmd := exec.Command("osascript", "-e", appleScript)
		return cmd.Start()

	default: // Linux
		script := fmt.Sprintf(`echo -e "\n\033[1;35m==> WinQEMU Package Installer\033[0m\n"; %s; echo -e "\n\033[1;32m==> Done! Press [Enter] to return to WinQEMU...\033[0m"; read -r`, baseCmd)

		terminals := []struct {
			bin  string
			args []string
		}{
			{"alacritty", []string{"-e", "bash", "-c", script}},
			{"kitty", []string{"bash", "-c", script}},
			{"ghostty", []string{"-e", "bash", "-c", script}},
			{"gnome-terminal", []string{"--", "bash", "-c", script}},
			{"konsole", []string{"-e", "bash", "-c", script}},
			{"xfce4-terminal", []string{"-e", fmt.Sprintf("bash -c '%s'", script)}},
			{"xterm", []string{"-e", "bash", "-c", script}},
		}

		for _, t := range terminals {
			if path, err := exec.LookPath(t.bin); err == nil {
				cmd := exec.Command(path, t.args...)
				return cmd.Start()
			}
		}

		// Fallback to pkexec if no desktop terminal emulator is found
		if path, err := exec.LookPath("pkexec"); err == nil {
			cmd := exec.Command(path, "bash", "-c", baseCmd)
			return cmd.Start()
		}

		// Direct shell fallback
		cmd := exec.Command("bash", "-c", baseCmd)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Start()
	}
}
