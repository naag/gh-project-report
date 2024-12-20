package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/naag/gh-project-report/pkg/github"
	"github.com/naag/gh-project-report/pkg/storage"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var (
	startField   string
	endField     string
	organization string
)

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Capture the current state of a GitHub Project",
	Long: `Capture command fetches the current state of a GitHub Project and saves it locally.
The state includes all metadata such as custom fields, priorities, and dates.`,
	RunE: runCapture,
}

func init() {
	rootCmd.AddCommand(captureCmd)
	captureCmd.Flags().StringVar(&startField, "start-field", "Start", "Field name containing start date")
	captureCmd.Flags().StringVar(&endField, "end-field", "End", "Field name containing end date")
	captureCmd.Flags().StringVarP(&organization, "organization", "o", "", "GitHub organization name (optional)")
}

func runCapture(cmd *cobra.Command, args []string) error {
	// Get GitHub token from environment
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	// Get verbose flag from root command
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return fmt.Errorf("failed to get verbose flag: %w", err)
	}

	// Setup GitHub client
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	if verbose {
		log.Printf("Using GitHub token: %s...\n", token[:10])
	}

	client := github.NewClient(httpClient, verbose)

	// Fetch project state
	state, err := client.FetchProjectState(projectNumber, organization, startField, endField)
	if err != nil {
		return fmt.Errorf("failed to fetch project state: %w", err)
	}

	// Create storage
	store, err := storage.NewStore("")
	if err != nil {
		return fmt.Errorf("failed to create storage: %w", err)
	}

	// Save state
	filename, err := store.SaveState(state)
	if err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	log.Printf("State captured and saved to %s\n", filename)
	return nil
}
