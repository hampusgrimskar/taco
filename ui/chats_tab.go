package ui

import (
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/chats"
	"github.com/hampusgrimskar/taco/repos"
)

// chatPhase is the current step of the new-chat wizard.
type chatPhase int

const (
	chatPickAgent chatPhase = iota
	chatBrowseDir
	chatRunning // launched; waiting for the session to be captured on return
)

// chatWizard holds the state of the new-chat flow.
type chatWizard struct {
	phase chatPhase

	// Agent pick phase.
	agents      []string
	agentCursor int

	// Directory browse phase (reuses the same idea as the add wizard).
	dir     string
	entries []string
	cursor  int

	// Chosen values.
	agent string

	// Snapshot of session ids taken before launch, to detect the new one.
	before map[string]bool

	err string
}

// chatsList state on the Chats tab.
// (cursor/scroll live on the model alongside the repos ones.)

// openChatWizard begins the new-chat flow at the agent picker.
func (m *model) openChatWizard() tea.Cmd {
	agents, err := m.provider.Agents()
	m.chatting = true
	m.chat = chatWizard{phase: chatPickAgent, agents: agents}
	if err != nil {
		m.chat.err = err.Error()
	}
	return nil
}

func (m *model) closeChatWizard() {
	m.chatting = false
	m.chat = chatWizard{}
}

func (a *chatWizard) loadDir(dir string) {
	dirs, err := repos.ListDirs(dir)
	if err != nil {
		a.err = err.Error()
		return
	}
	a.err = ""
	a.dir = dir
	entries := make([]string, 0, len(dirs)+1)
	if filepath.Dir(dir) != dir {
		entries = append(entries, "..")
	}
	entries = append(entries, dirs...)
	a.entries = entries
	a.cursor = 0
}

// updateChatWizard handles a key while the new-chat wizard is open.
func (m *model) updateChatWizard(key string) tea.Cmd {
	switch m.chat.phase {
	case chatPickAgent:
		return m.updateChatAgent(key)
	case chatBrowseDir:
		return m.updateChatBrowse(key)
	}
	return nil
}

func (m *model) updateChatAgent(key string) tea.Cmd {
	c := &m.chat
	switch key {
	case "esc":
		m.closeChatWizard()
	case "up":
		if c.agentCursor > 0 {
			c.agentCursor--
		}
	case "down":
		if c.agentCursor < len(c.agents)-1 {
			c.agentCursor++
		}
	case "enter":
		if len(c.agents) > 0 {
			c.agent = c.agents[c.agentCursor]
		}
		// Move to the directory browser, starting at $HOME.
		home, err := os.UserHomeDir()
		if err != nil {
			home = "/"
		}
		c.phase = chatBrowseDir
		c.loadDir(home)
	}
	return nil
}

func (m *model) updateChatBrowse(key string) tea.Cmd {
	c := &m.chat
	switch key {
	case "esc":
		m.closeChatWizard()
	case "up":
		if c.cursor > 0 {
			c.cursor--
		}
	case "down":
		if c.cursor < len(c.entries)-1 {
			c.cursor++
		}
	case "enter", "right":
		if len(c.entries) == 0 {
			return nil
		}
		name := c.entries[c.cursor]
		if name == ".." {
			c.loadDir(filepath.Dir(c.dir))
		} else {
			c.loadDir(filepath.Join(c.dir, name))
		}
	case "left":
		c.loadDir(filepath.Dir(c.dir))
	case "ctrl+s":
		// Start the chat in the current directory.
		return m.startChat(c.dir)
	}
	return nil
}

// startChat snapshots existing sessions, then launches the provider's start
// command in the chosen directory via tea.ExecProcess.
func (m *model) startChat(dir string) tea.Cmd {
	before, _ := m.provider.SessionSnapshot()
	m.chat.before = before
	m.chat.phase = chatRunning

	cmd := m.provider.StartCommand(dir, m.chat.agent)
	return tea.ExecProcess(cmd, func(err error) tea.Msg {
		return chatFinishedMsg{dir: dir, err: err}
	})
}

// onChatFinished captures the newly created session (if any) after the chat
// tmux session is detached.
func (m *model) onChatFinished(msg chatFinishedMsg) {
	before := m.chat.before
	m.closeChatWizard()
	if msg.err != nil {
		m.lastErr = msg.err
		return
	}
	id, ok := m.provider.DetectNewSession(before, msg.dir)
	if !ok {
		return // no new session detected; nothing to save
	}
	// Fallback name: live title if present, else the directory base name.
	name := m.provider.SessionTitle(id)
	if name == "" {
		name = filepath.Base(msg.dir)
	}
	_ = chats.Add(name, id, m.chat.agent)
}

// resumeSelectedChat resumes the chat under the cursor.
func (m *model) resumeSelectedChat() (tea.Model, tea.Cmd) {
	list := m.filteredChats()
	if len(list) == 0 || m.chatCursor < 0 || m.chatCursor >= len(list) {
		return m, nil
	}
	c := list[m.chatCursor]
	cmd := m.provider.ResumeCommand(c.SessionID)
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return chatFinishedMsg{dir: m.provider.SessionCWD(c.SessionID), err: err, resume: true}
	})
}

// chatDisplayName resolves the label for a chat. A user-set (renamed) name
// takes precedence; otherwise the live provider title is preferred, falling
// back to the stored name.
func (m model) chatDisplayName(c *chats.Chat) string {
	if c.Renamed && c.Name != "" {
		return c.Name
	}
	if title := m.provider.SessionTitle(c.SessionID); title != "" {
		return title
	}
	if c.Name != "" {
		return c.Name
	}
	return "untitled"
}

// chatAgent resolves the agent for a chat: live provider value preferred,
// falling back to the persisted agent.
func (m model) chatAgent(c *chats.Chat) string {
	if a := m.provider.SessionAgent(c.SessionID); a != "" {
		return a
	}
	if c.Agent != "" {
		return c.Agent
	}
	return "—"
}

// chatDir resolves the working directory for a chat from the provider.
func (m model) chatDir(c *chats.Chat) string {
	if d := m.provider.SessionCWD(c.SessionID); d != "" {
		return d
	}
	return "—"
}

// --- chat delete confirmation ---

// chatDeleteData holds the confirmation state for removing a chat.
type chatDeleteData struct {
	sessionID string
	name      string
	confirm   dialogButton
	err       string
}

// openChatDelete opens the delete confirmation for the selected chat.
func (m *model) openChatDelete() {
	list := m.filteredChats()
	if len(list) == 0 || m.chatCursor < 0 || m.chatCursor >= len(list) {
		return
	}
	c := list[m.chatCursor]
	m.chatDelete = true
	m.chatDelData = chatDeleteData{
		sessionID: c.SessionID,
		name:      m.chatDisplayName(c),
		confirm:   buttonCancel, // safe default
	}
}

func (m *model) closeChatDelete() {
	m.chatDelete = false
	m.chatDelData = chatDeleteData{}
}

// updateChatDelete handles a key while the chat delete confirmation is open.
func (m *model) updateChatDelete(key string) {
	switch key {
	case "esc", "n":
		m.closeChatDelete()
	case "left", "right", "tab", "shift+tab":
		if m.chatDelData.confirm == buttonSave {
			m.chatDelData.confirm = buttonCancel
		} else {
			m.chatDelData.confirm = buttonSave
		}
	case "y":
		m.commitChatDelete()
	case "enter":
		if m.chatDelData.confirm == buttonCancel {
			m.closeChatDelete()
			return
		}
		m.commitChatDelete()
	}
}

// commitChatDelete removes the chat from taco's tracking (the underlying kiro
// session file is left untouched) and clamps the cursor.
func (m *model) commitChatDelete() {
	if err := chats.Delete(m.chatDelData.sessionID); err != nil {
		m.chatDelData.err = err.Error()
		return
	}
	m.closeChatDelete()

	count := len(m.filteredChats())
	if m.chatCursor >= count {
		m.chatCursor = count - 1
	}
	if m.chatCursor < 0 {
		m.chatCursor = 0
	}
}

// renderChatDelete draws the chat delete confirmation modal.
func (m model) renderChatDelete() string {
	title := dialogTitleStyle.Render("Remove chat")
	question := "Remove \"" + m.chatDelData.name + "\" from taco?"
	note := muted("(the kiro session file is kept)")

	cancel := buttonStyle.Render("Cancel")
	del := buttonStyle.Render("Remove")
	if m.chatDelData.confirm == buttonSave {
		del = buttonFocusedStyle.Render("Remove")
	} else {
		cancel = buttonFocusedStyle.Render("Cancel")
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, cancel, del)

	parts := []string{title, "", question, note, ""}
	if m.chatDelData.err != "" {
		parts = append(parts, dialogErrStyle.Render(m.chatDelData.err), "")
	}
	parts = append(parts, buttons)
	body := lipgloss.JoinVertical(lipgloss.Center, parts...)
	return deleteBoxStyle.Render(body)
}

// --- chat rename ---

// chatRenameData holds the state of the chat rename modal.
type chatRenameData struct {
	sessionID string
	input     string
	focus     dialogButton
	err       string
}

// openChatRename opens the rename dialog for the selected chat, seeded with its
// current display name.
func (m *model) openChatRename() {
	list := m.filteredChats()
	if len(list) == 0 || m.chatCursor < 0 || m.chatCursor >= len(list) {
		return
	}
	c := list[m.chatCursor]
	m.chatRename = true
	m.chatRenData = chatRenameData{
		sessionID: c.SessionID,
		input:     m.chatDisplayName(c),
		focus:     buttonSave,
	}
}

func (m *model) closeChatRename() {
	m.chatRename = false
	m.chatRenData = chatRenameData{}
}

// updateChatRename handles a key while the chat rename dialog is open.
func (m *model) updateChatRename(key string) {
	switch key {
	case "esc":
		m.closeChatRename()
	case "left", "right", "tab", "shift+tab":
		if m.chatRenData.focus == buttonSave {
			m.chatRenData.focus = buttonCancel
		} else {
			m.chatRenData.focus = buttonSave
		}
	case "enter":
		if m.chatRenData.focus == buttonCancel {
			m.closeChatRename()
			return
		}
		m.commitChatRename()
	case "backspace":
		if m.chatRenData.input != "" {
			m.chatRenData.input = m.chatRenData.input[:len(m.chatRenData.input)-1]
		}
	case " ", "space":
		m.chatRenData.input += " "
	default:
		if len(key) == 1 && key[0] >= 0x20 && key[0] < 0x7f {
			m.chatRenData.input += key
		}
	}
}

// commitChatRename persists the new name (marking the chat as user-renamed).
func (m *model) commitChatRename() {
	if err := chats.Rename(m.chatRenData.sessionID, m.chatRenData.input); err != nil {
		m.chatRenData.err = err.Error()
		return
	}
	m.closeChatRename()
}

// renderChatRename draws the chat rename modal.
func (m model) renderChatRename() string {
	title := dialogTitleStyle.Render("Rename chat")

	inputWidth := lipgloss.Width(m.chatRenData.input) + 12
	if inputWidth < 24 {
		inputWidth = 24
	}
	input := dialogInputStyle.Width(inputWidth).Render(m.chatRenData.input + "▏")

	save := buttonStyle.Render("Save")
	cancel := buttonStyle.Render("Cancel")
	if m.chatRenData.focus == buttonSave {
		save = buttonFocusedStyle.Render("Save")
	} else {
		cancel = buttonFocusedStyle.Render("Cancel")
	}
	buttons := lipgloss.JoinHorizontal(lipgloss.Top, save, cancel)

	parts := []string{title, "", input, ""}
	if m.chatRenData.err != "" {
		parts = append(parts, dialogErrStyle.Render(m.chatRenData.err), "")
	}
	parts = append(parts, buttons)
	body := lipgloss.JoinVertical(lipgloss.Center, parts...)
	return dialogBoxStyle.Render(body)
}

func (m *model) moveChatCursor(delta int) {
	count := len(m.filteredChats())
	next := m.chatCursor + delta
	if next >= 0 && next < count {
		m.chatCursor = next
	}
}

// filteredChats returns the chats matching the current search query, matched
// case-insensitively across the Name, Agent, and Directory columns. An empty
// query returns all chats.
func (m model) filteredChats() []*chats.Chat {
	all := chats.All()
	q := strings.ToLower(strings.TrimSpace(m.query))
	if q == "" {
		return all
	}
	out := make([]*chats.Chat, 0, len(all))
	for _, c := range all {
		hay := strings.ToLower(m.chatDisplayName(c) + " " + m.chatAgent(c) + " " + m.chatDir(c))
		if strings.Contains(hay, q) {
			out = append(out, c)
		}
	}
	return out
}

// --- rendering ---

// chatsSepStyle styles the (subtle) header text.
var chatsSepStyle = lipgloss.NewStyle().Faint(true)

// chatsTableStyle is the rounded border around the whole table.
var chatsTableStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	Padding(1, 2)

// chatsSelectedRowStyle is the full-width accent bar for the selected row.
var chatsSelectedRowStyle = lipgloss.NewStyle()

func (m model) renderChatsTab(width, height int) string {
	// Inner content rows available after the border (2) + vertical padding (2).
	innerRows := height - 4
	if innerRows < 1 {
		innerRows = 1
	}

	if m.chatting {
		body := m.renderChatWizard(innerRows)
		// Center the wizard within the box's inner area.
		innerW := width - 6 // border (2) + horizontal padding (4)
		if innerW < 1 {
			innerW = 1
		}
		centered := lipgloss.Place(innerW, innerRows, lipgloss.Center, lipgloss.Center, body)
		return chatsTableStyle.BorderForeground(colorBorder).
			Width(width).Height(height).Render(centered)
	}

	list := m.filteredChats()
	if len(list) == 0 {
		msg := "No chats yet. Press space to start one."
		if strings.TrimSpace(m.query) != "" {
			msg = "No matches."
		}
		return chatsTableStyle.BorderForeground(colorBorder).
			Width(width).Height(height).
			Render(muted(msg))
	}

	// Collect raw cell values; truncation happens per final column width.
	names := make([]string, len(list))
	agents := make([]string, len(list))
	dirs := make([]string, len(list))
	for i, c := range list {
		names[i] = m.chatDisplayName(c)
		agents[i] = m.chatAgent(c)
		dirs[i] = m.chatDir(c)
	}

	// Lay the three columns out across the full inner width, each getting an
	// equal share (Directory takes any remainder). Values are left-aligned
	// within their cell, so columns spread evenly across the box.
	// chatsTableStyle: border (2) + horizontal padding (4) = 6.
	innerW := width - 6
	if innerW < 12 {
		innerW = 12
	}
	colW := innerW / 3
	nameW := colW
	agentW := colW
	dirW := innerW - 2*colW // remainder to the last column

	rowWidth := innerW

	// Minimum spacing kept between columns: truncate values to leave room so a
	// long value never butts up against the next column.
	const colGap = 2

	// fit truncates a value to (w - colGap) then pads it to exactly w cells,
	// guaranteeing at least colGap spaces before the next column.
	fit := func(s string, w int) string {
		return pad2(truncateRight(s, w-colGap), w)
	}
	rowText := func(n, a, d string) string {
		return fit(n, nameW) + fit(a, agentW) + fit(d, dirW)
	}

	// Header row + a horizontal rule beneath it (header separator only).
	header := chatsSepStyle.Render(rowText("Name", "Agent", "Directory"))
	rule := chatsSepStyle.Render(strings.Repeat("─", innerW))

	// Rows that fit after the header + rule lines.
	rowsAvail := innerRows - 2
	if rowsAvail < 1 {
		rowsAvail = 1
	}
	start, end := windowBounds(m.chatCursor, len(list), rowsAvail)

	lines := make([]string, 0, end-start+2)
	lines = append(lines, header, rule)
	for i := start; i < end; i++ {
		line := rowText(names[i], agents[i], dirs[i])
		if i == m.chatCursor {
			// Full-width accent bar across the whole row.
			line = chatsSelectedRowStyle.Width(rowWidth).Render(line)
		}
		lines = append(lines, line)
	}

	table := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return chatsTableStyle.BorderForeground(colorBorder).
		Width(width).Height(height).Render(table)
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

// pad2 pads s with trailing spaces to exactly w display cells.
func pad2(s string, w int) string {
	return s + spaces(w-lipgloss.Width(s))
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

func (m model) renderChatWizard(visibleRows int) string {
	switch m.chat.phase {
	case chatPickAgent:
		return m.renderChatAgentPick(visibleRows)
	case chatBrowseDir:
		return m.renderChatBrowse(visibleRows)
	case chatRunning:
		return dialogTitleStyle.Render("Starting chat …")
	}
	return ""
}

func (m model) renderChatAgentPick(visibleRows int) string {
	title := dialogTitleStyle.Render("New chat — choose agent")
	if len(m.chat.agents) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, title, "", muted("No agents found."), "", Help("esc cancel"))
	}
	if visibleRows < 1 {
		visibleRows = 1
	}
	start, end := windowBounds(m.chat.agentCursor, len(m.chat.agents), visibleRows)
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		lines = append(lines, Row(m.chat.agents[i], i == m.chat.agentCursor))
	}
	list := lipgloss.JoinVertical(lipgloss.Left, lines...)
	hint := Help("↑/↓ move · enter select · esc cancel")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", list, "", hint)
}

func (m model) renderChatBrowse(visibleRows int) string {
	title := dialogTitleStyle.Render("New chat — pick directory")
	path := muted(m.chat.dir)
	if visibleRows < 1 {
		visibleRows = 1
	}
	start, end := windowBounds(m.chat.cursor, len(m.chat.entries), visibleRows)
	lines := make([]string, 0, end-start)
	for i := start; i < end; i++ {
		name := m.chat.entries[i]
		label := "> " + name
		if name == ".." {
			label = "^ .."
		}
		lines = append(lines, Row(label, i == m.chat.cursor))
	}
	list := lipgloss.JoinVertical(lipgloss.Left, lines...)
	hint := Help("↑/↓ move · enter/→ open · ← up · ctrl+s start here · esc cancel")
	parts := []string{title, path, "", list, "", hint}
	if m.chat.err != "" {
		parts = append(parts, dialogErrStyle.Render(m.chat.err))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}
