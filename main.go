package main

import (
	"fmt"
	"os"
)

func main() {

	if len(os.Args) > 1 {
		firstArgument := os.Args[1]

		switch firstArgument {
		case "run":
			run()
		case "child":
			child()
		default:
			fmt.Printf("unknown command: %s\n", firstArgument)
		}
	} else {
		fmt.Println("Usage: warden <run|child> <command>")
	}
}
