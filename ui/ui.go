package ui

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/provider"
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
	repoScroll int    // index of the first visible row (scroll offset)
	query      string // current search box contents

	// gitCache holds git metadata per repo path for the info panel.
	gitCache map[string]gitInfo

	// Chats tab state.
	provider    provider.Provider
	chatCursor  int
	chatting    bool
	chat        chatWizard
	chatDelete  bool
	chatDelData chatDeleteData
	chatRename  bool
	chatRenData chatRenameData

	// launchedAlias is the repo whose session was most recently launched,
	// so the cursor can follow it after the list reorders.
	launchedAlias string

	// Rename dialog modal state.
	renaming bool
	dialog   renameDialog

	// Add-repos wizard modal state.
	adding bool
	add    addDialog

	// Delete-confirmation modal state.
	deleting bool
	del      deleteDialog

	// Settings modal state (global).
	settingsOpen bool
	settings     settingsForm

	// lastErr holds the most recent session error, shown in the footer.
	lastErr error
}

// scanDoneMsg is delivered when an async git-repo scan completes.
type scanDoneMsg struct {
	paths []string
	err   error
}

// chatFinishedMsg is delivered when a chat's tmux session is detached.
type chatFinishedMsg struct {
	dir    string
	err    error
	resume bool
}

func initialModel() model {
	return model{
		activeTab: TabRepos,
		provider:  provider.NewKiro(),
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
		// Load git info for the initially selected repo.
		return m, m.ensureGitInfo()

	case gitInfoMsg:
		if m.gitCache == nil {
			m.gitCache = make(map[string]gitInfo)
		}
		m.gitCache[msg.path] = gitInfo{
			branch:  msg.branch,
			commits: msg.commits,
			loaded:  true,
		}

	case sessionFinishedMsg:
		// tmux exited (user detached or session ended); bubbletea resumed.
		// The launched repo has moved to the top, so follow it with the cursor.
		m.lastErr = msg.err
		if m.launchedAlias != "" {
			ordered := orderedRepos(m.query)
			for i, repo := range ordered {
				if repo.Alias == m.launchedAlias {
					m.repoCursor = i
					break
				}
			}
			m.clampScroll(len(ordered))
			m.launchedAlias = ""
		}

	case scanDoneMsg:
		m.onScanDone(msg)

	case chatFinishedMsg:
		m.onChatFinished(msg)

	case tea.KeyPressMsg:
		key := msg.String()

		// Settings is a global modal: it opens from any tab and consumes keys.
		if m.settingsOpen {
			m.updateSettings(key)
			return m, nil
		}

		// When the rename dialog is open it is modal: it consumes all keys.
		if m.renaming {
			m.updateRenameDialog(key)
			return m, nil
		}

		// When the add wizard is open it is modal too (and may run a scan Cmd).
		if m.adding {
			return m, m.updateAddDialog(key)
		}

		// When the delete confirmation is open it is modal.
		if m.deleting {
			m.updateDeleteDialog(key)
			return m, nil
		}

		// When the new-chat wizard is open it is modal (may run an Exec Cmd).
		if m.chatting {
			return m, m.updateChatWizard(key)
		}

		// When the chat delete confirmation is open it is modal.
		if m.chatDelete {
			m.updateChatDelete(key)
			return m, nil
		}

		// When the chat rename dialog is open it is modal.
		if m.chatRename {
			m.updateChatRename(key)
			return m, nil
		}

		switch key {

		// Quit.
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		// Open global settings.
		case "ctrl+p":
			m.openSettings()

		// Open the rename dialog for the selected repo/chat.
		case "ctrl+r":
			if m.activeTab == TabChats {
				m.openChatRename()
			} else {
				m.openRenameDialog()
			}

		// Open the add-repos wizard.
		case "ctrl+n":
			m.openAddDialog()

		// Open the delete confirmation for the selected repo/chat.
		case "ctrl+d":
			if m.activeTab == TabChats {
				m.openChatDelete()
			} else {
				m.openDeleteDialog()
			}

		// Switch tabs (Tab / Shift+Tab only, so ←/→ are free for in-tab use).
		case "shift+tab":
			m.activeTab = m.prevTab()
		case "tab":
			m.activeTab = m.nextTab()

		// Navigate within the active tab.
		case "up":
			if m.activeTab == TabChats {
				m.moveChatCursor(-1)
				return m, nil
			}
			m.moveCursor(-1)
			return m, m.ensureGitInfo()
		case "down":
			if m.activeTab == TabChats {
				m.moveChatCursor(1)
				return m, nil
			}
			m.moveCursor(1)
			return m, m.ensureGitInfo()

		// Enter: resume a chat on the Chats tab, else launch a repo session.
		case "enter":
			if m.activeTab == TabChats {
				return m.resumeSelectedChat()
			}
			return m.launchSelectedRepo()

		// Space: on the Chats tab, start a new chat when not searching; once a
		// search is in progress, space is part of the query. Elsewhere it is
		// search input.
		case " ", "space":
			if m.activeTab == TabChats && m.query == "" {
				return m, m.openChatWizard()
			}
			m.setQuery(m.query + " ")
			return m, m.ensureGitInfo()

		// Clear the search query.
		case "esc", "alt+backspace":
			m.setQuery("")
			return m, m.ensureGitInfo()

		// Delete the last search character.
		case "backspace":
			if m.query != "" {
				m.setQuery(m.query[:len(m.query)-1])
			}
			return m, m.ensureGitInfo()

		// Any single printable character is typed into the search box.
		default:
			if len(key) == 1 && key[0] >= 0x20 && key[0] < 0x7f {
				m.setQuery(m.query + key)
				return m, m.ensureGitInfo()
			}
		}
	}

	return m, nil
}

func (m model) View() tea.View {
	// Until the first WindowSizeMsg arrives we have no real dimensions.
	// Render an empty full-screen view to avoid negative-size math.
	if m.width <= 0 || m.height <= 0 {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	// Tab bar across the top.
	tabBar := TabBar(tabTitles, int(m.activeTab))

	// Panel fills the window minus tab bar (1), search box (3), footer (1).
	panelHeight := m.height - 5
	_, innerHeight := panelInner(m.width, panelHeight)

	// Body for the active tab. The repos list is windowed to the panel's
	// inner height so it scrolls rather than overflowing, then centered.
	var panel string
	switch m.activeTab {
	case TabRepos:
		switch {
		case m.renaming:
			// Modal: show the dialog centered in the panel.
			panel = PanelCentered(m.renderRenameDialog(), m.width, panelHeight)
		case m.adding:
			panel = PanelCentered(m.renderAddDialog(innerHeight-8), m.width, panelHeight)
		case m.deleting:
			panel = PanelCentered(m.renderDeleteDialog(), m.width, panelHeight)
		default:
			if m.width >= infoWidthThreshold {
				// List (left) + info panel (right), side by side.
				listWidth := m.width - infoPanelWidth
				_, innerH := panelInner(listWidth, panelHeight)
				body := m.renderReposTab(innerH)
				listBox := PanelCentered(body, listWidth, panelHeight)
				infoBox := m.renderRepoInfo(m.selectedRepo(), panelHeight)
				panel = lipgloss.JoinHorizontal(lipgloss.Top, listBox, infoBox)
			} else {
				// Narrow terminal: list only, full width.
				body := m.renderReposTab(innerHeight)
				panel = PanelCentered(body, m.width, panelHeight)
			}
		}
	case TabChats:
		switch {
		case m.chatDelete:
			panel = PanelCentered(m.renderChatDelete(), m.width, panelHeight)
		case m.chatRename:
			panel = PanelCentered(m.renderChatRename(), m.width, panelHeight)
		default:
			// The table's own border is the panel box (fills the whole area).
			panel = m.renderChatsTab(m.width, panelHeight)
		}
	case TabComingSoon:
		panel = PanelCentered(m.renderPlaceholder("Coming Soon"), m.width, panelHeight)
	}

	// Global settings modal overlays whatever tab is active.
	if m.settingsOpen {
		panel = PanelCentered(m.renderSettings(), m.width, panelHeight)
	}

	search := SearchBox(m.query, m.width)

	footer := Help(m.footerHint())
	if m.lastErr != nil {
		footer = Help("session error: "+m.lastErr.Error()) + "\n" + footer
	}

	content := lipgloss.JoinVertical(lipgloss.Left, tabBar, panel, search, footer)

	// Take over the whole terminal window.
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}

// footerHint returns the key hint line appropriate to the current context:
// an open modal's own keys take priority, otherwise the active tab's keys.
func (m model) footerHint() string {
	const common = "tab switch tabs · ctrl+p settings · ctrl+c quit"

	switch {
	case m.settingsOpen:
		return "↑/↓ field · ←/→ change · esc/enter close"
	case m.renaming:
		return "type to edit · ←/→ buttons · enter confirm · esc cancel"
	case m.deleting:
		return "y delete · n/esc cancel · ←/→ switch · enter confirm"
	case m.adding:
		return "↑/↓ move · enter/→ open · ← up · ctrl+s scan here · esc cancel"
	case m.chatting:
		return "↑/↓ move · enter select/open · ← up · ctrl+s start here · esc cancel"
	case m.chatDelete:
		return "y remove · n/esc cancel · ←/→ switch · enter confirm"
	case m.chatRename:
		return "type to edit · ←/→ buttons · enter confirm · esc cancel"
	}

	switch m.activeTab {
	case TabRepos:
		return "↑/↓ navigate · type to search · ctrl+n add · ctrl+r rename · ctrl+d delete · enter open · " + common
	case TabChats:
		return "↑/↓ navigate · type to search · space new chat · ctrl+r rename · ctrl+d remove · enter resume · " + common
	default:
		return common
	}
}

func CreateProgram() *tea.Program {
	return tea.NewProgram(initialModel())
}
