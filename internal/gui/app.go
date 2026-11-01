package gui

import (
	"github.com/go-gui-org/go-gui/gui"
	"github.com/go-gui-org/go-gui/gui/backend"
)

// RunGUI initializes and launches the desktop GUI window using Go-Gui.
func RunGUI() error {
	state := NewAppState()
	gui.SetTheme(gui.ThemeDark.WithBorders(true))

	w := gui.NewWindow(gui.WindowCfg{
		State:  state,
		Title:  "WinQEMU ⚡ Virtual Machine Studio",
		Width:  980,
		Height: 680,
		OnInit: func(w *gui.Window) {
			w.SetView(MainView)
		},
	})

	backend.Run(w)
	return nil
}
