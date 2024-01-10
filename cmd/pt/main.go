package main

import (
	"fmt"
	"os"

	"github.com/jesseward/modfile"
)

func exit(err error) {
	if err != nil {
		fmt.Fprint(os.Stderr, "[ERROR] ", err, "\n")
		os.Exit(1)
	}

	os.Exit(0)
}

func main() {

	if len(os.Args) < 2 {
		exit(fmt.Errorf("Usage: %s <.mod file>", os.Args[0]))
	}

	buffer, err := os.ReadFile(os.Args[1])
	if err != nil {
		exit(err)
	}

	pt, err := modfile.Read(buffer)
	if err != nil {
		exit(err)
	}
	fmt.Println(pt.PrintModInfo())
}
