package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "gh-project-report",
		Short: "A tool to track changes in GitHub Projects",
		Long: `gh-project-report is a CLI tool that helps track changes in GitHub Projects (new version) over time.
It captures the state of project items periodically and allows you to compare states between different timestamps.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Shared flags
	verbose       bool
	projectNumber int
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().IntVarP(&projectNumber, "project-number", "p", 0, "GitHub Project number (required)")
	rootCmd.MarkPersistentFlagRequired("project-number")

	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose debug output")
}
