package ui

import (
	"os"
	"path/filepath"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/hampusgrimskar/taco/repos"
)

// addPhase is the current step of the add-repos wizard.
type addPhase int

const (
	phaseBrowse addPhase = iota // choosing a directory to scan
	phaseScanning
	phaseSelect // choosing which found repos to add
	phaseDone   // summary after adding
)

// addDialog holds the state of the add-repos wizard modal.
type addDialog struct {
	phase addPhase

	// Browse phase.
	dir     string   // current directory being browsed
	entries []string // subdirectory names of dir (with ".." prepended)
	cursor  int      // selection in entries
	scroll  int      // scroll offset in entries

	// Select phase.
	found    []string     // repo paths found by the scan
	selected map[int]bool // which found repos are toggled for adding
	selCur   int          // selection in found
	selScr   int          // scroll offset in found

	// Result / error.
	added int    // how many repos were added
	err   string // error message to display
}

// openAddDialog starts the add-repos wizard at $HOME.
func (m *model) openAddDialog() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "/"
	}
	m.adding = true
	m.add = addDialog{phase: phaseBrowse}
	m.add.loadDir(home)
}

// closeAddDialog exits the wizard.
func (m *model) closeAddDialog() {
	m.adding = false
	m.add = addDialog{}
}

// loadDir populates the browser entries for the given directory.
func (a *addDialog) loadDir(dir string) {
	dirs, err := repos.ListDirs(dir)
	if err != nil {
		a.err = err.Error()
		return
	}
	a.err = ""
	a.dir = dir
	// Prepend ".." to allow going up (unless at the filesystem root).
	entries := make([]string, 0, len(dirs)+1)
	if filepath.Dir(dir) != dir {
		entries = append(entries, "..")
	}
	entries = append(entries, dirs...)
	a.entries = entries
	a.cursor = 0
	a.scroll = 0
}

// updateAddDialog handles a key while the wizard is open. It may return a Cmd
// (e.g. to start an async scan).
func (m *model) updateAddDialog(key string) tea.Cmd {
	switch m.add.phase {
	case phaseBrowse:
		return m.updateBrowse(key)
	case phaseScanning:
		if key == "esc" {
			m.closeAddDialog()
		}
	case phaseSelect:
		m.updateSelect(key)
	case phaseDone:
		// Any key closes the summary.
		m.closeAddDialog()
	}
	return nil
}

func (m *model) updateBrowse(key string) tea.Cmd {
	a := &m.add
	switch key {
	case "esc":
		m.closeAddDialog()
	case "up":
		if a.cursor > 0 {
			a.cursor--
		}
	case "down":
		if a.cursor < len(a.entries)-1 {
			a.cursor++
		}
	case "enter", "right":
		// Descend into the selected directory (or go up via "..").
		if len(a.entries) == 0 {
			return nil
		}
		name := a.entries[a.cursor]
		if name == ".." {
			a.loadDir(filepath.Dir(a.dir))
		} else {
			a.loadDir(filepath.Join(a.dir, name))
		}
	case "left":
		a.loadDir(filepath.Dir(a.dir))
	case "ctrl+s":
		// Scan the current directory.
		a.phase = phaseScanning
		return scanCmd(a.dir)
	}
	return nil
}

func (m *model) updateSelect(key string) {
	a := &m.add
	switch key {
	case "esc":
		m.closeAddDialog()
	case "up":
		a.selCur = a.prevSelectable(a.selCur)
	case "down":
		a.selCur = a.nextSelectable(a.selCur)
	case " ", "space":
		// Already-added repos are locked and cannot be toggled.
		if !a.locked(a.selCur) {
			a.selected[a.selCur] = !a.selected[a.selCur]
		}
	case "enter":
		m.commitAdd()
	}
}

// locked reports whether the found repo at index i is already registered and
// therefore not interactable.
func (a *addDialog) locked(i int) bool {
	if i < 0 || i >= len(a.found) {
		return false
	}
	return repos.HasPath(a.found[i])
}

// nextSelectable returns the next index at or after cur+1 that is not locked,
// or cur if none.
func (a *addDialog) nextSelectable(cur int) int {
	for i := cur + 1; i < len(a.found); i++ {
		if !a.locked(i) {
			return i
		}
	}
	return cur
}

// prevSelectable returns the previous index at or before cur-1 that is not
// locked, or cur if none.
func (a *addDialog) prevSelectable(cur int) int {
	for i := cur - 1; i >= 0; i-- {
		if !a.locked(i) {
			return i
		}
	}
	return cur
}

// firstSelectable returns the first non-locked index, or 0 if all are locked.
func (a *addDialog) firstSelectable() int {
	for i := range a.found {
		if !a.locked(i) {
			return i
		}
	}
	return 0
}

// commitAdd registers all selected repos, then shows a summary.
func (m *model) commitAdd() {
	a := &m.add
	count := 0
	for i, path := range a.found {
		if !a.selected[i] {
			continue
		}
		alias, err := repos.AddRepo(path)
		if err != nil {
			a.err = err.Error()
			return
		}
		if alias != "" {
			count++
		}
	}
	a.added = count
	a.phase = phaseDone
}

// scanCmd runs a git-repo scan asynchronously.
func scanCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		paths, err := repos.ScanGitRepos(dir)
		return scanDoneMsg{paths: paths, err: err}
	}
}

// onScanDone transitions the wizard from scanning to the select phase.
func (m *model) onScanDone(msg scanDoneMsg) {
	if !m.adding {
		return
	}
	if msg.err != nil {
		m.add.err = msg.err.Error()
		m.add.phase = phaseBrowse
		return
	}
	m.add.found = msg.paths
	m.add.selected = make(map[int]bool)
	// Pre-check repos that are already registered so their state is clear.
	for i, path := range msg.paths {
		if repos.HasPath(path) {
			m.add.selected[i] = true
		}
	}
	m.add.selCur = 0
	m.add.selScr = 0
	m.add.phase = phaseSelect
	// Start on the first selectable (non already-added) entry.
	m.add.selCur = m.add.firstSelectable()
}

// --- rendering ---

var addDialogStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(colorAccent).
	Padding(1, 2)

// renderAddDialog draws the wizard for the current phase.
func (m model) renderAddDialog(maxRows int) string {
	var body string
	switch m.add.phase {
	case phaseBrowse:
		body = m.renderBrowse(maxRows)
	case phaseScanning:
		body = dialogTitleStyle.Render("Scanning " + m.add.dir + " …")
	case phaseSelect:
		body = m.renderSelect(maxRows)
	case phaseDone:
		body = m.renderAddSummary()
	}

	s := addDialogStyle
	// Fix the box width so it does not resize/wrap while scrolling. lipgloss
	// .Width is the TOTAL width (border + padding included), so add those to
	// the widest content line.
	if w := m.addDialogContentWidth(); w > 0 {
		s = s.Width(w + panelBorderW + panelPaddingW)
	}
	return s.Render(body)
}

// addDialogContentWidth returns the inner content width (in display cells) the
// dialog should use for the current phase, based on the longest possible line.
func (m model) addDialogContentWidth() int {
	widest := 0
	consider := func(s string) {
		if w := lipgloss.Width(s); w > widest {
			widest = w
		}
	}

	switch m.add.phase {
	case phaseBrowse:
		consider(dialogTitleStyle.Render("Add repos — browse"))
		consider(muted(m.add.dir))
		consider(Help("↑/↓ move · enter/→ open · ← up · ctrl+s scan here · esc cancel"))
		for _, name := range m.add.entries {
			label := "📁 " + name
			if name == ".." {
				label = "⬆  .."
			}
			consider(Row(label, true))
		}
	case phaseSelect:
		consider(dialogTitleStyle.Render("Add repos — select"))
		consider(Help("↑/↓ move · space toggle · enter add selected · esc cancel"))
		for _, path := range m.add.found {
			// Measure both the selected and unselected rendered forms so the
			// box is wide enough for whichever is showing (they differ by the
			// cursor prefix) and never forces a wrap.
			consider(Row("[x] "+path, true))
			consider(Row("[x] "+path, false))
		}
	default:
		return 0 // let scanning/done size to content
	}

	return widest
}

func (m model) renderBrowse(maxRows int) string {
	title := dialogTitleStyle.Render("Add repos — browse")
	path := muted(m.add.dir)

	if maxRows < 3 {
		maxRows = 3
	}
	start, end := windowBounds(m.add.cursor, len(m.add.entries), maxRows)

	lines := make([]string, 0, maxRows)
	for i := start; i < end; i++ {
		name := m.add.entries[i]
		label := "📁 " + name
		if name == ".." {
			label = "⬆  .."
		}
		lines = append(lines, Row(label, i == m.add.cursor))
	}
	// Pad to a fixed number of rows so the dialog height is constant while
	// scrolling.
	for len(lines) < maxRows {
		lines = append(lines, "")
	}
	list := lipgloss.JoinVertical(lipgloss.Left, lines...)

	hint := Help("↑/↓ move · enter/→ open · ← up · ctrl+s scan here · esc cancel")

	parts := []string{title, path, "", list, "", hint}
	if m.add.err != "" {
		parts = append(parts, dialogErrStyle.Render(m.add.err))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m model) renderSelect(maxRows int) string {
	title := dialogTitleStyle.Render("Add repos — select")

	if len(m.add.found) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			title, "", muted("No git repos found here."), "",
			Help("esc close"))
	}

	if maxRows < 3 {
		maxRows = 3
	}
	start, end := windowBounds(m.add.selCur, len(m.add.found), maxRows)

	lines := make([]string, 0, maxRows)
	for i := start; i < end; i++ {
		check := "[ ]"
		if m.add.selected[i] {
			check = "[x]"
		}
		label := check + " " + m.add.found[i]

		if m.add.locked(i) {
			// Already added: greyed out, not interactable, no cursor.
			lines = append(lines, muted("  "+label))
		} else {
			lines = append(lines, Row(label, i == m.add.selCur))
		}
	}
	// Pad to a fixed number of rows so the dialog height is constant while
	// scrolling.
	for len(lines) < maxRows {
		lines = append(lines, "")
	}
	list := lipgloss.JoinVertical(lipgloss.Left, lines...)

	hint := Help("↑/↓ move · space toggle · enter add selected · esc cancel")

	parts := []string{title, "", list, "", hint}
	if m.add.err != "" {
		parts = append(parts, dialogErrStyle.Render(m.add.err))
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (m model) renderAddSummary() string {
	title := dialogTitleStyle.Render("Repos added")
	msg := muted("Added " + itoa(m.add.added) + " repo(s).")
	return lipgloss.JoinVertical(lipgloss.Left, title, "", msg, "", Help("press any key to close"))
}

// windowBounds returns a [start, end) slice keeping cursor visible within size
// rows, edge-scrolling style.
func windowBounds(cursor, total, size int) (int, int) {
	if size <= 0 || total <= size {
		return 0, total
	}
	start := cursor - size/2
	if start < 0 {
		start = 0
	}
	if start > total-size {
		start = total - size
	}
	return start, start + size
}

// itoa is a tiny int-to-string helper to avoid importing strconv here.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
