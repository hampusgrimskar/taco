package commands

import "os/exec"

func Example() *exec.Cmd {
	return exec.Command("ls")
}
