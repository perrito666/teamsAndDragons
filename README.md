# Career Progression Tracker

A Go application for tracking software engineer career progression with markdown-based storage and a terminal user interface (TUI).

## Features

- Track multiple engineers' career progressions
- Define milestones (e.g., "Senior 1 → Senior 2")
- Set objectives within each milestone
- Record positive and negative actions contributing to objectives
- Manual review system with tally of pros/cons
- Mark objectives as achieved with optional notes
- Human-readable markdown storage with YAML frontmatter
- HTML export for sharing and reporting
- Clean architecture for easy extension

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd teamsAndDragons

# Install dependencies
go mod tidy

# Build the application
go build -o tracker ./cmd/tui

# Or run directly
go run ./cmd/tui
```

## Usage

### Running the TUI

```bash
# Run with default data directory (./data)
./tracker

# Specify a custom data directory
./tracker -data-dir /path/to/data

# Set log level (debug, info, warn, error)
./tracker -log-level debug
```

### Command-Line Export

```bash
# Export a single person to HTML
./tracker -export "John Doe" -output john-report.html

# Export all people to a directory
./tracker -export-all -output ./exports/
```

### TUI Navigation

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate lists |
| `Enter` | Select/view item |
| `a` | Add new item |
| `e` | Edit selected item |
| `d` | Delete selected item |
| `t` or `Space` | Toggle objective achieved |
| `r` | Open review screen |
| `x` | Export current person to HTML |
| `Esc` or `Backspace` | Go back |
| `q` or `Ctrl+C` | Quit |

### Form Navigation

| Key | Action |
|-----|--------|
| `Tab` | Next field |
| `Shift+Tab` | Previous field |
| `←/→` | Change impact (positive/negative) |
| `Ctrl+S` | Save |
| `Esc` | Cancel |

## Data Storage

Data is stored as human-readable markdown files organized by person:

```
data/
├── john-doe/
│   ├── 00-profile.md
│   ├── 01-milestone-senior-1-to-senior-2.md
│   └── 02-milestone-senior-2-to-staff.md
└── jane-smith/
    ├── 00-profile.md
    └── 01-milestone-mid-to-senior-1.md
```

### Profile Format (`00-profile.md`)

```markdown
---
id: "abc123"
name: "John Doe"
created_at: 2024-01-15T10:00:00Z
updated_at: 2024-03-20T14:30:00Z
---

# John Doe

Senior engineer on the Platform team. Strong in Go and distributed systems.
Areas of growth: communication and mentoring.
```

### Milestone Format (`01-milestone-*.md`)

```markdown
---
id: "xyz789"
number: 1
name: "Senior 1 → Senior 2"
status: "in_progress"
created_at: 2024-01-15T10:00:00Z
updated_at: 2024-03-20T14:30:00Z
---

# Milestone: Senior 1 → Senior 2

Focus on technical leadership and cross-team collaboration.

## Objectives

### Better communication of knowledge
<!-- id: obj-001 -->
<!-- achieved: false -->

Demonstrate ability to share knowledge effectively with the team.

> **Review Notes:** Showing progress but needs more consistency.

#### Actions

- **[+] 2024-02-15:** Led team workshop on Go patterns
  <!-- id: act-001 -->
  Well received by team. Good Q&A engagement.

- **[-] 2024-03-01:** Missed documentation deadline
  <!-- id: act-002 -->
  Was overloaded with other priorities.

- **[+] 2024-03-10:** Created onboarding guide for new hires
  <!-- id: act-003 -->
  Now used as standard onboarding material.

### Technical leadership
<!-- id: obj-002 -->
<!-- achieved: true -->

Take ownership of technical decisions and guide the team.

> **Review Notes:** Consistently demonstrates this. Marked achieved 2024-03-15.

#### Actions

- **[+] 2024-01-20:** Led architecture review for new service
  <!-- id: act-004 -->
  Made key decisions that improved system design.
```

### Parsing Rules

- `## Objectives` marks the objectives section
- `### {name}` starts a new objective
- `<!-- id: xxx -->` stores the unique identifier (hidden when rendered)
- `<!-- achieved: true/false -->` stores achievement status
- `> **Review Notes:**` captures optional review notes
- `#### Actions` starts the actions list
- `- **[+] YYYY-MM-DD:**` = positive action
- `- **[-] YYYY-MM-DD:**` = negative action
- Indented text under an action is the notes

## Code Structure

```
teamsAndDragons/
├── cmd/
│   └── tui/
│       └── main.go              # CLI entry point with flags
├── internal/
│   ├── domain/                  # Core domain types
│   │   ├── types.go             # Person, Milestone, Objective, Action
│   │   ├── errors.go            # Domain-specific errors
│   │   └── types_test.go        # Domain tests
│   ├── storage/                 # Persistence layer
│   │   ├── storage.go           # Storage interfaces
│   │   ├── markdown.go          # Markdown parser/writer
│   │   ├── filesystem.go        # File system operations
│   │   ├── export.go            # HTML export
│   │   └── markdown_test.go     # Storage tests
│   ├── service/                 # Business logic
│   │   ├── service.go           # Service interface
│   │   ├── tracker.go           # Service implementation
│   │   └── tracker_test.go      # Service tests
│   └── tui/                     # Terminal UI
│       ├── app.go               # Bubbletea model and views
│       └── styles.go            # Lipgloss styles
├── pkg/
│   └── tracker/                 # Public API
│       ├── tracker.go           # Public tracker API
│       └── options.go           # Configuration options
├── main.go                      # Root entry point
├── go.mod
└── README.md
```

### Architecture Layers

1. **Domain** (`internal/domain/`) - Pure types with validation, no external dependencies
2. **Storage** (`internal/storage/`) - Markdown file I/O with YAML frontmatter parsing
3. **Service** (`internal/service/`) - Business logic orchestrating storage operations
4. **TUI** (`internal/tui/`) - Terminal user interface using bubbletea/lipgloss
5. **Public API** (`pkg/tracker/`) - Clean interface for external consumers

### Key Design Decisions

- **Markdown-first**: Files are human-readable and editable outside the app
- **HTML comments for metadata**: IDs are invisible when rendered but parseable
- **Clean interfaces**: Each layer has well-defined interfaces for testability
- **Functional options**: Public API uses functional options pattern for configuration

## Using as a Library

The `pkg/tracker` package provides a clean API for external use:

```go
package main

import (
    "fmt"
    "time"

    "teamsAndDragons/pkg/tracker"
)

func main() {
    // Create a tracker instance
    t, err := tracker.New(
        tracker.WithDataDir("./my-data"),
    )
    if err != nil {
        panic(err)
    }

    // Create a person
    person, _ := t.CreatePerson("John Doe", "Senior engineer")

    // Create a milestone
    milestone, _ := t.CreateMilestone(person.ID, "Senior 1 → Senior 2", "Growth goals")

    // Add an objective
    milestone, _ = t.AddObjective(person.ID, milestone.ID, "Technical leadership", "Lead projects")

    // Add actions
    objID := milestone.Objectives[0].ID
    t.AddAction(person.ID, milestone.ID, objID, "Led architecture review",
        tracker.ImpactPositive, time.Now(), "Great outcome")

    // Mark objective as achieved
    t.SetObjectiveAchieved(person.ID, milestone.ID, objID, true, "Consistently demonstrated")

    // Export to HTML
    t.ExportPerson(person.ID, "john-report.html")
}
```

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/domain/...
go test ./internal/storage/...
go test ./internal/service/...
```

## Dependencies

- [bubbletea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [lipgloss](https://github.com/charmbracelet/lipgloss) - Style definitions for TUI
- [bubbles](https://github.com/charmbracelet/bubbles) - TUI components (list, text input, etc.)
- [goldmark](https://github.com/yuin/goldmark) - Markdown parsing for HTML export
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML frontmatter parsing

## License

MIT License
