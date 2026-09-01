package main

import (
	"fmt"
	"os"

	"github.com/hampusgrimskar/taco/ui"
)

func main() {
	program := ui.CreateProgram()

	_, error := program.Run()

	if error != nil {
		fmt.Printf("Error: %v", error)
		os.Exit(1)
	}
}
