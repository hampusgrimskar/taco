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

// moveCursor adjusts the cursor for whichever tab is active.
func (m *model) moveCursor(delta int) {
	if m.activeTab != TabRepos {
		return
	}
	count := len(orderedRepos(m.query))
	next := m.repoCursor + delta
	if next >= 0 && next < count {
		m.repoCursor = next
	}
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
// by an indicator aligned in a column to the right of the longest alias.
func (m model) renderReposTab() string {
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

	rows := make([]string, len(ordered))
	for i, repo := range ordered {
		rows[i] = RepoRow(repo.Alias, indicatorCol, repo.Session != nil, m.repoCursor == i)
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}
