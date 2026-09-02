package ui

import (
	"os/exec"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/commands"
	"github.com/hampusgrimskar/taco/repos"
	"github.com/hampusgrimskar/taco/session"
)

// orderedRepos returns the repos in display order: active sessions first
// (alphabetical), then the rest (alphabetical), filtered by a case-insensitive
// substring match against query. This is the single source of truth for the
// Repos tab, so the cursor index, rendering, and launch all agree.
func orderedRepos(query string) []*repos.Repo {
	all := repos.All()
	q := strings.ToLower(strings.TrimSpace(query))

	active := make([]*repos.Repo, 0, len(all))
	inactive := make([]*repos.Repo, 0, len(all))
	for _, r := range all {
		if q != "" && !strings.Contains(strings.ToLower(r.Alias), q) {
			continue
		}
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

// visibleRepoRows returns how many repo rows fit in the panel's inner height
// for the current terminal size.
func (m model) visibleRepoRows() int {
	panelHeight := m.height - 5
	_, innerHeight := panelInner(m.width, panelHeight)
	if innerHeight < 1 {
		return 1
	}
	return innerHeight
}

// moveCursor adjusts the cursor for whichever tab is active, scrolling the
// repo list only when the cursor would move past a visible edge.
func (m *model) moveCursor(delta int) {
	if m.activeTab != TabRepos {
		return
	}
	count := len(orderedRepos(m.query))
	next := m.repoCursor + delta
	if next < 0 || next >= count {
		return
	}
	m.repoCursor = next
	m.clampScroll(count)
}

// clampScroll nudges the scroll offset so the cursor stays within the visible
// window, scrolling only at the edges. It reserves rows for scroll hints so
// the visible content plus hints fits.
func (m *model) clampScroll(total int) {
	visible := m.visibleRepoRows()
	rows := visibleItemRows(total, visible)

	// Scroll up if the cursor is above the window.
	if m.repoCursor < m.repoScroll {
		m.repoScroll = m.repoCursor
	}
	// Scroll down if the cursor is at/below the bottom visible row.
	if m.repoCursor >= m.repoScroll+rows {
		m.repoScroll = m.repoCursor - rows + 1
	}
	// Keep the offset in range.
	maxScroll := total - rows
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.repoScroll > maxScroll {
		m.repoScroll = maxScroll
	}
	if m.repoScroll < 0 {
		m.repoScroll = 0
	}
}

// visibleItemRows returns how many item rows are shown given the total item
// count and the available height, reserving rows for scroll hints when the
// list does not fully fit.
func visibleItemRows(total, visible int) int {
	if total <= visible {
		return visible
	}
	rows := visible - 2 // reserve for top + bottom hints
	if rows < 1 {
		rows = 1
	}
	return rows
}

// launchSelectedRepo starts (or attaches to) the tmux session for the repo
// under the cursor, handing the terminal to tmux via tea.ExecProcess. The
// bubbletea program is suspended while tmux runs and resumes on exit.
func (m model) launchSelectedRepo() (tea.Model, tea.Cmd) {
	ordered := orderedRepos(m.query)
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

	// Reset the search so the full list is shown again on return.
	m.query = ""
	m.repoScroll = 0

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

// renderReposTab draws the repo menu with active sessions first, each marked
// by an indicator aligned in a column to the right of the longest alias. Only
// visibleRows rows are drawn; the window scrolls to keep the cursor in view.
func (m model) renderReposTab(visibleRows int) string {
	// Before the first WindowSizeMsg (or on a tiny terminal) the computed
	// height can be zero or negative; clamp so slice bounds stay valid.
	if visibleRows < 1 {
		visibleRows = 1
	}

	ordered := orderedRepos(m.query)
	if len(repos.All()) == 0 {
		return muted("No repos yet.")
	}
	if len(ordered) == 0 {
		return muted("No matches.")
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

	total := len(ordered)
	rows := visibleItemRows(total, visibleRows)

	// Window using the persistent scroll offset; clamp defensively in case the
	// terminal was resized since the last cursor move.
	start := m.repoScroll
	if start > total-rows {
		start = total - rows
	}
	if start < 0 {
		start = 0
	}
	end := start + rows
	if end > total {
		end = total
	}

	// Build exactly visibleRows lines so the panel height never changes as
	// scroll hints appear or disappear. When scrolling is possible, the first
	// and last lines are reserved as hint slots (blank when not scrolled).
	lines := make([]string, 0, visibleRows)

	scrollable := total > visibleRows
	if scrollable {
		if start > 0 {
			lines = append(lines, muted("  ↑ more"))
		} else {
			lines = append(lines, "")
		}
	}

	for i := start; i < end; i++ {
		repo := ordered[i]
		lines = append(lines, RepoRow(repo.Alias, indicatorCol, repo.Session != nil, m.repoCursor == i))
	}

	if scrollable {
		if end < total {
			lines = append(lines, muted("  ↓ more"))
		} else {
			lines = append(lines, "")
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
