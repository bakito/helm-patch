package main

import (
	"os"

	"github.com/bakito/helm-patch/cmd"
)

func main() {
	migrateCmd := cmd.NewRootCmd(os.Stdout, os.Args[1:])

	if err := migrateCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
