package commands

import (
	"fmt"
	"os/exec"
)

const CREATE_SESSION_COMMAND = `tmux new-session -s %s \
   -c %s \
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
   zsh"
  `
const ATTACH_SESSION_COMMAND = `tmux attach -t %s`
const DETACH_SESSION_COMMAND = `tmux detach -s %s`
const TERMINATE_SESSION_COMMAND = `tmux kill-session -t %s`

func CreateSession(sessionId string, directory string) *exec.Cmd {
	cmd := fmt.Sprintf(CREATE_SESSION_COMMAND, sessionId, directory)
	return exec.Command(cmd)
}

func AttachToSession(sessionId string) *exec.Cmd {
	cmd := fmt.Sprintf(ATTACH_SESSION_COMMAND, sessionId)
	return exec.Command(cmd)
}

func DetachFromSession(sessionId string) *exec.Cmd {
	cmd := fmt.Sprintf(DETACH_SESSION_COMMAND, sessionId)
	return exec.Command(cmd)
}

func TerminateSession(sessionId string) *exec.Cmd {
	cmd := fmt.Sprintf(TERMINATE_SESSION_COMMAND, sessionId)
	return exec.Command(cmd)
}
