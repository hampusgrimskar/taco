package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hampusgrimskar/taco/chats"
	"github.com/hampusgrimskar/taco/commands"
	"github.com/hampusgrimskar/taco/repos"
	"github.com/hampusgrimskar/taco/settings"
	"github.com/hampusgrimskar/taco/ui"
)

func initializeRepos() {
	if err := repos.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}

// initializeSettings loads persisted settings and applies the saved theme.
func initializeSettings() {
	if err := settings.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
	// Apply the persisted color theme (defaults to "Default" if unset).
	ui.ApplyThemeByName(settings.Get("theme", "Default"))
}

// initializeChats loads the saved chat sessions.
func initializeChats() {
	if err := chats.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}

// cleanupOnce guards terminateAllSessions so it runs at most once, whether
// triggered by the deferred call or the signal handler.
var cleanupOnce sync.Once

// terminateAllSessions kills every tmux session started during this run.
func terminateAllSessions() {
	cleanupOnce.Do(func() {
		for _, repo := range repos.WithSessions() {
			cmd := commands.TerminateSession(repo.Session.ID)
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "failed to terminate session %s: %v\n", repo.Session.ID, err)
			}
			repo.Session = nil
		}
	})
}

func main() {
	initializeRepos()
	initializeSettings()
	initializeChats()

	// Normal-exit and Ctrl+C path: bubbletea catches SIGINT itself and
	// returns from Run(), so a deferred cleanup covers both.
	defer terminateAllSessions()

	// Backstop for signals bubbletea does not handle (e.g. SIGTERM): run
	// cleanup, then exit. Ctrl+C (SIGINT) is normally consumed by bubbletea,
	// but we listen for it too in case the program is not in the UI loop.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		terminateAllSessions()
		os.Exit(1)
	}()

	program := ui.CreateProgram()

	if _, err := program.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		terminateAllSessions()
		os.Exit(1)
	}
}
