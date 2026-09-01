package main

import (
	"fmt"

	"github.com/hampusgrimskar/taco/commands"
)

func main() {
	output, error := commands.Example().Output()
	if error != nil {
		fmt.Println(error.Error())
	}
	fmt.Println(string(output))
}
