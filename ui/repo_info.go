package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/repos"
)

// gitInfo is the cached git metadata for a repo path.
type gitInfo struct {
	branch  string
	commits []string
	loaded  bool
}

// gitInfoMsg delivers the result of an async git-info fetch.
type gitInfoMsg struct {
	path    string
	branch  string
	commits []string
}

// gitInfoCmd fetches git metadata for a repo path off the UI thread.
func gitInfoCmd(path string) tea.Cmd {
	return func() tea.Msg {
		return gitInfoMsg{
			path:    path,
			branch:  repos.GitBranch(path),
			commits: repos.GitRecentCommits(path, 5),
		}
	}
}

// infoWidthThreshold is the minimum terminal width at which the info panel is
// shown alongside the list. Narrower than this, the list takes the full width.
const infoWidthThreshold = 90

// infoPanelWidth is the outer width of the info panel.
const infoPanelWidth = 40

// --- extensible section model ---
//
// The info panel is composed of independent sections. To add or remove
// information, add or remove an entry in infoSections — each section decides
// its own title and how to render its body from the given repo + git info.

// infoSection is one titled block in the info panel.
type infoSection struct {
	title  string
	render func(repo *repos.Repo, info gitInfo) string
}

// infoSections is the ordered list of sections shown in the info panel.
// Append or remove entries here to change what the panel displays.
var infoSections = []infoSection{
	{
		title: "Branch",
		render: func(_ *repos.Repo, info gitInfo) string {
			if info.branch == "" {
				return muted("—")
			}
			return infoBranchStyle.Render(info.branch)
		},
	},
	{
		title: "Recent commits",
		render: func(_ *repos.Repo, info gitInfo) string {
			if len(info.commits) == 0 {
				return muted("—")
			}
			lines := make([]string, len(info.commits))
			for i, c := range info.commits {
				lines[i] = "• " + c
			}
			return lipgloss.JoinVertical(lipgloss.Left, lines...)
		},
	},
}

// --- styles ---

var (
	infoPanelStyle        lipgloss.Style
	infoSectionTitleStyle lipgloss.Style
	infoBranchStyle       lipgloss.Style
)

// renderRepoInfo renders the info panel for the given repo, sized to the panel
// height. Sections are drawn in order with a blank line between them.
func (m model) renderRepoInfo(repo *repos.Repo, height int) string {
	blocks := make([]string, 0, len(infoSections)*3)

	if repo == nil {
		blocks = append(blocks, muted("No repo selected."))
	} else {
		info := m.gitCache[repo.Path]
		blocks = append(blocks, infoSectionTitleStyle.Render(repo.Alias), "")
		for i, sec := range infoSections {
			blocks = append(blocks, infoSectionTitleStyle.Render(sec.title))
			blocks = append(blocks, sec.render(repo, info))
			if i < len(infoSections)-1 {
				blocks = append(blocks, "")
			}
		}
	}

	body := lipgloss.JoinVertical(lipgloss.Left, blocks...)

	s := infoPanelStyle
	// Match Panel: .Width/.Height are total (border + padding included).
	s = s.Width(infoPanelWidth)
	if height > 0 {
		s = s.Height(height)
	}
	return s.Render(body)
}

// selectedRepo returns the repo currently under the cursor, or nil.
func (m model) selectedRepo() *repos.Repo {
	ordered := orderedRepos(m.query)
	if len(ordered) == 0 || m.repoCursor < 0 || m.repoCursor >= len(ordered) {
		return nil
	}
	return ordered[m.repoCursor]
}

// ensureGitInfo returns a Cmd to fetch git info for the selected repo if it is
// not already cached, or nil if nothing needs loading.
func (m *model) ensureGitInfo() tea.Cmd {
	repo := m.selectedRepo()
	if repo == nil {
		return nil
	}
	if m.gitCache == nil {
		m.gitCache = make(map[string]gitInfo)
	}
	if _, ok := m.gitCache[repo.Path]; ok {
		return nil // already loaded (or loading result stored)
	}
	// Mark as loading to avoid firing duplicate commands.
	m.gitCache[repo.Path] = gitInfo{}
	return gitInfoCmd(repo.Path)
}
