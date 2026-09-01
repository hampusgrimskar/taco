package ui

import (
	"os/exec"
	"sort"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/commands"
	"github.com/hampusgrimskar/taco/repos"
	"github.com/hampusgrimskar/taco/session"
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

// orderedRepos returns the repos in display order: active sessions first
// (alphabetical), then the rest (alphabetical). This is the single source of
// truth for the Repos tab, so the cursor index, rendering, and launch all
// agree on the same ordering.
func orderedRepos() []*repos.Repo {
	all := repos.All()

	active := make([]*repos.Repo, 0, len(all))
	inactive := make([]*repos.Repo, 0, len(all))
	for _, r := range all {
		if r.Session != nil {
			active = append(active, r)
		} else {
			inactive = append(inactive, r)
		}
	}

	byAlias := func(list []*repos.Repo) {
		sort.Slice(list, func(i, j int) bool {
			return list[i].Alias < list[j].Alias
		})
	}
	byAlias(active)
	byAlias(inactive)

	return append(active, inactive...)
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
			for i, repo := range orderedRepos() {
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

		// Launch / attach the selected repo's tmux session.
		case "enter":
			return m.launchSelectedRepo()
		}
	}

	return m, nil
}

// launchSelectedRepo starts (or attaches to) the tmux session for the repo
// under the cursor, handing the terminal to tmux via tea.ExecProcess. The
// bubbletea program is suspended while tmux runs and resumes on exit.
func (m model) launchSelectedRepo() (tea.Model, tea.Cmd) {
	ordered := orderedRepos()
	if m.activeTab != TabRepos || len(ordered) == 0 {
		return m, nil
	}
	if m.repoCursor >= len(ordered) {
		m.repoCursor = len(ordered) - 1
	}

	repo := ordered[m.repoCursor]

	// Remember which repo we launched so the cursor can follow it to its
	// new position once the list reorders on return.
	m.launchedAlias = repo.Alias

	// Decide whether to create+attach (first launch) or just attach
	// (session already exists from a previous launch this run).
	var cmd *exec.Cmd
	if repo.Session == nil {
		repo.Session = session.New()
		cmd = commands.CreateSession(repo.Session.ID, repo.Path)
	} else {
		cmd = commands.AttachToSession(repo.Session.ID)
	}

	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return sessionFinishedMsg{err: err}
	})
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
	count := len(orderedRepos())
	next := m.repoCursor + delta
	if next >= 0 && next < count {
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

	footer := Help("←/→ switch tabs · ↑/↓ navigate · enter open · q quit")
	if m.lastErr != nil {
		footer = Help("session error: "+m.lastErr.Error()) + "\n" + footer
	}

	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, panel, footer)

	// Take over the whole terminal window.
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// renderReposTab draws the repo menu with active sessions first, each marked
// by an indicator aligned in a column to the right of the longest alias.
func (m model) renderReposTab() string {
	ordered := orderedRepos()
	if len(ordered) == 0 {
		return muted("No repos yet.")
	}

	// Find the longest alias so the indicator column aligns, then place the
	// indicator 5 positions to the right of it.
	longest := 0
	for _, repo := range ordered {
		if len(repo.Alias) > longest {
			longest = len(repo.Alias)
		}
	}
	indicatorCol := longest + 5

	rows := make([]string, len(ordered))
	for i, repo := range ordered {
		rows[i] = RepoRow(repo.Alias, indicatorCol, repo.Session != nil, m.repoCursor == i)
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
