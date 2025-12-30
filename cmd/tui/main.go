// Career Progression Tracker TUI
//
// A terminal user interface for tracking software engineer career progression.
//
// Usage:
//
//	tracker [flags]
//
// Flags:
//
//	-data-dir string   Path to data directory (default "./data")
//	-log-level string  Log level: debug, info, warn, error (default "info")
//	-export string     Export person to HTML by name and exit
//	-export-all        Export all people to HTML and exit
//	-output string     Output path for export (default: ./{name}-export.html)
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"teamsAndDragons/internal/service"
	"teamsAndDragons/internal/storage"
	"teamsAndDragons/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Parse flags
	dataDir := flag.String("data-dir", "./data", "Path to data directory")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	exportPerson := flag.String("export", "", "Export person to HTML by name and exit")
	exportAll := flag.Bool("export-all", false, "Export all people to HTML and exit")
	outputPath := flag.String("output", "", "Output path for export")
	flag.Parse()

	// Setup logger
	logger := setupLogger(*logLevel)

	// Initialize storage
	store, err := storage.NewFileStorage(*dataDir, logger)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing storage: %v\n", err)
		os.Exit(1)
	}

	// Initialize exporter and service
	exporter := storage.NewHTMLExporter(store)
	svc := service.NewTracker(store, exporter, logger)

	// Handle export commands
	if *exportPerson != "" {
		if err := handleExport(svc, *exportPerson, *outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *exportAll {
		if err := handleExportAll(svc, *outputPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Start TUI
	model := tui.NewModel(svc)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Log to file in production to avoid interfering with TUI
	logFile, err := os.OpenFile("tracker.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// Fall back to discard if we can't create log file
		return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	}

	return slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: logLevel}))
}

func handleExport(svc service.TrackerService, personName, outputPath string) error {
	// Find person by name
	people, err := svc.ListPeople()
	if err != nil {
		return err
	}

	var personID string
	for _, p := range people {
		if p.Name == personName || p.Slug() == personName {
			personID = p.ID
			break
		}
	}

	if personID == "" {
		return fmt.Errorf("person '%s' not found", personName)
	}

	if err := svc.ExportPerson(personID, outputPath, true); err != nil {
		return err
	}

	if outputPath == "" {
		outputPath = personName + "-export.html"
	}
	fmt.Printf("Exported to %s\n", outputPath)
	return nil
}

func handleExportAll(svc service.TrackerService, outputDir string) error {
	if err := svc.ExportAll(outputDir, true); err != nil {
		return err
	}

	fmt.Println("Exported all people to HTML")
	return nil
}
