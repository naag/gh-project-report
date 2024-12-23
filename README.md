# GitHub Project Report

A tool to track changes in GitHub Projects (new version) over time. This tool captures the state of project items periodically and allows you to compare states between different timestamps to see what has changed.

## Features

- Captures project item metadata including custom fields, priorities, and dates
- Stores historical state locally in a clean, organized format
- Allows diffing between two timestamps to see changes
- Focuses on metadata changes rather than content changes

## Storage Format

States are stored locally in the following structure:
```
./states/
└── project=<number>/
    ├── 1704067200.json
    ├── 1704153600.json
    └── 1704240000.json
```

- States are stored in the `states` directory within your project
- Each project gets its own directory using hive-style naming (`project=123`)
- Files are named using Unix timestamps for easy sorting and comparison
- Each file contains a complete snapshot of the project state at that time

## Usage

```bash
# View help
gh-project-report --help

# Capture current state (user project)
gh-project-report capture -p 123

# Capture organization project state
gh-project-report capture -p 123 -o myorg

# Capture with custom field names
gh-project-report capture -p 123 --start-field "Timeline Start" --end-field "Timeline End"

# Compare states between two timestamps
gh-project-report diff -p 123 -f "2024-01-01" -t "2024-01-15"

# Compare states using human-readable format
gh-project-report diff -p 123 --range "last week"

# View changes in the last day
gh-project-report diff -p 123 --range "1 day"

# Compare specific dates with times
gh-project-report diff -p 123 -f "2024-01-01T09:00:00" -t "2024-01-02T17:00:00"
```

### Example Output

```
Changes between 2024-01-01 00:00:00 and 2024-01-15 00:00:00

Added Items:
- "New Feature X" (ID: 123)
  Status: Todo
  Priority: High
  Due Date: 2024-02-01

Removed Items:
- "Deprecated Task" (ID: 456)
  Status: Done

Changed Items:
- "Existing Task" (ID: 789)
  Status: Todo → In Progress
  Priority: Medium → High
  Timeline: Extended by 5 days (now ends 2024-01-15)
```

## Requirements

- Go 1.21 or higher
- GitHub Personal Access Token with appropriate permissions
- GitHub Project ID

## Installation

```bash
go install github.com/naag/gh-project-report@latest
```

## Configuration

The tool requires the following environment variables:
- `GITHUB_TOKEN`: Your GitHub Personal Access Token with access to the projects you want to track

The following flags are available for all commands:
- `-p` or `--project`: GitHub Project ID (required)
- `-v` or `--verbose`: Enable verbose output (optional)

### capture command flags
- `-o` or `--organization`: GitHub organization name for org-level projects (optional)
- `--start-field`: Field name containing start date (default: "Start")
- `--end-field`: Field name containing end date (default: "End")

### diff command flags
- `--range`: Compare states using relative time (e.g., "last week", "1 day")

The tool will find the closest state files to the specified dates for comparison.

## Development

### Running Tests

```bash
make test
```

### Project Structure

```
.
├── cmd/                    # Command-line interface
├── pkg/
│   ├── diff/              # Diff generation
│   ├── format/            # Output formatting
│   ├── github/            # GitHub API client
│   ├── storage/           # State storage
│   └── types/             # Core types
└── states/                # State storage (generated)
```