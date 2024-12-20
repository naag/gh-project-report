package cmd

import (
	"fmt"
	"time"

	"github.com/naag/gh-project-report/pkg/format"
	"github.com/naag/gh-project-report/pkg/storage"
	"github.com/naag/gh-project-report/pkg/types"
	"github.com/spf13/cobra"
)

var (
	fromDate     string
	toDate       string
	timeRange    string
	moderateRisk int
	highRisk     int
	extremeRisk  int
	outputFormat string
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Compare project states between two timestamps",
	Long: `Diff command compares two project states and shows what has changed.
It will find the closest state files to the specified dates and compare them.

You can specify the time range in two ways:
1. Using --from and --to flags with ISO8601 timestamps (e.g., 2024-01-01T15:04:05Z)
2. Using --range flag with human-readable format like "last 30 minutes" or "last 2 hours"

The output format can be specified using the --format flag:
- text: Plain text output (default)
- markdown: Markdown table output

Examples:
  gh-project-report diff --from 2024-01-01T15:04:05Z --to 2024-01-02T15:04:05Z
  gh-project-report diff --range "last 30 minutes"
  gh-project-report diff --range "last 2 hours" -p 123
  gh-project-report diff --range "last 1 day"
  gh-project-report diff --range "last 1 week"
  gh-project-report diff --range "last 1 month"
  gh-project-report diff --range "last 1 week" --format markdown`,
	RunE: runDiff,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that either timeRange or both fromDate and toDate are provided
		hasTimeRange := cmd.Flags().Changed("range")
		hasFromTo := cmd.Flags().Changed("from") && cmd.Flags().Changed("to")

		if hasTimeRange == hasFromTo {
			return fmt.Errorf("must specify either --range or both --from and --to flags")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVarP(&fromDate, "from", "f", "", "Start date (ISO8601 format)")
	diffCmd.Flags().StringVarP(&toDate, "to", "t", "", "End date (ISO8601 format)")
	diffCmd.Flags().StringVarP(&timeRange, "range", "r", "", "Human-readable time range (e.g., \"last 30 minutes\", \"last 2 hours\")")
	diffCmd.Flags().IntVar(&moderateRisk, "moderate-risk", 7, "Days of delay to consider moderate risk (default: 7)")
	diffCmd.Flags().IntVar(&highRisk, "high-risk", 14, "Days of delay to consider high risk (default: 14)")
	diffCmd.Flags().IntVar(&extremeRisk, "extreme-risk", 30, "Days of delay to consider extreme risk (default: 30)")
	diffCmd.Flags().StringVarP(&outputFormat, "format", "o", "text", "Output format (text or markdown)")
}

func runDiff(cmd *cobra.Command, args []string) error {
	// Validate output format
	if outputFormat != "text" && outputFormat != "markdown" {
		return fmt.Errorf("invalid output format: %s (must be 'text' or 'markdown')", outputFormat)
	}

	// Create formatter with custom options
	var formatter format.Formatter
	opts := []func(*format.FormatterOptions){
		format.WithModerateDelayThreshold(moderateRisk),
		format.WithHighDelayThreshold(highRisk),
		format.WithExtremeDelayThreshold(extremeRisk),
	}

	if outputFormat == "text" {
		formatter = format.NewTextFormatter(opts...)
	} else {
		formatter = format.NewTableFormatter(opts...)
	}

	// If timeRange is provided, use it
	if cmd.Flags().Changed("range") {
		from, to, err := format.ParseHumanRange(timeRange)
		if err != nil {
			return fmt.Errorf("error parsing time range: %w", err)
		}

		// Create storage
		store, err := storage.NewStore("")
		if err != nil {
			return fmt.Errorf("failed to create storage: %w", err)
		}

		// Load states
		fromState, err := store.LoadState(projectNumber, from)
		if err != nil {
			return fmt.Errorf("failed to load from state: %w", err)
		}

		toState, err := store.LoadState(projectNumber, to)
		if err != nil {
			return fmt.Errorf("failed to load to state: %w", err)
		}

		// Compare states
		diff := types.CompareProjectStates(fromState, toState)
		fmt.Print(formatter.Format(diff))
		return nil
	}

	// Otherwise, parse from and to dates
	fromTime, err := time.Parse(time.RFC3339, fromDate)
	if err != nil {
		return fmt.Errorf("invalid 'from' date format (must be ISO8601): %w", err)
	}

	toTime, err := time.Parse(time.RFC3339, toDate)
	if err != nil {
		return fmt.Errorf("invalid 'to' date format (must be ISO8601): %w", err)
	}

	// Create storage
	store, err := storage.NewStore("")
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Load states
	fromState, err := store.LoadState(projectNumber, fromTime)
	if err != nil {
		return fmt.Errorf("failed to load from state: %w", err)
	}

	toState, err := store.LoadState(projectNumber, toTime)
	if err != nil {
		return fmt.Errorf("failed to load to state: %w", err)
	}

	// Compare states
	diff := types.CompareProjectStates(fromState, toState)
	fmt.Print(formatter.Format(diff))
	return nil
}
