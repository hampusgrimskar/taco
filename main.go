package main

import (
	"fmt"
	"os"

	"github.com/hampusgrimskar/taco/repos"
	"github.com/hampusgrimskar/taco/ui"
)

func initializeRepos() {
	if err := repos.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "error: ", err)
		os.Exit(1)
	}
}

func main() {

	initializeRepos()

	program := ui.CreateProgram()

	_, error := program.Run()

	if error != nil {
		fmt.Printf("Error: %v", error)
		os.Exit(1)
	}
}
