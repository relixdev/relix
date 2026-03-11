package main

import (
	"os"

	"github.com/relixdev/relix/relixctl/cmd"
)

func main() {
	root := cmd.New("")
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
