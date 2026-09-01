package session

import (
	"crypto/rand"
	"encoding/hex"
)

// Session represents a running tmux session attached to a repo.
type Session struct {
	// ID is a unique, tmux-safe identifier for the session.
	ID string
}

// New creates a Session with a freshly generated unique ID.
func New() *Session {
	return &Session{ID: newID()}
}

// newID returns a random 16-character hex string.
func newID() string {
	b := make([]byte, 8)
	// crypto/rand.Read never returns an error on supported platforms,
	// but guard anyway to avoid returning an empty ID silently.
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
