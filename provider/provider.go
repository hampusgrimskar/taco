// Package provider defines the contract taco needs from a chat CLI (e.g.
// kiro-cli) and concrete implementations of it.
//
// The design mirrors Java's interface + implementation: Provider is the
// contract, and each concrete type (e.g. kiroProvider) satisfies it. Go uses
// implicit (structural) satisfaction — no "implements" keyword — so a type is
// a Provider simply by having the required methods.
package provider

import "os/exec"

// Provider abstracts a chat CLI so taco can start, resume, and inspect chat
// sessions without knowing provider-specific details.
type Provider interface {
	// Name is the human-readable provider name.
	Name() string

	// StartCommand returns a command that starts a NEW chat in the given
	// working directory using the given agent (agent may be empty for the
	// provider default). The command is interactive and intended to be run
	// inside a tmux session.
	StartCommand(dir, agent string) *exec.Cmd

	// ResumeCommand returns a command that resumes the chat identified by
	// sessionID. Interactive; run inside tmux.
	ResumeCommand(sessionID string) *exec.Cmd

	// SessionsDir is the directory where session files are stored.
	SessionsDir() string

	// Agents returns the available agent names.
	Agents() ([]string, error)

	// SessionSnapshot returns the set of currently existing session IDs. Used
	// before starting a chat so the newly created session can be detected
	// afterwards via DetectNewSession.
	SessionSnapshot() (map[string]bool, error)

	// DetectNewSession finds the session created since the given snapshot whose
	// working directory matches dir (preferring a top-level, non-subagent
	// session). Returns the new session ID and true, or "" and false.
	DetectNewSession(before map[string]bool, dir string) (string, bool)

	// SessionTitle returns the live display title for a session ID (may be
	// empty if the provider has not set one yet).
	SessionTitle(sessionID string) string

	// SessionCWD returns the working directory recorded for a session ID.
	SessionCWD(sessionID string) string

	// SessionAgent returns the agent name recorded for a session ID (may be
	// empty if the provider has not recorded one).
	SessionAgent(sessionID string) string
}
