package ui

// setQuery updates the search query and resets the cursor to the top of the
// filtered list so the selection stays valid.
func (m *model) setQuery(q string) {
	m.query = q
	m.repoCursor = 0
}
