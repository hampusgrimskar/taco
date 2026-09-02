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

// activeIndicatorStyle marks repos with a running session.
var activeIndicatorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("42")) // green

// RepoRow renders a repo menu row with a cursor marker on the left and an
// active-session indicator aligned to indicatorCol on the right. indicatorCol
// is the column (measured from the start of the alias) where the indicator
// is drawn, so all indicators line up regardless of alias length.
func RepoRow(alias string, indicatorCol int, active, selected bool) string {
	cursor := "  "
	if selected {
		cursor = "› "
	}

	// Pad the alias out to the indicator column so the indicator aligns.
	pad := indicatorCol - len(alias)
	if pad < 1 {
		pad = 1
	}
	padding := strings.Repeat(" ", pad)

	indicator := " "
	if active {
		indicator = activeIndicatorStyle.Render("●")
	}

	line := cursor + alias + padding + indicator
	if selected {
		return selectedRowStyle.Render(line)
	}
	return rowStyle.Render(line)
}

// Panel border + padding costs, used to compute the inner content area.
const (
	panelBorderW  = 2 // left + right border
	panelBorderH  = 2 // top + bottom border
	panelPaddingW = 4 // left + right padding (2 each)
	panelPaddingH = 2 // top + bottom padding (1 each)
)

// panelInner returns the content area (width, height) available inside a panel
// of the given outer width/height, after border and padding.
func panelInner(width, height int) (int, int) {
	return width - panelBorderW - panelPaddingW, height - panelBorderH - panelPaddingH
}

// Panel wraps content in a bordered box sized to the given width/height.
// When width/height are <= 0 the panel sizes to its content.
// Note: lipgloss .Width/.Height set the TOTAL box size (border + padding
// included), so we pass width/height directly.
func Panel(content string, width, height int) string {
	s := panelStyle
	if width > 0 {
		s = s.Width(width)
	}
	if height > 0 {
		s = s.Height(height)
	}
	return s.Render(content)
}

// PanelCentered wraps content in a bordered box like Panel, but centers the
// content both horizontally and vertically within the panel's inner area.
func PanelCentered(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return Panel(content, width, height)
	}
	innerW, innerH := panelInner(width, height)
	centered := lipgloss.Place(innerW, innerH, lipgloss.Center, lipgloss.Center, content)
	return Panel(centered, width, height)
}

// Help renders footer help text in a muted style.
func Help(text string) string {
	return helpStyle.Render(text)
}

// searchBoxStyle frames the search input. It uses the same horizontal padding
// as the panel so both boxes render to the same outer width.
var searchBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorBorder).
	Padding(0, 2)

// searchPromptStyle styles the leading search glyph.
var searchPromptStyle = lipgloss.NewStyle().Foreground(colorAccent)

// SearchBox renders the search input, sized to match the panel's outer width.
// When the query is empty a muted placeholder is shown.
func SearchBox(query string, width int) string {
	prompt := searchPromptStyle.Render("> ")

	var text string
	if query == "" {
		text = muted("search…")
	} else {
		text = query + "▏" // simple block cursor
	}

	s := searchBoxStyle
	if width > 0 {
		// lipgloss .Width is the total box width (border + padding included),
		// matching Panel so both boxes span the same outer width.
		s = s.Width(width)
	}
	return s.Render(prompt + text)
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
