package ui

import (
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/repos"
)

// dialogButton identifies the focused button in the rename dialog.
type dialogButton int

const (
	buttonSave dialogButton = iota
	buttonCancel
)

// renameDialog holds the state of the modal rename dialog. It is active only
// while model.renaming is true.
type renameDialog struct {
	original string       // alias being renamed (identity key)
	input    string       // current text input contents
	focus    dialogButton // which button is focused
	err      string       // validation error to display, if any
}

// openRenameDialog puts the model into rename mode for the currently selected
// repo, seeding the input with its existing alias.
func (m *model) openRenameDialog() {
	ordered := orderedRepos(m.query)
	if m.activeTab != TabRepos || len(ordered) == 0 {
		return
	}
	if m.repoCursor >= len(ordered) {
		m.repoCursor = len(ordered) - 1
	}
	alias := ordered[m.repoCursor].Alias

	m.renaming = true
	m.dialog = renameDialog{
		original: alias,
		input:    alias,
		focus:    buttonSave,
	}
}

// closeRenameDialog exits rename mode.
func (m *model) closeRenameDialog() {
	m.renaming = false
	m.dialog = renameDialog{}
}

// updateRenameDialog handles a key while the rename dialog is open. It returns
// true if the dialog consumed the key (it always does, being modal).
func (m *model) updateRenameDialog(key string) {
	switch key {
	case "esc":
		m.closeRenameDialog()

	case "left", "right", "tab", "shift+tab":
		// Toggle button focus.
		if m.dialog.focus == buttonSave {
			m.dialog.focus = buttonCancel
		} else {
			m.dialog.focus = buttonSave
		}

	case "enter":
		if m.dialog.focus == buttonCancel {
			m.closeRenameDialog()
			return
		}
		m.commitRename()

	case "backspace":
		if m.dialog.input != "" {
			m.dialog.input = m.dialog.input[:len(m.dialog.input)-1]
		}

	case " ", "space":
		m.dialog.input += " "

	default:
		// Any single printable character is typed into the input.
		if len(key) == 1 && key[0] >= 0x20 && key[0] < 0x7f {
			m.dialog.input += key
		}
	}
}

// commitRename attempts to persist the new alias. On success it closes the
// dialog and moves the cursor to follow the renamed repo; on failure it keeps
// the dialog open and shows the error.
func (m *model) commitRename() {
	newAlias := m.dialog.input
	if err := repos.Rename(m.dialog.original, newAlias); err != nil {
		m.dialog.err = err.Error()
		return
	}

	m.closeRenameDialog()

	// Follow the renamed repo with the cursor in the reordered list.
	for i, repo := range orderedRepos(m.query) {
		if repo.Alias == newAlias {
			m.repoCursor = i
			break
		}
	}
	m.clampScroll(len(orderedRepos(m.query)))
}

// --- rendering ---

// Theme-independent dialog styles.
var (
	dialogTitleStyle = lipgloss.NewStyle().Bold(true)
	buttonStyle      = lipgloss.NewStyle().Padding(0, 3).Margin(0, 1)
	dialogErrStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red
)

// Theme-dependent dialog styles, (re)assigned by rebuildThemedStyles.
var (
	dialogBoxStyle     lipgloss.Style
	dialogInputStyle   lipgloss.Style
	buttonFocusedStyle lipgloss.Style
)

// renderRenameDialog draws the modal rename dialog.
func (m model) renderRenameDialog() string {
	title := dialogTitleStyle.Render("Rename repo")

	inputWidth := len(m.dialog.original) + 12
	if inputWidth < 24 {
		inputWidth = 24
	}
	input := dialogInputStyle.Width(inputWidth).Render(m.dialog.input + "▏")

	save := buttonStyle.Render("Save")
	cancel := buttonStyle.Render("Cancel")
	if m.dialog.focus == buttonSave {
		save = buttonFocusedStyle.Render("Save")
	} else {
		cancel = buttonFocusedStyle.Render("Cancel")
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, save, cancel)

	parts := []string{title, "", input, ""}
	if m.dialog.err != "" {
		parts = append(parts, dialogErrStyle.Render(m.dialog.err), "")
	}
	parts = append(parts, buttons)

	body := lipgloss.JoinVertical(lipgloss.Center, parts...)
	return dialogBoxStyle.Render(body)
}
