// Package chats persists the list of chat sessions taco knows about, so they
// can be resumed later. Each entry pairs a fallback name with a provider
// session id; the live display title is resolved from the provider at render
// time (see the ui package).
package chats

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Chat is a saved chat session.
type Chat struct {
	// Name is a persisted fallback label (used when the provider has no live
	// title yet). When Renamed is true, Name is a user-set name that takes
	// precedence over the provider's live title.
	Name string
	// SessionID identifies the chat in the provider (e.g. kiro session id).
	SessionID string
	// Agent is a persisted fallback agent name (used when the provider has no
	// live agent recorded).
	Agent string
	// Renamed is true when the user explicitly renamed this chat, so Name
	// should override the provider's live title.
	Renamed bool
	// Session is the tmux session handle while the chat is running, or nil.
	Session *string
}

// Instance is the global list of saved chats.
var Instance []*Chat

var filePath string

// Init loads chats from ~/.taco/chats, creating the directory if needed.
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	filePath = filepath.Join(home, ".taco", "chats")

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}

	chats, err := load()
	if err != nil {
		return err
	}
	Instance = chats
	return nil
}

func load() ([]*Chat, error) {
	list := make([]*Chat, 0)

	f, err := os.Open(filePath)
	if os.IsNotExist(err) {
		return list, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: name#sessionID#agent#renamed (agent, renamed optional).
		parts := strings.SplitN(line, "#", 4)
		c := &Chat{Name: parts[0]}
		if len(parts) >= 2 {
			c.SessionID = parts[1]
		} else {
			// No '#': treat the whole line as the session id.
			c.SessionID = parts[0]
		}
		if len(parts) >= 3 {
			c.Agent = parts[2]
		}
		if len(parts) >= 4 {
			c.Renamed = parts[3] == "1"
		}
		list = append(list, c)
	}
	return list, scanner.Err()
}

func save() error {
	var b strings.Builder

	sorted := make([]*Chat, len(Instance))
	copy(sorted, Instance)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	for _, c := range sorted {
		renamed := "0"
		if c.Renamed {
			renamed = "1"
		}
		fmt.Fprintf(&b, "%s#%s#%s#%s\n", c.Name, c.SessionID, c.Agent, renamed)
	}

	tmp := filePath + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, filePath)
}

// All returns the live chat list.
func All() []*Chat {
	return Instance
}

// FindByID returns the chat with the given session id, or nil.
func FindByID(sessionID string) *Chat {
	for _, c := range Instance {
		if c.SessionID == sessionID {
			return c
		}
	}
	return nil
}

// Add records a new chat (name + session id + agent) and persists. If the
// session id is already known it is a no-op.
func Add(name, sessionID, agent string) error {
	if FindByID(sessionID) != nil {
		return nil
	}
	Instance = append(Instance, &Chat{Name: name, SessionID: sessionID, Agent: agent})
	return save()
}

// Delete removes the chat with the given session id and persists.
func Delete(sessionID string) error {
	for i, c := range Instance {
		if c.SessionID == sessionID {
			Instance = append(Instance[:i], Instance[i+1:]...)
			return save()
		}
	}
	return nil
}

// Rename changes a chat's fallback name and persists.
func Rename(sessionID, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return fmt.Errorf("name cannot be empty")
	}
	c := FindByID(sessionID)
	if c == nil {
		return fmt.Errorf("no chat with id %q", sessionID)
	}
	c.Name = newName
	c.Renamed = true
	return save()
}
