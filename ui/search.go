package ui

// setQuery updates the search query and resets the cursor and scroll to the
// top of the filtered list so the selection stays valid.
func (m *model) setQuery(q string) {
	m.query = q
	m.repoCursor = 0
	m.repoScroll = 0
	m.chatCursor = 0
}
