package main

import (
	"fmt"
	"os"

	"github.com/jesseward/modfile"
)

func main() {

	buffer, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	pt, err := modfile.Read(buffer)
	if err != nil {
		panic(err)
	}
	fmt.Println(pt.PrintModInfo())
}
