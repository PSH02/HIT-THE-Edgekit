package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "edgekit",
	Short:   "EdgeKit — scaffold production-ready Go backend projects",
	Long:    "EdgeKit CLI generates new Go backend projects following hexagonal architecture with batteries included.",
	Version: version,
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
