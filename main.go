package main

import (
	"os"

	"github.com/kazegusuri/channelzcli/cmd"
)

func main() {
	if err := cmd.NewRootCommand(os.Stdin, os.Stdout).Command().Execute(); err != nil {
		os.Exit(1)
	}
}
