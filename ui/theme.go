package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// Palette — change colors here to restyle the whole app.
var (
	colorAccent = lipgloss.Color("205") // pink/magenta highlight
	colorMuted  = lipgloss.Color("240") // dim gray
	colorBorder = lipgloss.Color("63")  // panel border
)

// Shared style definitions. These are the single source of truth for the
// look of the app; every render site draws from them.
var (
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(colorAccent).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Faint(true).
				Padding(0, 2)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(1, 2)

	selectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorAccent)

	rowStyle = lipgloss.NewStyle()

	helpStyle = lipgloss.NewStyle().Faint(true)
)

// TabLabel renders a single tab label, styled by whether it is active.
func TabLabel(label string, active bool) string {
	if active {
		return activeTabStyle.Render(label)
	}
	return inactiveTabStyle.Render(label)
}

// TabBar joins tab labels into a single bar.
func TabBar(labels []string, active int) string {
	rendered := make([]string, len(labels))
	for i, label := range labels {
		rendered[i] = TabLabel(label, i == active)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// Row renders a menu row, highlighted when selected.
func Row(text string, selected bool) string {
	if selected {
		return selectedRowStyle.Render("› " + text)
	}
	return rowStyle.Render("  " + text)
}

// Panel wraps content in a bordered box sized to the given width/height.
// When width/height are <= 0 the panel sizes to its content.
func Panel(content string, width, height int) string {
	s := panelStyle
	if width > 0 {
		// Account for the border (2) and horizontal padding (4).
		s = s.Width(width - 2)
	}
	if height > 0 {
		// Account for the border (2) and vertical padding (2).
		s = s.Height(height - 2)
	}
	return s.Render(content)
}

// Help renders footer help text in a muted style.
func Help(text string) string {
	return helpStyle.Render(text)
}

// Center places content in the middle of the given box.
func Center(width, height int, content string) string {
	if width <= 0 || height <= 0 {
		return content
	}
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

// muted is a small helper for dimming arbitrary text.
func muted(text string) string {
	return lipgloss.NewStyle().Foreground(colorMuted).Render(strings.TrimRight(text, "\n"))
}
