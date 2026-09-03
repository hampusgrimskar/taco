package ui

import (
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

// Theme is a named color palette. Fields are ANSI-256 color codes.
type Theme struct {
	Name       string
	Accent     string // highlight / selection
	Foreground string // text on accent background
	Muted      string // dim text
	Border     string // box borders
	Good       string // positive indicator (e.g. active session, branch)
}

// Themes is the ordered list of available themes. Add entries here to offer
// more choices in the settings form.
var Themes = []Theme{
	{Name: "Default", Accent: "205", Foreground: "230", Muted: "240", Border: "63", Good: "42"},
	{Name: "Ocean", Accent: "39", Foreground: "231", Muted: "244", Border: "24", Good: "45"},
	{Name: "Forest", Accent: "40", Foreground: "233", Muted: "240", Border: "22", Good: "120"},
	{Name: "Mono", Accent: "252", Foreground: "232", Muted: "242", Border: "245", Good: "250"},
}

// activeTheme is the currently applied theme (defaults to the first).
var activeTheme = Themes[0]

// Active color variables, derived from activeTheme. Style helpers read these.
var (
	colorAccent color.Color
	colorFg     color.Color
	colorMuted  color.Color
	colorBorder color.Color
	colorGood   color.Color
)

// Theme-dependent styles. They are (re)assigned by rebuildStyles whenever the
// active theme changes. Theme-independent styles are declared with their own
// initializers elsewhere.
var (
	activeTabStyle   lipgloss.Style
	inactiveTabStyle lipgloss.Style
	panelStyle       lipgloss.Style
	selectedRowStyle lipgloss.Style
	rowStyle         lipgloss.Style
	helpStyle        lipgloss.Style
)

func init() {
	// Apply the default theme so styles are ready before first render.
	applyThemeColors(activeTheme)
	rebuildStyles()
}

// ThemeNames returns the available theme names in order.
func ThemeNames() []string {
	names := make([]string, len(Themes))
	for i, t := range Themes {
		names[i] = t.Name
	}
	return names
}

// ApplyThemeByName switches to the named theme (if it exists) and rebuilds all
// theme-dependent styles. Unknown names are ignored.
func ApplyThemeByName(name string) {
	for _, t := range Themes {
		if t.Name == name {
			activeTheme = t
			applyThemeColors(t)
			rebuildStyles()
			return
		}
	}
}

// ActiveThemeName returns the name of the current theme.
func ActiveThemeName() string {
	return activeTheme.Name
}

func applyThemeColors(t Theme) {
	colorAccent = lipgloss.Color(t.Accent)
	colorFg = lipgloss.Color(t.Foreground)
	colorMuted = lipgloss.Color(t.Muted)
	colorBorder = lipgloss.Color(t.Border)
	colorGood = lipgloss.Color(t.Good)
}

// rebuildStyles reassigns every theme-dependent style from the current colors.
// This is the single place that ties colors to styles, so switching themes
// updates the whole app.
func rebuildStyles() {
	activeTabStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(colorFg).
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

	// Styles defined in other files that depend on theme colors.
	rebuildThemedStyles()
}

// rebuildThemedStyles reassigns theme-dependent styles that are declared in
// other files of the ui package. Kept together here so every color-driven
// style is rebuilt in one pass when the theme changes.
func rebuildThemedStyles() {
	// theme.go
	searchBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(0, 2)
	searchPromptStyle = lipgloss.NewStyle().Foreground(colorAccent)

	// rename_dialog.go
	dialogBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2)
	dialogInputStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1)
	buttonFocusedStyle = lipgloss.NewStyle().
		Padding(0, 3).Margin(0, 1).
		Bold(true).
		Foreground(colorFg).
		Background(colorAccent)

	// add_dialog.go
	addDialogStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2)

	// repo_info.go
	infoPanelStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorder).
		Padding(1, 2)
	infoSectionTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
	infoBranchStyle = lipgloss.NewStyle().Foreground(colorGood)

	// repos_tab.go — table header, selected-row bar, active status marker.
	reposHeaderStyle = lipgloss.NewStyle().Faint(true)
	reposSelectedStyle = lipgloss.NewStyle().
		Foreground(colorFg).
		Background(colorAccent)
	reposActiveStyle = lipgloss.NewStyle().Foreground(colorGood)
}

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
var searchBoxStyle lipgloss.Style

// searchPromptStyle styles the leading search glyph.
var searchPromptStyle lipgloss.Style

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
