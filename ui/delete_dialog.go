package ui

import (
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/commands"
	"github.com/hampusgrimskar/taco/repos"
)

// deleteDialog holds the state of the delete-confirmation modal.
type deleteDialog struct {
	alias   string       // repo being deleted
	confirm dialogButton // focused button (buttonSave = Delete, buttonCancel = Cancel)
	err     string
}

// openDeleteDialog opens the confirmation for the selected repo.
func (m *model) openDeleteDialog() {
	ordered := orderedRepos(m.query)
	if m.activeTab != TabRepos || len(ordered) == 0 {
		return
	}
	if m.repoCursor >= len(ordered) {
		m.repoCursor = len(ordered) - 1
	}
	m.deleting = true
	m.del = deleteDialog{
		alias:   ordered[m.repoCursor].Alias,
		confirm: buttonCancel, // default to the safe choice
	}
}

// closeDeleteDialog exits delete mode.
func (m *model) closeDeleteDialog() {
	m.deleting = false
	m.del = deleteDialog{}
}

// updateDeleteDialog handles a key while the confirmation is open.
func (m *model) updateDeleteDialog(key string) {
	switch key {
	case "esc", "n":
		m.closeDeleteDialog()

	case "left", "right", "tab", "shift+tab":
		if m.del.confirm == buttonSave {
			m.del.confirm = buttonCancel
		} else {
			m.del.confirm = buttonSave
		}

	case "y":
		m.commitDelete()

	case "enter":
		if m.del.confirm == buttonCancel {
			m.closeDeleteDialog()
			return
		}
		m.commitDelete()
	}
}

// commitDelete terminates any live session for the repo, removes it, and moves
// the cursor to a valid position.
func (m *model) commitDelete() {
	alias := m.del.alias
	if repo := repos.Find(alias); repo != nil && repo.Session != nil {
		// Best-effort terminate; ignore error (session may already be gone).
		_ = commands.TerminateSession(repo.Session.ID).Run()
		repo.Session = nil
	}

	if err := repos.Delete(alias); err != nil {
		m.del.err = err.Error()
		return
	}

	m.closeDeleteDialog()

	// Keep the cursor within bounds after removal.
	count := len(orderedRepos(m.query))
	if m.repoCursor >= count {
		m.repoCursor = count - 1
	}
	if m.repoCursor < 0 {
		m.repoCursor = 0
	}
	m.clampScroll(count)
}

// --- rendering ---

var deleteBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("196")). // red border for a destructive action
	Padding(1, 2)

// renderDeleteDialog draws the confirmation modal.
func (m model) renderDeleteDialog() string {
	title := dialogTitleStyle.Render("Delete repo")
	question := "Remove \"" + m.del.alias + "\" from taco?"

	cancel := buttonStyle.Render("Cancel")
	del := buttonStyle.Render("Delete")
	if m.del.confirm == buttonSave {
		del = buttonFocusedStyle.Render("Delete")
	} else {
		cancel = buttonFocusedStyle.Render("Cancel")
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, cancel, del)

	parts := []string{title, "", question, ""}
	if m.del.err != "" {
		parts = append(parts, dialogErrStyle.Render(m.del.err), "")
	}
	parts = append(parts, buttons, "", Help("y delete · n/esc cancel · ←/→ switch · enter confirm"))

	body := lipgloss.JoinVertical(lipgloss.Center, parts...)
	return deleteBoxStyle.Render(body)
}
