package main

import (
	"log"
	"os"

	"github.com/mcheviron/tmpl/cmd"
	"github.com/spf13/cobra"
)

func main() {
	eLogger := log.New(os.Stdout, "ERROR: ", log.Lshortfile)

	var rootCmd = &cobra.Command{Use: "tmpl"}
	rootCmd.AddCommand(cmd.New(eLogger), cmd.Add(eLogger))
	rootCmd.Execute()
}
