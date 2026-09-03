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

// pad2 pads s with trailing spaces to exactly w display cells.
func pad2(s string, w int) string {
	return s + spaces(w-lipgloss.Width(s))
}

// truncateRight shortens s to at most max display cells, appending "…".
func truncateRight(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	runes := []rune(s)
	keep := max - 1
	if keep > len(runes) {
		keep = len(runes)
	}
	return string(runes[:keep]) + "…"
}

// spaces returns n space characters.
func spaces(n int) string {
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

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

// visibleRepoRows returns how many repo data rows fit in the table for the
// current terminal size, after the panel border/padding and the table's
// header + rule lines.
func (m model) visibleRepoRows() int {
	panelHeight := m.height - 5
	_, innerHeight := panelInner(m.width, panelHeight)
	rows := innerHeight - 2 // table header + rule
	if rows < 1 {
		return 1
	}
	return rows
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

// Repos table styles.
var (
	reposTableStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
	reposHeaderStyle   lipgloss.Style // (re)built by theme
	reposSelectedStyle lipgloss.Style // full-width selected-row bar
	reposActiveStyle   lipgloss.Style // active status marker
)

// renderReposTab draws the repos table (Name · Branch · Status) filling the
// given width/height. Active sessions sort first; the window scrolls to keep
// the cursor visible.
func (m model) renderReposTab(width, height int) string {
	innerRows := height - 4 // border (2) + vertical padding (2)
	if innerRows < 1 {
		innerRows = 1
	}

	if len(repos.All()) == 0 {
		return reposTableStyle.BorderForeground(colorBorder).
			Width(width).Height(height).Render(muted("No repos yet. Press ctrl+n to add."))
	}
	ordered := orderedRepos(m.query)
	if len(ordered) == 0 {
		return reposTableStyle.BorderForeground(colorBorder).
			Width(width).Height(height).Render(muted("No matches."))
	}

	// Column layout across the inner width.
	// reposTableStyle: border (2) + horizontal padding (4) = 6.
	innerW := width - 6
	if innerW < 12 {
		innerW = 12
	}
	// Reserve one space on each side inside the row so text (and the selected
	// highlight bar) isn't flush against the edges.
	const sidePad = 1
	contentW := innerW - 2*sidePad
	if contentW < 8 {
		contentW = 8
	}
	// Name and Branch share the width; Status is a fixed narrow column.
	const statusW = 8
	remaining := contentW - statusW
	if remaining < 4 {
		remaining = 4
	}
	nameW := remaining / 2
	branchW := remaining - nameW
	rowWidth := innerW

	const colGap = 2
	sp := spaces(sidePad)
	fit := func(s string, w int) string { return pad2(truncateRight(s, w-colGap), w) }
	rowText := func(name, branch, status string) string {
		return sp + fit(name, nameW) + fit(branch, branchW) + status + sp
	}

	header := reposHeaderStyle.Render(rowText("Name", "Branch", "Status"))
	rule := reposHeaderStyle.Render(strings.Repeat("─", innerW))

	// Rows available after header + rule.
	rowsAvail := innerRows - 2
	if rowsAvail < 1 {
		rowsAvail = 1
	}

	total := len(ordered)
	rows := visibleItemRows(total, rowsAvail)

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

	lines := make([]string, 0, rowsAvail+2)
	lines = append(lines, header, rule)

	if start > 0 {
		lines = append(lines, muted("↑ more"))
	}
	for i := start; i < end; i++ {
		repo := ordered[i]
		branch := repos.GitBranch(repo.Path)
		if branch == "" {
			branch = "—"
		}
		active := repo.Session != nil
		statusText := "idle"
		if active {
			statusText = "● active"
		}

		selected := i == m.repoCursor
		if selected {
			// Selected row: keep all cells plain so the selected style's
			// (contrasting) foreground applies uniformly and stays readable.
			line := rowText(repo.Alias, branch, statusText)
			lines = append(lines, reposSelectedStyle.Width(rowWidth).Render(line))
		} else {
			// Unselected: color the status cell.
			status := muted(statusText)
			if active {
				status = reposActiveStyle.Render(statusText)
			}
			lines = append(lines, rowText(repo.Alias, branch, status))
		}
	}
	if end < total {
		lines = append(lines, muted("↓ more"))
	}

	table := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return reposTableStyle.BorderForeground(colorBorder).
		Width(width).Height(height).Render(table)
}
