// Copyright 2026 Kartoza
// SPDX-License-Identifier: MIT

package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Kartoza Brand Colors - matching the official Kartoza website
// See: https://kartoza.com
var (
	// Primary Kartoza blue palette
	KartozaBlue      = lipgloss.Color("#3B9DD9") // Primary blue
	KartozaBlueDark  = lipgloss.Color("#1B6B9B") // Dark blue
	KartozaBlueLight = lipgloss.Color("#5BB5E8") // Light blue

	// Accent orange/gold palette
	KartozaOrange      = lipgloss.Color("#E8A331") // Primary orange
	KartozaOrangeDark  = lipgloss.Color("#D4922A") // Dark orange
	KartozaOrangeLight = lipgloss.Color("#F0B84D") // Light orange

	// Semantic colors using Kartoza palette
	Primary   = KartozaBlue
	Secondary = KartozaOrange
	Accent    = KartozaOrangeLight
	Danger    = lipgloss.Color("#E55B3C") // Kartoza-style red

	// Neutral colors - dark theme
	Background  = lipgloss.Color("#1a2a3a") // Kartoza dark
	Surface     = lipgloss.Color("#2a3a4a") // Slightly lighter
	SurfaceHigh = lipgloss.Color("#3d4f5f") // Even lighter for active elements
	Border      = lipgloss.Color("#4D6370") // Kartoza text-muted
	Muted       = lipgloss.Color("#6B7B8D") // Muted text
	Text        = lipgloss.Color("#e8ecf0") // Light text
	TextBright  = lipgloss.Color("#FFFFFF") // Bright text

	// Special colors using Kartoza palette
	Selected   = KartozaBlueLight          // Selected item
	Directory  = KartozaBlue               // Directory color
	Executable = lipgloss.Color("#4CAF50") // Success green
	GeoFile    = KartozaOrangeLight        // Geospatial files
	StyleFile  = lipgloss.Color("#9f7aea") // Style files (SLD/CSS)

	// Status colors
	Success = lipgloss.Color("#4CAF50")
	Warning = KartozaOrange
	Error   = Danger
	Info    = KartozaBlueLight
)

// App-wide styles
var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Background(Background).
			Foreground(Text)

	// Title bar - Kartoza branded
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(TextBright).
			Background(KartozaBlueDark).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	// Muted text style
	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Panel styles
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	ActivePanelStyle = PanelStyle.
				BorderForeground(Primary)

	PanelHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(KartozaBlue).
				Padding(0, 1).
				MarginBottom(1)

	// List item styles
	ItemStyle = lipgloss.NewStyle().
			Foreground(Text).
			Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(TextBright).
				Background(SurfaceHigh).
				Bold(true).
				Padding(0, 1)

	ActiveItemStyle = lipgloss.NewStyle().
			Foreground(TextBright).
			Background(KartozaBlue).
			Bold(true).
			Padding(0, 1)

	// Directory style
	DirectoryStyle = lipgloss.NewStyle().
			Foreground(Directory).
			Bold(true)

	// File type styles
	GeoFileStyle = lipgloss.NewStyle().
			Foreground(GeoFile)

	StyleFileStyle = lipgloss.NewStyle().
			Foreground(StyleFile)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(Surface).
			Padding(0, 1)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(TextBright).
			Background(KartozaBlueDark).
			Padding(0, 1)

	StatusValueStyle = lipgloss.NewStyle().
				Foreground(Text).
				Background(Surface).
				Padding(0, 1)

	// Help bar
	HelpBarStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Background(Surface).
			Padding(0, 1)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(Selected).
			Bold(true)

	HelpTextStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Error styles
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Bold(true)

	ErrorMsgStyle = lipgloss.NewStyle().
			Foreground(Danger).
			Background(Surface).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Danger)

	// Success styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	// Loading styles
	LoadingStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Italic(true)

	// Dialog styles - using rounded borders for consistency
	DialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(KartozaBlue).
			Background(Surface).
			Padding(1, 2)

	DialogTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(KartozaBlue).
				MarginBottom(1)

	// Input styles - using rounded borders
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(KartozaBlue).
				Padding(0, 1)

	// Button styles - using rounded borders
	ButtonStyle = lipgloss.NewStyle().
			Foreground(Text).
			Background(Surface).
			Padding(0, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)

	FocusedButtonStyle = lipgloss.NewStyle().
				Foreground(TextBright).
				Background(KartozaBlue).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(KartozaBlue)

	// Tree styles
	TreeBranchStyle = lipgloss.NewStyle().
			Foreground(Muted)

	ExpandedNodeStyle = lipgloss.NewStyle().
				Foreground(Secondary)

	CollapsedNodeStyle = lipgloss.NewStyle().
				Foreground(Muted)

	// Connection status
	ConnectedStyle = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	DisconnectedStyle = lipgloss.NewStyle().
				Foreground(Danger)

	// Progress styles - Kartoza branded
	ProgressBarStyle = lipgloss.NewStyle().
				Foreground(KartozaBlue)

	ProgressTextStyle = lipgloss.NewStyle().
				Foreground(Text).
				Italic(true)

	// Count badge style for tree view item counts
	CountBadgeStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	// Dialog box style with background overlay effect - using rounded borders
	DialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(KartozaBlue).
			Background(Surface).
			Padding(1, 2)

	// Dialog label style
	DialogLabelStyle = lipgloss.NewStyle().
				Foreground(Text).
				Bold(true)

	// Dialog selected label style
	DialogSelectedLabelStyle = lipgloss.NewStyle().
					Foreground(Selected).
					Bold(true)

	// Dialog option style
	DialogOptionStyle = lipgloss.NewStyle().
				Foreground(Text)

	// Dialog selected option style
	DialogSelectedOptionStyle = lipgloss.NewStyle().
					Foreground(TextBright).
					Background(SurfaceHigh).
					Bold(true)

	// Dialog description style
	DialogDescStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	// Dialog help style
	DialogHelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Align(lipgloss.Center)

	// Accent text style
	AccentStyle = lipgloss.NewStyle().
			Foreground(Accent).
			Bold(true)

	// Input field styles for forms - using rounded borders
	InputSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Selected).
				Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(KartozaBlue).
				Padding(0, 1)

	// Textarea styles - using rounded borders
	TextAreaStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)

	TextAreaSelectedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(Selected)

	TextAreaFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(KartozaBlue)
)

// Helper functions for building complex layouts
func JoinHorizontal(items ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, items...)
}

func JoinVertical(items ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, items...)
}

func Center(width, height int, content string) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

// RenderHelpKey renders a key-value help item
func RenderHelpKey(key, desc string) string {
	return HelpKeyStyle.Render(key) + " " + HelpTextStyle.Render(desc)
}

// RenderStatusItem renders a status bar item
func RenderStatusItem(key, value string) string {
	return StatusKeyStyle.Render(key) + StatusValueStyle.Render(value)
}
