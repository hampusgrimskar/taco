package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/repos"
)

// Tab identifies a top-level view in the UI.
type Tab int

const (
	TabRepos Tab = iota
	TabChats
	TabComingSoon
)

// tabTitles are the labels shown in the tab bar, in order.
var tabTitles = []string{"Repos", "Chats", "Coming Soon"}

type model struct {
	activeTab Tab

	// Terminal dimensions, updated on every resize.
	width  int
	height int

	// Repos tab state.
	repoAliases []string
	repoCursor  int
}

func initialModel() model {
	return model{
		activeTab:   TabRepos,
		repoAliases: repos.Aliases(),
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch msg.String() {

		// Quit.
		case "ctrl+c", "q":
			return m, tea.Quit

		// Switch tabs.
		case "left", "h", "shift+tab":
			m.activeTab = m.prevTab()
		case "right", "l", "tab":
			m.activeTab = m.nextTab()

		// Move within the active tab.
		case "up", "k":
			m.moveCursor(-1)
		case "down", "j":
			m.moveCursor(1)
		}
	}

	return m, nil
}

// nextTab / prevTab cycle through the tabs.
func (m model) nextTab() Tab {
	return Tab((int(m.activeTab) + 1) % len(tabTitles))
}

func (m model) prevTab() Tab {
	return Tab((int(m.activeTab) - 1 + len(tabTitles)) % len(tabTitles))
}

// moveCursor adjusts the cursor for whichever tab is active.
func (m *model) moveCursor(delta int) {
	if m.activeTab != TabRepos {
		return
	}
	next := m.repoCursor + delta
	if next >= 0 && next < len(m.repoAliases) {
		m.repoCursor = next
	}
}

func (m model) View() tea.View {
	// Tab bar across the top.
	tabBar := TabBar(tabTitles, int(m.activeTab))

	// Body for the active tab.
	var body string
	switch m.activeTab {
	case TabRepos:
		body = m.renderReposTab()
	case TabChats:
		body = m.renderPlaceholder("Chats")
	case TabComingSoon:
		body = m.renderPlaceholder("Coming Soon")
	}

	// Wrap the body in a bordered panel that fills the window (minus the
	// tab bar and footer rows).
	panelHeight := m.height - 4 // tab bar (1) + blank (1) + footer (2)
	panel := Panel(body, m.width, panelHeight)

	footer := Help("←/→ switch tabs · ↑/↓ navigate · q quit")

	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, panel, footer)

	// Take over the whole terminal window.
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderReposTab draws the repo menu.
func (m model) renderReposTab() string {
	if len(m.repoAliases) == 0 {
		return muted("No repos yet.")
	}

	rows := make([]string, len(m.repoAliases))
	for i, alias := range m.repoAliases {
		rows[i] = Row(alias, m.repoCursor == i)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

// renderPlaceholder is shown for tabs with no content yet.
func (m model) renderPlaceholder(name string) string {
	return muted(name + " — nothing here yet.")
}

func CreateProgram() *tea.Program {
	return tea.NewProgram(initialModel())
}
