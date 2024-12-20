package storage

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/naag/gh-project-report/pkg/types"
)

// Store represents a storage for project states
type Store struct {
	baseDir string
}

// NewStore creates a new store
func NewStore(baseDir string) (*Store, error) {
	if baseDir == "" {
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Create base directory if it doesn't exist
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Store{
		baseDir: baseDir,
	}, nil
}

// SaveState saves a project state to disk
func (s *Store) SaveState(state *types.ProjectState) (string, error) {
	// Validate state
	err := validateState(state)
	if err != nil {
		return "", fmt.Errorf("invalid state: %w", err)
	}

	// Create states directory if it doesn't exist
	statesDir := filepath.Join(s.baseDir, "states")
	err = os.MkdirAll(statesDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create states directory: %w", err)
	}

	// Create project directory if it doesn't exist
	projectDir := filepath.Join(statesDir, fmt.Sprintf("project=%d", state.ProjectNumber))
	err = os.MkdirAll(projectDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create project directory: %w", err)
	}

	// Create filename with unix timestamp
	filename := filepath.Join(projectDir, fmt.Sprintf("%d.json", state.Timestamp.Unix()))

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write state file: %w", err)
	}

	return filename, nil
}

// LoadState loads a project state from disk
func (s *Store) LoadState(projectNumber int, timestamp time.Time) (*types.ProjectState, error) {
	// Find closest state file
	filename, err := s.FindClosestState(projectNumber, timestamp)
	if err != nil {
		return nil, err
	}

	return s.LoadStateFile(filename)
}

// findClosestState finds the state file closest to the given timestamp
func (s *Store) FindClosestState(projectNumber int, timestamp time.Time) (string, error) {
	// Get list of state files
	projectDir := filepath.Join(s.baseDir, "states", fmt.Sprintf("project=%d", projectNumber))
	files, err := ioutil.ReadDir(projectDir)
	if err != nil {
		return "", fmt.Errorf("failed to read project directory: %w", err)
	}

	// Filter and sort state files
	var stateFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".json") {
			stateFiles = append(stateFiles, filepath.Join(projectDir, file.Name()))
		}
	}

	if len(stateFiles) == 0 {
		return "", fmt.Errorf("no state files found for project %d", projectNumber)
	}

	// Sort files by timestamp
	sort.Slice(stateFiles, func(i, j int) bool {
		return extractTimestamp(stateFiles[i]).Before(extractTimestamp(stateFiles[j]))
	})

	// Find closest file
	var closestFile string
	var minDiff time.Duration
	for _, file := range stateFiles {
		diff := timestamp.Sub(extractTimestamp(file))
		if diff < 0 {
			diff = -diff
		}
		if closestFile == "" || diff < minDiff {
			closestFile = file
			minDiff = diff
		}
	}

	return closestFile, nil
}

// LoadStateFile loads a project state from a specific file
func (s *Store) LoadStateFile(filename string) (*types.ProjectState, error) {
	// Read file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal JSON
	var state types.ProjectState
	err = json.Unmarshal(data, &state)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// extractTimestamp extracts the timestamp from a state filename
func extractTimestamp(filename string) time.Time {
	base := filepath.Base(filename)
	if !strings.HasSuffix(base, ".json") {
		return time.Time{}
	}
	timeStr := strings.TrimSuffix(base, ".json")
	unixTime, err := strconv.ParseInt(timeStr, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(unixTime, 0)
}

// validateState validates a project state
func validateState(state *types.ProjectState) error {
	if state.ProjectNumber == 0 {
		return fmt.Errorf("project number is required")
	}

	for i, item := range state.Items {
		// Check required fields
		if item.ID == "" {
			return fmt.Errorf("item %d: ID is required", i)
		}

		if item.GetTitle() == "" {
			return fmt.Errorf("item %d: title is required", i)
		}

		// Check field values
		for field, value := range item.Attributes {
			if value == nil {
				return fmt.Errorf("item %d: field %q has nil value", i, field)
			}
		}
	}

	return nil
}
