package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// sessionFinishedMsg is sent when a tmux session (run via tea.ExecProcess)
// exits and the bubbletea program resumes.
type sessionFinishedMsg struct {
	err error
}

type model struct {
	activeTab Tab

	// Terminal dimensions, updated on every resize.
	width  int
	height int

	// Repos tab state.
	repoCursor int
	query      string // current search box contents

	// launchedAlias is the repo whose session was most recently launched,
	// so the cursor can follow it after the list reorders.
	launchedAlias string

	// lastErr holds the most recent session error, shown in the footer.
	lastErr error
}

func initialModel() model {
	return model{
		activeTab: TabRepos,
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

	case sessionFinishedMsg:
		// tmux exited (user detached or session ended); bubbletea resumed.
		// The launched repo has moved to the top, so follow it with the cursor.
		m.lastErr = msg.err
		if m.launchedAlias != "" {
			for i, repo := range orderedRepos(m.query) {
				if repo.Alias == m.launchedAlias {
					m.repoCursor = i
					break
				}
			}
			m.launchedAlias = ""
		}

	case tea.KeyPressMsg:
		switch msg.String() {

		// Quit.
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		// Switch tabs.
		case "shift+tab", "left":
			m.activeTab = m.prevTab()
		case "tab", "right":
			m.activeTab = m.nextTab()

		// Navigate within the active tab.
		case "up":
			m.moveCursor(-1)
		case "down":
			m.moveCursor(1)

		// Launch / attach the selected repo's tmux session.
		case "enter":
			return m.launchSelectedRepo()

		// Clear the search query.
		case "esc", "alt+backspace":
			m.setQuery("")

		// Delete the last search character.
		case "backspace":
			if m.query != "" {
				m.setQuery(m.query[:len(m.query)-1])
			}

		// Any single printable character is typed into the search box.
		default:
			if s := msg.String(); len(s) == 1 && s[0] >= 0x20 && s[0] < 0x7f {
				m.setQuery(m.query + s)
			}
		}
	}

	return m, nil
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
	// tab bar, search box, and footer rows), with the content centered inside.
	panelHeight := m.height - 5 // tab bar (1) + search box (3, bordered) + footer (1)
	panel := PanelCentered(body, m.width, panelHeight)

	search := SearchBox(m.query, m.width)

	footer := Help("↑/↓ navigate · ←/→ tabs · type to search · esc clear · enter open · ctrl+c quit")
	if m.lastErr != nil {
		footer = Help("session error: "+m.lastErr.Error()) + "\n" + footer
	}

	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, panel, search, footer)

	// Take over the whole terminal window.
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

func CreateProgram() *tea.Program {
	return tea.NewProgram(initialModel())
}
