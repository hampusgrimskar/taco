package ui

import (
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/settings"
)

// Setting keys persisted in ~/.taco/settings.
const settingKeyTheme = "theme"

// settingsForm holds the state of the settings modal.
type settingsForm struct {
	cursor int // focused field index
}

// --- extensible field registry ---
//
// Each setting is a settingField. To add a new setting, append a descriptor to
// settingFields — the form renders and edits it generically. A field exposes
// its label, its current display value, the list of options it cycles through,
// and an apply hook invoked (live) whenever its value changes.

type settingField struct {
	key     string
	label   string
	options func() []string    // selectable values, in order
	current func() string      // currently selected value
	apply   func(value string) // called live when the value changes (persist here)
}

// settingFields is the ordered list of settings shown in the form.
var settingFields = []settingField{
	{
		key:     settingKeyTheme,
		label:   "Color theme",
		options: ThemeNames,
		current: ActiveThemeName,
		apply: func(value string) {
			ApplyThemeByName(value) // live preview
			_ = settings.Set(settingKeyTheme, value)
		},
	},
}

// openSettings opens the settings modal.
func (m *model) openSettings() {
	m.settingsOpen = true
	m.settings = settingsForm{}
}

// closeSettings exits the settings modal.
func (m *model) closeSettings() {
	m.settingsOpen = false
	m.settings = settingsForm{}
}

// updateSettings handles a key while the settings modal is open.
func (m *model) updateSettings(key string) {
	switch key {
	case "esc", "ctrl+p", "enter":
		m.closeSettings()
	case "up":
		if m.settings.cursor > 0 {
			m.settings.cursor--
		}
	case "down":
		if m.settings.cursor < len(settingFields)-1 {
			m.settings.cursor++
		}
	case "left":
		m.cycleField(-1)
	case "right":
		m.cycleField(1)
	}
}

// cycleField moves the focused field's value by delta options and applies it.
func (m *model) cycleField(delta int) {
	if m.settings.cursor < 0 || m.settings.cursor >= len(settingFields) {
		return
	}
	field := settingFields[m.settings.cursor]
	opts := field.options()
	if len(opts) == 0 {
		return
	}

	cur := field.current()
	idx := 0
	for i, o := range opts {
		if o == cur {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(opts)) % len(opts)
	field.apply(opts[idx])
}

// --- rendering ---

var settingsBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1, 2)

// renderSettings draws the settings modal.
func (m model) renderSettings() string {
	title := dialogTitleStyle.Render("Settings")

	rows := make([]string, 0, len(settingFields))
	for i, field := range settingFields {
		label := field.label + ":"
		value := "‹ " + field.current() + " ›"

		line := label + "  " + value
		if i == m.settings.cursor {
			line = selectedRowStyle.Render("› " + line)
		} else {
			line = "  " + line
		}
		rows = append(rows, line)
	}
	list := lipgloss.JoinVertical(lipgloss.Left, rows...)

	hint := Help("↑/↓ field · ←/→ change · esc/enter close")

	body := lipgloss.JoinVertical(lipgloss.Left, title, "", list, "", hint)

	// Border uses the accent color so it re-tints live as the theme changes.
	return settingsBoxStyle.BorderForeground(colorAccent).Render(body)
}
