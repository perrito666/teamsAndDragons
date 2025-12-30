/*
	 Career Progression Tracker TUI

	 A terminal user interface for tracking software engineer career progression.

	 Usage:

		tracker [flags]

	 Flags:

		-data-dir string   Path to data directory (default "./data")
		-log-level string  Log level: debug, info, warn, error (default "info")
		-export string     Export person to HTML by name and exit
		-export-all        Export all people to HTML and exit
		-output string     Output path for export (default: ./{name}-export.html)
*/
package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"teamsAndDragons/internal/service"
	"teamsAndDragons/internal/storage"
	"teamsAndDragons/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := mainFn(); err != nil {
		_, stdErr := fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if stdErr != nil {
			// stdErr implies we were not able to print the error to stderr.
			fmt.Printf("error: %v\n", err)
		}
		os.Exit(1)
	}

}

func mainFn() error {
	// Parse flags
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if xdgDataHome == "" {
		var err error
		xdgDataHome, err = os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("could not determine home directory: %w", err)
		}
		// This is the default for XDG_DATA_HOME according to https://specifications.freedesktop.org/basedir/latest/#variables
		xdgDataHome = filepath.Join(xdgDataHome, ".local", "share")
	}
	dataDir := flag.String("data-dir", filepath.Join(xdgDataHome, "teamsanddragons"), "Path to data directory")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	logFile := flag.String("log-file", "./tracker.log", "Log file")
	exportPerson := flag.String("export", "", "Export person to HTML by name and exit")
	exportAll := flag.Bool("export-all", false, "Export all people to HTML and exit")
	outputPath := flag.String("output", "dat_dm_book", "Output path for export")
	flag.Parse()

	// Setup logger
	logger := setupLogger(*logLevel, *logFile)

	// Initialize storage
	store, err := storage.NewFileStorage(*dataDir, logger)
	if err != nil {
		return fmt.Errorf("could not initialize storage: %w", err)
	}

	// Initialize exporter and service
	exporter := storage.NewHTMLExporter(store)
	svc := service.NewTracker(store, exporter, logger)

	// Handle export commands
	if *exportPerson != "" {
		if err := handleExport(svc, *exportPerson, *outputPath); err != nil {
			return fmt.Errorf("could not export person: %w", err)
		}
		return nil
	}

	if *exportAll {
		if err := handleExportAll(svc, *outputPath); err != nil {
			return fmt.Errorf("could not export all: %w", err)
		}
	}

	// Start TUI
	model := tui.NewModel(svc)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err = p.Run(); err != nil {
		return fmt.Errorf("could not run program: %w", err)
	}
	return nil
}

func setupLogger(level string, sinkPath string) *slog.Logger {
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

	// Log to file to avoid interfering with TUI
	logFile, err := os.OpenFile(sinkPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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

	// outputPath is the folder, unless it is a file
	stat, err := os.Stat(outputPath)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(outputPath, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not create output directory: %w", err)
		}
		// I am aware this can fail, but it is unlikely
		stat, _ = os.Stat(outputPath)
	} else if err != nil {
		return fmt.Errorf("could not stat output directory: %w", err)
	}
	if stat.IsDir() {
		outputPath = personName + "-export.html"
	}

	if err = svc.ExportPerson(personID, outputPath, true); err != nil {
		return fmt.Errorf("could not export person: %w", err)
	}

	if outputPath == "" {
		outputPath = personName + "-export.html"
	}
	fmt.Printf("Exported to %s\n", outputPath)
	return nil
}

func handleExportAll(svc service.TrackerService, outputDir string) error {
	stat, err := os.Stat(outputDir)
	if errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("could not create output directory: %w", err)
		}
		// I am aware this can fail, but it is unlikely
		stat, _ = os.Stat(outputDir)
	} else if err != nil {
		return fmt.Errorf("could not stat output directory: %w", err)
	}
	if !stat.IsDir() {
		return fmt.Errorf("output directory is not a directory")
	}
	if err = svc.ExportAll(outputDir, true); err != nil {
		return err
	}

	fmt.Println("Exported all people to HTML")
	return nil
}
