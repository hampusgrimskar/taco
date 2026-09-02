package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// kiroProvider implements Provider for the kiro-cli tool.
type kiroProvider struct {
	home string
}

// compile-time assertion that kiroProvider satisfies Provider (like Java's
// "implements", but checked structurally by the compiler).
var _ Provider = (*kiroProvider)(nil)

// NewKiro returns a Provider backed by kiro-cli.
func NewKiro() Provider {
	home, _ := os.UserHomeDir()
	return &kiroProvider{home: home}
}

func (k *kiroProvider) Name() string { return "kiro-cli" }

func (k *kiroProvider) SessionsDir() string {
	return filepath.Join(k.home, ".kiro", "sessions", "cli")
}

func (k *kiroProvider) agentsDir() string {
	return filepath.Join(k.home, ".kiro", "agents")
}

// buildTmuxCommand wraps an inner command line in a tmux session (rooted at
// dir) with taco's standard bindings, matching how repo sessions are launched.
func buildTmuxCommand(tmuxName, dir, inner string) *exec.Cmd {
	script := fmt.Sprintf(`tmux new-session -A -s %s -c %s \
   "tmux bind-key -n Escape run 'if \
   [ #{pane_current_command} != vi ] && \
   [ #{pane_current_command} != vim ] && \
   [ #{pane_current_command} != nvim ] && \
   [ #{pane_current_command} != k9s ] && \
   [ #{pane_current_command} != git ]; \
   then tmux detach; \
   else tmux send-keys Escape; fi'; \
   tmux set-option -g escape-time 0; \
   tmux set -g mouse on; \
   %s"`, tmuxName, dir, inner)
	return exec.Command("sh", "-c", script)
}

// StartCommand starts a new kiro-cli chat in dir with the given agent.
func (k *kiroProvider) StartCommand(dir, agent string) *exec.Cmd {
	inner := "kiro-cli"
	if agent != "" {
		inner = fmt.Sprintf("kiro-cli --agent %s", shellQuote(agent))
	}
	// tmux session name derived from the directory (sanitised).
	name := "taco-chat-" + sanitize(filepath.Base(dir))
	return buildTmuxCommand(name, dir, inner)
}

// ResumeCommand resumes an existing kiro-cli chat by session id.
func (k *kiroProvider) ResumeCommand(sessionID string) *exec.Cmd {
	dir := k.SessionCWD(sessionID)
	if dir == "" {
		dir = k.home
	}
	inner := fmt.Sprintf("kiro-cli chat --resume-id %s", shellQuote(sessionID))
	name := "taco-chat-" + sanitize(sessionID)
	return buildTmuxCommand(name, dir, inner)
}

// Agents lists available agent names (the *.json file stems in agentsDir),
// skipping non-agent files.
func (k *kiroProvider) Agents() ([]string, error) {
	entries, err := os.ReadDir(k.agentsDir())
	if err != nil {
		return nil, err
	}
	agents := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		if strings.HasSuffix(name, ".json.example") || strings.HasPrefix(name, ".") {
			continue
		}
		agents = append(agents, strings.TrimSuffix(name, ".json"))
	}
	sort.Strings(agents)
	return agents, nil
}

// SessionSnapshot returns the set of existing session IDs (json file stems).
func (k *kiroProvider) SessionSnapshot() (map[string]bool, error) {
	ids := map[string]bool{}
	entries, err := os.ReadDir(k.SessionsDir())
	if err != nil {
		return ids, err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".json") {
			ids[strings.TrimSuffix(e.Name(), ".json")] = true
		}
	}
	return ids, nil
}

// sessionMeta is the subset of a kiro session file taco reads.
type sessionMeta struct {
	SessionID            string `json:"session_id"`
	CWD                  string `json:"cwd"`
	Title                string `json:"title"`
	SessionCreatedReason string `json:"session_created_reason"`
	UpdatedAt            string `json:"updated_at"`
	SessionState         struct {
		AgentName *string `json:"agent_name"`
	} `json:"session_state"`
}

func (k *kiroProvider) readMeta(sessionID string) (sessionMeta, bool) {
	var m sessionMeta
	data, err := os.ReadFile(filepath.Join(k.SessionsDir(), sessionID+".json"))
	if err != nil {
		return m, false
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, false
	}
	return m, true
}

// DetectNewSession finds a session that appeared since the snapshot whose cwd
// matches dir, preferring a non-subagent (top-level) session and the most
// recently updated one.
func (k *kiroProvider) DetectNewSession(before map[string]bool, dir string) (string, bool) {
	entries, err := os.ReadDir(k.SessionsDir())
	if err != nil {
		return "", false
	}

	var best sessionMeta
	found := false
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".json")
		if before[id] {
			continue // existed before launch
		}
		m, ok := k.readMeta(id)
		if !ok || m.CWD != dir {
			continue
		}
		// Prefer a non-subagent session; among those, the latest updated.
		if !found {
			best, found = m, true
			continue
		}
		bestSub := best.SessionCreatedReason == "subagent"
		curSub := m.SessionCreatedReason == "subagent"
		if bestSub && !curSub {
			best = m // prefer top-level over subagent
		} else if bestSub == curSub && m.UpdatedAt > best.UpdatedAt {
			best = m
		}
	}
	if !found {
		return "", false
	}
	return best.SessionID, true
}

func (k *kiroProvider) SessionTitle(sessionID string) string {
	m, ok := k.readMeta(sessionID)
	if !ok {
		return ""
	}
	return m.Title
}

func (k *kiroProvider) SessionCWD(sessionID string) string {
	m, ok := k.readMeta(sessionID)
	if !ok {
		return ""
	}
	return m.CWD
}

func (k *kiroProvider) SessionAgent(sessionID string) string {
	m, ok := k.readMeta(sessionID)
	if !ok || m.SessionState.AgentName == nil {
		return ""
	}
	return *m.SessionState.AgentName
}

// shellQuote single-quotes a string for safe inclusion in a shell command.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// sanitize makes a string safe as a tmux session name (no dots/colons/spaces).
func sanitize(s string) string {
	r := strings.NewReplacer(".", "_", ":", "_", " ", "_", "/", "_")
	return r.Replace(s)
}
