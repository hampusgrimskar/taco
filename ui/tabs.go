package ui

// Tab identifies a top-level view in the UI.
type Tab int

const (
	TabRepos Tab = iota
	TabChats
	TabComingSoon
)

// tabTitles are the labels shown in the tab bar, in order.
var tabTitles = []string{"Repos", "Chats", "Coming Soon"}

// nextTab / prevTab cycle through the tabs.
func (m model) nextTab() Tab {
	return Tab((int(m.activeTab) + 1) % len(tabTitles))
}

func (m model) prevTab() Tab {
	return Tab((int(m.activeTab) - 1 + len(tabTitles)) % len(tabTitles))
}

// renderPlaceholder is shown for tabs with no content yet.
func (m model) renderPlaceholder(name string) string {
	return muted(name + " — nothing here yet.")
}
