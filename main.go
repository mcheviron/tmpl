package main

import (
	"log/slog"
	"os"

	"github.com/mcheviron/tmpl/cmd"
	"github.com/spf13/cobra"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	var rootCmd = &cobra.Command{Use: "tmpl"}
	rootCmd.AddCommand(cmd.New(logger), cmd.Add(logger))
	rootCmd.Execute()
}
