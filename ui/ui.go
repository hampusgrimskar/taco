package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

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
	var b strings.Builder

	b.WriteString(m.renderTabBar())
	b.WriteString("\n\n")

	switch m.activeTab {
	case TabRepos:
		b.WriteString(m.renderReposTab())
	case TabChats:
		b.WriteString(m.renderPlaceholder("Chats"))
	case TabComingSoon:
		b.WriteString(m.renderPlaceholder("Coming Soon"))
	}

	footer := "←/→ switch tabs · ↑/↓ navigate · q quit"
	content := m.padToHeight(b.String(), footer)

	// Take over the whole terminal window.
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// padToHeight pushes the footer to the bottom of the terminal by inserting
// blank lines between the body and the footer, so the view fills the window.
func (m model) padToHeight(body, footer string) string {
	// Before the first WindowSizeMsg arrives we don't know the height.
	if m.height <= 0 {
		return body + "\n\n" + footer + "\n"
	}

	bodyLines := strings.Count(body, "\n") + 1
	footerLines := 1

	// Lines to add so body + padding + footer fills the height.
	padding := m.height - bodyLines - footerLines
	if padding < 1 {
		padding = 1
	}

	return body + strings.Repeat("\n", padding) + footer
}

// renderTabBar draws the tab titles, marking the active one.
func (m model) renderTabBar() string {
	parts := make([]string, len(tabTitles))
	for i, title := range tabTitles {
		if Tab(i) == m.activeTab {
			parts[i] = fmt.Sprintf("[ %s ]", title)
		} else {
			parts[i] = fmt.Sprintf("  %s  ", title)
		}
	}
	return strings.Join(parts, " ")
}

// renderReposTab draws the repo menu.
func (m model) renderReposTab() string {
	if len(m.repoAliases) == 0 {
		return "No repos yet."
	}

	var b strings.Builder
	for i, alias := range m.repoAliases {
		cursor := " "
		if m.repoCursor == i {
			cursor = ">"
		}
		fmt.Fprintf(&b, "%s %s\n", cursor, alias)
	}
	return b.String()
}

// renderPlaceholder is shown for tabs with no content yet.
func (m model) renderPlaceholder(name string) string {
	return fmt.Sprintf("%s — nothing here yet.", name)
}

func CreateProgram() *tea.Program {
	return tea.NewProgram(initialModel())
}
