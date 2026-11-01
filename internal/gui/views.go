package gui

import (
	"fmt"
	"strings"

	"github.com/go-gui-org/go-gui/gui"

	"winqemu/pkg/downloader"
	"winqemu/pkg/tuning"
	"winqemu/pkg/vm"
)

var (
	colorSidebarBG   = gui.ColorFromString("#14141e")
	colorContentBG   = gui.ColorFromString("#1a1a28")
	colorCardBG      = gui.ColorFromString("#222234")
	colorAccent      = gui.ColorFromString("#7d56f4")
	colorAccentHover = gui.ColorFromString("#916eff")
	colorSuccess     = gui.ColorFromString("#04b575")
	colorWarning     = gui.ColorFromString("#ffd700")
	colorDanger      = gui.ColorFromString("#ff4f68")
	colorTextLight   = gui.ColorFromString("#ffffff")
	colorTextMuted   = gui.ColorFromString("#a0a0b8")
	colorBorderSubtle= gui.ColorFromString("#2d2d44")
)

// MainView builds the top-level split view layout with sidebar and content.
func MainView(w *gui.Window) gui.View {
	return gui.Row(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Color:   colorContentBG,
		Padding: gui.NoPadding,
		Content: []gui.View{
			sidebarView(w),
			contentView(w),
		},
	})
}

func sidebarView(w *gui.Window) gui.View {
	state := gui.State[AppState](w)
	theme := gui.CurrentTheme()

	titleStyle := theme.TextStyleTitle
	titleStyle.Color = colorTextLight

	subStyle := theme.TextStyleCaption
	subStyle.Color = colorAccent

	navItems := []struct {
		tab   GuiTab
		label string
		icon  string
	}{
		{TabOverview, "Host Overview", "📊"},
		{TabOSSelection, "OS & Editions", "💿"},
		{TabMedia, "Installation Media", "📥"},
		{TabHardware, "Hardware Config", "⚙️"},
		{TabTuning, "Performance Tuning", "⚡"},
		{TabDeploy, "Build & Deploy", "🚀"},
		{TabExistingVMs, "VM Manager", "🗄️"},
	}

	var buttons []gui.View
	for _, item := range navItems {
		item := item
		isActive := state.ActiveTab == item.tab
		btnColor := colorSidebarBG
		textColor := colorTextMuted
		if isActive {
			btnColor = colorAccent
			textColor = colorTextLight
		}

		style := theme.TextStyleBody
		style.Color = textColor

		btn := gui.Button(gui.ButtonCfg{
			ID:       gui.ScopeID("nav", fmt.Sprintf("tab-%d", item.tab)),
			Sizing:   gui.FillFit,
			Colors:   gui.Flat(btnColor),
			Radius:   gui.RadiusSmall,
			Padding:  gui.PadVH(10, 14),
			MinWidth: 180,
			Content: []gui.View{
				gui.Text(gui.TextCfg{
					Text:      fmt.Sprintf("%s  %s", item.icon, item.label),
					TextStyle: style,
				}),
			},
			OnClick: func(ctx gui.EventCtx) {
				ctx.Consume()
				gui.State[AppState](ctx.Window).ActiveTab = item.tab
			},
		})
		buttons = append(buttons, btn)
	}

	// Host status indicator
	accelStatus := "KVM: Active"
	if !state.Host.KVMSupported {
		if state.Host.HVFSupported {
			accelStatus = "HVF: Active"
		} else if state.Host.WHPXSupported {
			accelStatus = "WHPX: Active"
		} else {
			accelStatus = "Accel: Disabled"
		}
	}

	statusStyle := theme.TextStyleCaption
	statusStyle.Color = colorSuccess

	footer := gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Padding: gui.PadAll(12),
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: fmt.Sprintf("OS: %s", state.Host.DistroPretty), TextStyle: subStyle}),
			gui.Text(gui.TextCfg{Text: accelStatus, TextStyle: statusStyle}),
		},
	})

	return gui.Column(gui.ContainerCfg{
		Sizing:      gui.FixedFill,
		Width:       220,
		Color:       colorSidebarBG,
		Padding:     gui.PadAll(12),
		Spacing:     gui.SpacingMedium,
		ColorBorder: colorBorderSubtle,
		SizeBorder:  gui.BorderPx(1),
		Content: []gui.View{
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.PadVH(8, 6),
				Content: []gui.View{
					gui.Text(gui.TextCfg{Text: "WinQEMU ⚡", TextStyle: titleStyle}),
					gui.Text(gui.TextCfg{Text: "Virtual Machine Studio", TextStyle: subStyle}),
				},
			}),
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SpacingSmall,
				Content: buttons,
			}),
			footer,
		},
	})
}

func contentView(w *gui.Window) gui.View {
	state := gui.State[AppState](w)

	var panel gui.View
	switch state.ActiveTab {
	case TabOverview:
		panel = viewOverview(w, state)
	case TabOSSelection:
		panel = viewOSSelection(w, state)
	case TabMedia:
		panel = viewMedia(w, state)
	case TabHardware:
		panel = viewHardware(w, state)
	case TabTuning:
		panel = viewTuning(w, state)
	case TabDeploy:
		panel = viewDeploy(w, state)
	case TabExistingVMs:
		panel = viewExistingVMs(w, state)
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Padding: gui.PadAll(24),
		Spacing: gui.SpacingLarge,
		Content: []gui.View{panel},
	})
}

// 1. Host Overview
func viewOverview(w *gui.Window, state *AppState) gui.View {
	theme := gui.CurrentTheme()
	hStyle := theme.TextStyleDisplay
	hStyle.Color = colorTextLight

	pStyle := theme.TextStyleBody
	pStyle.Color = colorTextMuted

	var depRows []gui.View
	allInstalled := true
	for _, dep := range state.Host.Dependencies {
		statusText := "✔ Ready"
		statusColor := colorSuccess
		if !dep.Installed {
			allInstalled = false
			if dep.Required {
				statusText = "✖ Missing (Required)"
				statusColor = colorDanger
			} else {
				statusText = "⚠ Optional"
				statusColor = colorWarning
			}
		}

		sStyle := theme.TextStyleCaption
		sStyle.Color = statusColor

		row := gui.Row(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Padding: gui.PadVH(6, 0),
			VAlign:  gui.VAlignMiddle,
			Content: []gui.View{
				gui.Column(gui.ContainerCfg{Width: 180, Sizing: gui.FixedFit, Content: []gui.View{
					gui.Text(gui.TextCfg{Text: dep.Name, TextStyle: pStyle}),
				}}),
				gui.Column(gui.ContainerCfg{Width: 220, Sizing: gui.FixedFit, Content: []gui.View{
					gui.Text(gui.TextCfg{Text: dep.Description, TextStyle: pStyle}),
				}}),
				gui.Text(gui.TextCfg{Text: statusText, TextStyle: sStyle}),
			},
		})
		depRows = append(depRows, row)
	}

	var installAction gui.View
	if !allInstalled {
		installAction = gui.Row(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Spacing: gui.SpacingMedium,
			Content: []gui.View{
				gui.Button(gui.ButtonCfg{
					ID:      "install-deps-btn",
					Colors:  gui.Flat(colorAccent),
					Radius:  gui.RadiusSmall,
					Padding: gui.PadVH(10, 16),
					Content: []gui.View{
						gui.Text(gui.TextCfg{Text: "📦 Install Missing Packages (Interactive)", TextStyle: theme.TextStyleTitle}),
					},
					OnClick: func(ctx gui.EventCtx) {
						ctx.Consume()
						_ = LaunchTerminalInstaller(state.Host)
					},
				}),
			},
		})
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingMedium,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "System Telemetry & Virtualization Health", TextStyle: hStyle}),
			gui.Text(gui.TextCfg{
				Text:      fmt.Sprintf("Host: %s • %d CPUs • %d MB RAM • Architecture: %s", state.Host.DistroPretty, state.Host.LogicalCPUs, state.Host.TotalRAMMB, state.Host.Arch),
				TextStyle: pStyle,
			}),
			gui.Column(gui.ContainerCfg{
				Sizing:      gui.FillFit,
				Color:       colorCardBG,
				Padding:     gui.PadAll(16),
				Radius:      gui.RadiusMedium,
				SizeBorder:  gui.BorderPx(1),
				ColorBorder: colorBorderSubtle,
				Content:     depRows,
			}),
			installAction,
		},
	})
}

// 2. OS & Editions Selection
func viewOSSelection(w *gui.Window, state *AppState) gui.View {
	theme := gui.CurrentTheme()
	hStyle := theme.TextStyleDisplay
	hStyle.Color = colorTextLight

	pStyle := theme.TextStyleBody
	pStyle.Color = colorTextMuted

	// Family Switcher Buttons
	var famButtons []gui.View
	for _, fam := range state.Catalog.Families {
		fam := fam
		isSelected := state.SelectedFamily != nil && state.SelectedFamily.ID == fam.ID
		bgColor := colorCardBG
		txtColor := colorTextMuted
		if isSelected {
			bgColor = colorAccent
			txtColor = colorTextLight
		}
		btnStyle := theme.TextStyleBody
		btnStyle.Color = txtColor

		btn := gui.Button(gui.ButtonCfg{
			ID:      gui.ScopeID("fam", fam.ID),
			Colors:  gui.Flat(bgColor),
			Radius:  gui.RadiusSmall,
			Padding: gui.PadVH(8, 14),
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: fam.DisplayName, TextStyle: btnStyle}),
			},
			OnClick: func(ctx gui.EventCtx) {
				ctx.Consume()
				st := gui.State[AppState](ctx.Window)
				st.SelectedFamily = &fam
				if len(fam.Editions) > 0 {
					st.SelectedEdition = &fam.Editions[0]
					st.VMName = strings.ReplaceAll(fam.Editions[0].Name, " ", "-")
				}
			},
		})
		famButtons = append(famButtons, btn)
	}

	// Editions List
	var editionCards []gui.View
	if state.SelectedFamily != nil {
		for _, ed := range state.SelectedFamily.Editions {
			ed := ed
			isSelected := state.SelectedEdition != nil && state.SelectedEdition.ID == ed.ID
			cardColor := colorCardBG
			if isSelected {
				cardColor = colorBorderSubtle
			}

			titleStyle := theme.TextStyleTitle
			titleStyle.Color = colorTextLight

			recBadgeStyle := theme.TextStyleCaption
			recBadgeStyle.Color = colorWarning

			cached, _, _ := downloader.IsISOCached(ed.ID, ed.Architecture, "")
			cachedBadge := ""
			if cached {
				cachedBadge = " [Cached - Ready]"
			}

			var badgeViews []gui.View
			badgeViews = append(badgeViews, gui.Text(gui.TextCfg{Text: ed.Name + cachedBadge, TextStyle: titleStyle}))
			if ed.IsRecommended {
				badgeViews = append(badgeViews, gui.Text(gui.TextCfg{Text: "⭐ Recommended by WispyDocs", TextStyle: recBadgeStyle}))
			}

			cardContent := []gui.View{
				gui.Row(gui.ContainerCfg{
					Sizing:  gui.FillFit,
					Spacing: gui.SpacingMedium,
					Content: badgeViews,
				}),
				gui.Text(gui.TextCfg{Text: ed.Recommendation, TextStyle: pStyle}),
			}

			btn := gui.Button(gui.ButtonCfg{
				ID:         gui.ScopeID("ed", ed.ID),
				Sizing:     gui.FillFit,
				Colors:     gui.ColorSet{Base: cardColor, Border: colorBorderSubtle},
				Radius:     gui.RadiusMedium,
				Padding:    gui.PadAll(12),
				SizeBorder: gui.BorderPx(1),
				Content:    cardContent,
				OnClick: func(ctx gui.EventCtx) {
					ctx.Consume()
					st := gui.State[AppState](ctx.Window)
					st.SelectedEdition = &ed
					st.VMName = strings.ReplaceAll(ed.Name, " ", "-")
				},
			})
			editionCards = append(editionCards, btn)
		}
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingMedium,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Choose Windows Generation & Edition", TextStyle: hStyle}),
			gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SpacingSmall,
				Content: famButtons,
			}),
			gui.Column(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Spacing: gui.SpacingSmall,
				Content: editionCards,
			}),
		},
	})
}

// 3. Media & Downloader
func viewMedia(w *gui.Window, state *AppState) gui.View {
	theme := gui.CurrentTheme()
	hStyle := theme.TextStyleDisplay
	hStyle.Color = colorTextLight

	pStyle := theme.TextStyleBody
	pStyle.Color = colorTextMuted

	if state.SelectedEdition == nil {
		return gui.Text(gui.TextCfg{Text: "Please choose an edition first.", TextStyle: pStyle})
	}

	cached, cachePath, sizeBytes := downloader.IsISOCached(state.SelectedEdition.ID, state.SelectedEdition.Architecture, "")

	var downloadWidget gui.View
	if state.IsDownloading {
		downloadWidget = gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Spacing: gui.SpacingSmall,
			Content: []gui.View{
				gui.Text(gui.TextCfg{
					Text:      fmt.Sprintf("Downloading... %.1f%% (%s)", state.DownloadProgress*100.0, state.DownloadSpeedStr),
					TextStyle: theme.TextStyleTitle,
				}),
			},
		})
	} else if cached {
		downloadWidget = gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Spacing: gui.SpacingSmall,
			Content: []gui.View{
				gui.Text(gui.TextCfg{
					Text:      fmt.Sprintf("✔ Genuine ISO is already cached on disk (%.2f GB)!", float64(sizeBytes)/(1024*1024*1024)),
					TextStyle: theme.TextStyleTitle,
				}),
				gui.Text(gui.TextCfg{Text: cachePath, TextStyle: pStyle}),
			},
		})
	} else {
		downloadWidget = gui.Button(gui.ButtonCfg{
			ID:      "start-download-btn",
			Colors:  gui.Flat(colorAccent),
			Radius:  gui.RadiusSmall,
			Padding: gui.PadVH(10, 18),
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: "📥 Download Genuine ISO from Microsoft CDN", TextStyle: theme.TextStyleTitle}),
			},
			OnClick: func(ctx gui.EventCtx) {
				ctx.Consume()
				state.StartDownload(func() {
					// re-render callback
				})
			},
		})
	}

	// Local ISO toggle & input
	localInput := gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFit,
		Spacing: gui.SpacingSmall,
		Content: []gui.View{
			gui.Checkbox(gui.ToggleCfg{
				Label:    "Use a local Windows ISO file instead of downloading",
				Selected: state.UseLocalISO,
				OnClick: func(ctx gui.EventCtx) {
					ctx.Consume()
					st := gui.State[AppState](ctx.Window)
					st.UseLocalISO = !st.UseLocalISO
				},
			}),
			gui.Input(gui.InputCfg{
				Label:       "Local ISO Path:",
				Text:        state.LocalISOPath,
				Placeholder: "/path/to/windows.iso",
				Sizing:      gui.FillFit,
				OnTextChanged: func(val string, ctx gui.EventCtx) {
					gui.State[AppState](ctx.Window).LocalISOPath = val
				},
			}),
		},
	})

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingMedium,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Windows Installation Media", TextStyle: hStyle}),
			gui.Text(gui.TextCfg{Text: fmt.Sprintf("Target: %s (%s)", state.SelectedEdition.Name, state.SelectedEdition.Architecture), TextStyle: pStyle}),
			gui.Column(gui.ContainerCfg{
				Sizing:      gui.FillFit,
				Color:       colorCardBG,
				Padding:     gui.PadAll(18),
				Radius:      gui.RadiusMedium,
				SizeBorder:  gui.BorderPx(1),
				ColorBorder: colorBorderSubtle,
				Content: []gui.View{
					downloadWidget,
					localInput,
				},
			}),
		},
	})
}

// 4. Hardware Resources
func viewHardware(w *gui.Window, state *AppState) gui.View {
	theme := gui.CurrentTheme()
	hStyle := theme.TextStyleDisplay
	hStyle.Color = colorTextLight

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingMedium,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Hardware Resource Allocation", TextStyle: hStyle}),
			gui.Column(gui.ContainerCfg{
				Sizing:      gui.FillFit,
				Color:       colorCardBG,
				Padding:     gui.PadAll(18),
				Radius:      gui.RadiusMedium,
				Spacing:     gui.SpacingMedium,
				SizeBorder:  gui.BorderPx(1),
				ColorBorder: colorBorderSubtle,
				Content: []gui.View{
					gui.Input(gui.InputCfg{
						Label: "VM Name:",
						Text:  state.VMName,
						OnTextChanged: func(val string, ctx gui.EventCtx) {
							gui.State[AppState](ctx.Window).VMName = val
						},
					}),
					gui.Input(gui.InputCfg{
						Label: fmt.Sprintf("RAM Memory (MB) [Host: %d MB]:", state.Host.TotalRAMMB),
						Text:  state.RAMMB,
						OnTextChanged: func(val string, ctx gui.EventCtx) {
							gui.State[AppState](ctx.Window).RAMMB = val
						},
					}),
					gui.Input(gui.InputCfg{
						Label: fmt.Sprintf("CPU Cores [Host: %d CPUs]:", state.Host.LogicalCPUs),
						Text:  state.CPUCores,
						OnTextChanged: func(val string, ctx gui.EventCtx) {
							gui.State[AppState](ctx.Window).CPUCores = val
						},
					}),
					gui.Input(gui.InputCfg{
						Label: "Virtual Disk Size (GB) [Sparse qcow2]:",
						Text:  state.DiskGB,
						OnTextChanged: func(val string, ctx gui.EventCtx) {
							gui.State[AppState](ctx.Window).DiskGB = val
						},
					}),
					gui.Input(gui.InputCfg{
						Label: "Custom Storage Directory:",
						Text:  state.CustomStorageDir,
						OnTextChanged: func(val string, ctx gui.EventCtx) {
							gui.State[AppState](ctx.Window).CustomStorageDir = val
						},
					}),
				},
			}),
		},
	})
}

// 5. Performance & Tuning
func viewTuning(w *gui.Window, state *AppState) gui.View {
	theme := gui.CurrentTheme()
	hStyle := theme.TextStyleDisplay
	hStyle.Color = colorTextLight

	isGaming := state.TuningConfig.Profile == tuning.ProfileGaming
	profBtnText := "Profile: Balanced (Standard)"
	if isGaming {
		profBtnText = "Profile: ⚡ Maximum Gaming & Low-Latency"
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingMedium,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Performance Optimization & QoL Presets", TextStyle: hStyle}),
			gui.Column(gui.ContainerCfg{
				Sizing:      gui.FillFit,
				Color:       colorCardBG,
				Padding:     gui.PadAll(18),
				Radius:      gui.RadiusMedium,
				Spacing:     gui.SpacingMedium,
				SizeBorder:  gui.BorderPx(1),
				ColorBorder: colorBorderSubtle,
				Content: []gui.View{
					gui.Button(gui.ButtonCfg{
						Colors:  gui.Flat(colorAccent),
						Radius:  gui.RadiusSmall,
						Padding: gui.PadVH(8, 14),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: profBtnText, TextStyle: theme.TextStyleTitle}),
						},
						OnClick: func(ctx gui.EventCtx) {
							ctx.Consume()
							st := gui.State[AppState](ctx.Window)
							isWin11 := st.SelectedEdition != nil && st.SelectedEdition.IsWin11
							if st.TuningConfig.Profile == tuning.ProfileGaming {
								st.TuningConfig = tuning.GetPreset(tuning.ProfileBalanced, isWin11)
							} else {
								st.TuningConfig = tuning.GetPreset(tuning.ProfileGaming, isWin11)
							}
						},
					}),
					gui.Checkbox(gui.ToggleCfg{
						Label:    "Bypass Windows 11 TPM 2.0, Secure Boot, & MS Account OOBE",
						Selected: state.TuningConfig.BypassTPMAndOOBE,
						OnClick: func(ctx gui.EventCtx) {
							ctx.Consume()
							st := gui.State[AppState](ctx.Window)
							st.TuningConfig.BypassTPMAndOOBE = !st.TuningConfig.BypassTPMAndOOBE
						},
					}),
					gui.Checkbox(gui.ToggleCfg{
						Label:    "Looking Glass IVSHMEM Support (Ultra Low-Latency Display Device)",
						Selected: state.TuningConfig.LookingGlassEnabled,
						OnClick: func(ctx gui.EventCtx) {
							ctx.Consume()
							st := gui.State[AppState](ctx.Window)
							st.TuningConfig.LookingGlassEnabled = !st.TuningConfig.LookingGlassEnabled
						},
					}),
					gui.Checkbox(gui.ToggleCfg{
						Label:    "VirtIO SCSI Storage & 10G VirtIO Network Drivers",
						Selected: state.TuningConfig.VirtIOSCSI,
						OnClick: func(ctx gui.EventCtx) {
							ctx.Consume()
							st := gui.State[AppState](ctx.Window)
							st.TuningConfig.VirtIOSCSI = !st.TuningConfig.VirtIOSCSI
						},
					}),
					gui.Checkbox(gui.ToggleCfg{
						Label:    "Run Microsoft Activation Scripts (MAS) on First Logon",
						Selected: state.TuningConfig.MASActivationPrompt,
						OnClick: func(ctx gui.EventCtx) {
							ctx.Consume()
							st := gui.State[AppState](ctx.Window)
							st.TuningConfig.MASActivationPrompt = !st.TuningConfig.MASActivationPrompt
						},
					}),
				},
			}),
		},
	})
}

// 6. Build & Deploy
func viewDeploy(w *gui.Window, state *AppState) gui.View {
	theme := gui.CurrentTheme()
	hStyle := theme.TextStyleDisplay
	hStyle.Color = colorTextLight

	pStyle := theme.TextStyleBody
	pStyle.Color = colorTextMuted

	var deployWidget gui.View
	if state.IsDeploying {
		deployWidget = gui.Text(gui.TextCfg{Text: "Deploying VM artifacts (qcow2, autounattend, Libvirt)...", TextStyle: theme.TextStyleTitle})
	} else if state.DeployResult != nil {
		deployWidget = gui.Column(gui.ContainerCfg{
			Sizing:  gui.FillFit,
			Spacing: gui.SpacingSmall,
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: fmt.Sprintf("🎉 VM '%s' Successfully Created!", state.DeployResult.VMName), TextStyle: theme.TextStyleTitle}),
				gui.Text(gui.TextCfg{Text: fmt.Sprintf("Virtual Disk: %s", state.DeployResult.DiskPath), TextStyle: pStyle}),
				gui.Text(gui.TextCfg{Text: fmt.Sprintf("Driver Floppy: %s", state.DeployResult.DriverDiskPath), TextStyle: pStyle}),
				gui.Button(gui.ButtonCfg{
					ID:      "launch-vm-now-btn",
					Colors:  gui.Flat(colorSuccess),
					Radius:  gui.RadiusSmall,
					Padding: gui.PadVH(10, 18),
					Content: []gui.View{
						gui.Text(gui.TextCfg{Text: "▶ Launch VM in virt-manager / QEMU Now", TextStyle: theme.TextStyleTitle}),
					},
					OnClick: func(ctx gui.EventCtx) {
						ctx.Consume()
						_ = vm.LaunchVirtManager(state.DeployResult.VMName)
					},
				}),
			},
		})
	} else {
		deployWidget = gui.Button(gui.ButtonCfg{
			ID:      "deploy-now-btn",
			Colors:  gui.Flat(colorSuccess),
			Radius:  gui.RadiusSmall,
			Padding: gui.PadVH(12, 24),
			Content: []gui.View{
				gui.Text(gui.TextCfg{Text: "🚀 Create & Register Virtual Machine", TextStyle: theme.TextStyleTitle}),
			},
			OnClick: func(ctx gui.EventCtx) {
				ctx.Consume()
				state.DeployVM(func() {
					// re-render callback
				})
			},
		})
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingMedium,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Deploy Virtual Machine", TextStyle: hStyle}),
			gui.Column(gui.ContainerCfg{
				Sizing:      gui.FillFit,
				Color:       colorCardBG,
				Padding:     gui.PadAll(18),
				Radius:      gui.RadiusMedium,
				SizeBorder:  gui.BorderPx(1),
				ColorBorder: colorBorderSubtle,
				Content:     []gui.View{deployWidget},
			}),
		},
	})
}

// 7. Existing VMs Manager
func viewExistingVMs(w *gui.Window, state *AppState) gui.View {
	theme := gui.CurrentTheme()
	hStyle := theme.TextStyleDisplay
	hStyle.Color = colorTextLight

	pStyle := theme.TextStyleBody
	pStyle.Color = colorTextMuted

	var vmRows []gui.View
	if len(state.ExistingVMs) == 0 {
		vmRows = append(vmRows, gui.Text(gui.TextCfg{
			Text:      "No WinQEMU virtual machines found in this storage directory.",
			TextStyle: pStyle,
		}))
	} else {
		for _, v := range state.ExistingVMs {
			v := v
			row := gui.Row(gui.ContainerCfg{
				Sizing:  gui.FillFit,
				Padding: gui.PadVH(8, 0),
				VAlign:  gui.VAlignMiddle,
				Content: []gui.View{
					gui.Column(gui.ContainerCfg{Width: 200, Sizing: gui.FixedFit, Content: []gui.View{
						gui.Text(gui.TextCfg{Text: v.Name, TextStyle: theme.TextStyleTitle}),
					}}),
					gui.Column(gui.ContainerCfg{Width: 120, Sizing: gui.FixedFit, Content: []gui.View{
						gui.Text(gui.TextCfg{Text: fmt.Sprintf("%d MB Disk", v.DiskSizeMB), TextStyle: pStyle}),
					}}),
					gui.Button(gui.ButtonCfg{
						ID:      gui.ScopeID("start", v.Name),
						Colors:  gui.Flat(colorAccent),
						Radius:  gui.RadiusSmall,
						Padding: gui.PadVH(6, 12),
						Content: []gui.View{
							gui.Text(gui.TextCfg{Text: "▶ Start VM", TextStyle: theme.TextStyleBody}),
						},
						OnClick: func(ctx gui.EventCtx) {
							ctx.Consume()
							_ = vm.LaunchExistingVM(v)
						},
					}),
				},
			})
			vmRows = append(vmRows, row)
		}
	}

	return gui.Column(gui.ContainerCfg{
		Sizing:  gui.FillFill,
		Spacing: gui.SpacingMedium,
		Content: []gui.View{
			gui.Text(gui.TextCfg{Text: "Existing Virtual Machines", TextStyle: hStyle}),
			gui.Column(gui.ContainerCfg{
				Sizing:      gui.FillFit,
				Color:       colorCardBG,
				Padding:     gui.PadAll(18),
				Radius:      gui.RadiusMedium,
				SizeBorder:  gui.BorderPx(1),
				ColorBorder: colorBorderSubtle,
				Content:     vmRows,
			}),
		},
	})
}
